package demo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-app-blazar/blazar/blazar"
	"github.com/go-app-blazar/blazar/htmlevent"
	"github.com/go-app-blazar/googleoauthjs"
	"github.com/go-app-blazar/googlepickerjs"
	"github.com/go-app-blazar/tabler-font"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/tekkamanendless/effective-school-boards-self-evaluation/effectiveschoolboards"
	"golang.org/x/oauth2"
	admin "google.golang.org/api/admin/reports/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SurveyPage struct {
	app.Compo

	errorMessage string

	authClient  *googleoauthjs.TokenClient
	userInfo    *googleoauth2.Userinfo
	userToken   *oauth2.Token
	adminClient *admin.Service

	ready     bool // This will be true when the initial setup is all ready.
	loggedIn  bool // This will be true when the user is logged in.
	loggingIn bool // This will be true when the user is logging in.

	model          *effectiveschoolboards.Model
	items          []effectiveschoolboards.Item
	answers        []string
	score          effectiveschoolboards.Score
	fileID         string
	pickedFile     bool // This is true when the user has picked a file.
	pickedFileName string
	sheetData      *SheetData
}

func (c *SurveyPage) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "SurveyPage: OnMount")

	c.ready = false
	c.loggedIn = false
	c.loggingIn = false

	ctx.GetState("answers", &c.answers)
	ctx.GetState("file-id", &c.fileID)

	model, err := effectiveschoolboards.LoadModel()
	if err != nil {
		slog.ErrorContext(ctx.Context, "SurveyPage: OnMount: Could not load model.", "err", err)
	} else {
		c.model = model
		c.items = model.AllItems(c.answers)
		c.addAnswer(ctx)
	}

	ctx.Handle("SurveyPage.login", c.login)

	{
		loginCallback := func(token oauth2.Token) {
			slog.InfoContext(ctx.Context, "SurveyPage: OnMount: loginCallback", "token", token)
			c.userToken = &token
			ctx.SetState("user-token", *c.userToken).Persist()

			// NOTE: We're not going to do anything with the token; the rest of the login workflow is handled when needed.
		}

		errorCallback := func(err error) {
			slog.ErrorContext(ctx.Context, "SurveyPage: OnMount: errorCallback.", "err", err)
		}

		c.authClient = googleoauthjs.NewTokenClient(ctx.Context,
			googleoauthjs.WithClientID("733119922879-fr690517hpl3s0dfsk2ugme2s3uau0ji.apps.googleusercontent.com"),
			googleoauthjs.WithScopes([]string{
				"email",
				"profile",
				"https://www.googleapis.com/auth/drive.file",
			}),
			googleoauthjs.WithCallback(loginCallback),
			googleoauthjs.WithErrorCallback(errorCallback),
		)
	}

	ctx.Async(func() {
		ctx.GetState("user-token", &c.userToken)
		ctx.GetState("user-info", &c.userInfo)

		if c.userToken != nil && c.userToken.AccessToken != "" {
			slog.InfoContext(ctx.Context, "SurveyPage: OnMount: Saved token found.", "token", c.userToken)

			loggedIn := false
			if c.tokenIsExpired(ctx) {
				slog.InfoContext(ctx.Context, "SurveyPage: OnMount: Token is expired, refreshing token.")

				err := c.refreshToken(ctx)
				if err != nil {
					slog.ErrorContext(ctx.Context, "SurveyPage: OnMount: Could not refresh token.", "err", err)
				} else {
					loggedIn = true
				}
			} else {
				slog.InfoContext(ctx.Context, "SurveyPage: OnMount: Token is not expired, logging in with saved token.")

				err := c.setTokenAsPartOfLoginWorkflow(ctx, *c.userToken)
				if err != nil {
					slog.ErrorContext(ctx.Context, "SurveyPage: OnMount: Could not log in with saved token.", "err", err)
				} else {
					loggedIn = true
				}
			}

			if loggedIn {
				// Take action on login.

				if c.fileID != "" {
					c.loadFileInfo(ctx)
				}
			}
		}

		c.ready = true
		slog.InfoContext(ctx.Context, "SurveyPage: OnMount: Ready.")
		ctx.Update()
	})

	app.Window().Get("gapi").Call("load", "picker" /*, onPickerApiLoad*/)
}

func (c *SurveyPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "SurveyPage: OnNav")

	href := app.Window().Get("location").Get("href").String()
	slog.InfoContext(ctx.Context, "SurveyPage: OnNav: href.", "href", href)

	hrefURL, err := url.Parse(href)
	if err != nil {
		slog.ErrorContext(ctx.Context, "SurveyPage: OnNav: error parsing href.", "err", err)
		return
	}

	if value := hrefURL.Query().Get("file-id"); value != "" {
		slog.InfoContext(ctx.Context, "SurveyPage: OnNav: file-id found.", "file-id", value)

		c.fileID = value
		ctx.SetState("file-id", c.fileID).Persist()

		newQuery := hrefURL.Query()
		newQuery.Del("file-id")
		hrefURL.RawQuery = newQuery.Encode()
		newHref := hrefURL.String()
		slog.InfoContext(ctx.Context, "SurveyPage: OnNav: new href.", "newHref", newHref)
		app.Window().Get("location").Set("href", newHref)
	}
}

func (c *SurveyPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "SurveyPage: Render: ready", "ready", c.ready, "loggingIn", c.loggingIn, "loggedIn", c.loggedIn)

	formatItemText := func(text string) []app.UI {
		boldWords := []string{
			"but not",
			"did not",
			"conflict of interest",
			"does not",
			"has not",
			"have not",
			"previous thirty-six month period",
			"previous twelve month period",
			"without approval",
		}

		for _, word := range boldWords {
			text = strings.ReplaceAll(text, word, "<b>"+word+"</b>")
		}

		var output []app.UI
		for len(text) > 0 {
			index := strings.Index(text, "<b>")
			if index == -1 {
				output = append(output, app.Text(text))
				break
			}

			output = append(output, app.Text(text[:index]))
			text = text[index+len("<b>"):]

			index = strings.Index(text, "</b>")
			if index == -1 {
				output = append(output, app.B().Text(text))
				break
			}
			output = append(output, app.B().Text(text[:index]))
			text = text[index+len("</b>"):]
		}

		return output
	}

	return blazar.Page().
		Body(
			blazar.Snackbar().
				Observe("snackbar-message"),
			app.IfSlice(!c.ready, func() []app.UI {
				return []app.UI{
					app.Div().Text("Performing initial setup..."),
					blazar.ProgressBar().
						Indeterminate(true),
				}
			}).ElseSlice(func() []app.UI {
				return []app.UI{
					app.IfSlice(c.loggingIn, func() []app.UI {
						return []app.UI{
							app.Div().Text("Logging in with Google..."),
							blazar.ProgressBar().
								Indeterminate(true),
						}
					}).ElseIf(!c.loggedIn, func() app.UI {
						return app.Div().
							Body(
								blazar.Button().
									Label("Sign in with Google").
									Icon(tabler.IconFilledBrandGoogle).
									Disabled(c.loggingIn).
									OnClick(func(ctx app.Context, e htmlevent.MouseEvent) {
										ctx.NewAction("SurveyPage.login")
									}),
							)
					}).ElseSlice(func() []app.UI {
						return []app.UI{
							app.Div().
								Style("display", "flex").
								Style("flex-direction", "row").
								Style("gap", "1em").
								Body(
									app.Div().
										Style("display", "flex").
										Style("flex-direction", "row").
										Style("gap", "1em").
										Style("align-items", "center").
										Body(
											app.Img().
												Src(c.userInfo.Picture).
												Width(32).
												Height(32).
												Style("border-radius", "50%").
												Attr("referrerpolicy", "no-referrer"),
											app.Div().
												Body(
													app.Div().
														Text(c.userInfo.Name),
													app.Div().
														Style("font-size", "0.8em").
														Text(c.userInfo.Email),
												),
										),
									app.Div().
										Style("flex", "1"),
									app.Div().
										Body(
											blazar.Button().
												Label("Log out").
												Icon(tabler.IconOutlineLogout).
												OnClick(func(ctx app.Context, e htmlevent.MouseEvent) {
													c.logout(ctx)
													ctx.Update()
												}),
										),
								),
							app.If(c.errorMessage != "", func() app.UI {
								return blazar.StatusBar().
									Bad().
									Text(c.errorMessage)
							}),
							app.Div().
								Body(
									app.Div().
										Body(
											app.Text("In order to save your responses, you need to pick a file to save them to."),
											app.Text(" "),
											app.Text("If you provide an empty spreadsheet, then it will be populated with a header row and can be used for subsequent responses."),
											app.Text(" "),
											app.Text("If you provide a spreadsheet without the proper header row, then it will not be used."),
										),
									blazar.Button().
										Icon(tabler.IconOutlineFileSpreadsheet).
										Label("Pick file").
										OnClick(func(ctx app.Context, e htmlevent.MouseEvent) {
											c.pickFile(ctx)
										}),
								),
							app.If(c.pickedFile, func() app.UI {
								return blazar.Card().
									Body(
										app.Div().
											Style("display", "flex").
											Style("flex-direction", "row").
											Style("gap", "0.5em").
											Style("align-items", "center").
											Body(
												blazar.Icon().
													Icon(tabler.IconOutlineFileSpreadsheet),
												app.Text(c.pickedFileName),
												app.Span().
													Style("flex", "1"),
												blazar.Copy().
													Icon(tabler.IconOutlineShare).
													Value(func() string {
														href := app.Window().Get("location").Get("href").String()
														hrefURL, err := url.Parse(href)
														if err != nil {
															return ""
														}
														newQuery := hrefURL.Query()
														newQuery.Set("file-id", c.fileID)
														hrefURL.RawQuery = newQuery.Encode()
														return hrefURL.String()
													}()).
													OnClick(func(ctx app.Context, e htmlevent.MouseEvent) {
														blazar.SendSnackbarMessage(ctx, "snackbar-message", blazar.SnackbarMessage().Closable(true).Text("Link copied to clipboard."))
													}),
											),
										app.FieldSet().
											Body(
												app.Legend().Text("My Responses"),
												blazar.Button().
													Icon(tabler.IconOutlineDownload).
													Label("Load My Latest Responses").
													OnClick(func(ctx app.Context, e htmlevent.MouseEvent) {
														c.loadMyLatestResponses(ctx)
													}),
												blazar.Button().
													Icon(tabler.IconOutlineDeviceFloppy).
													Label("Save My Current Responses").
													Disabled(len(c.answers) == 0).
													OnClick(func(ctx app.Context, e htmlevent.MouseEvent) {
														c.saveMyCurrentResponses(ctx)
													}),
											),
										app.If(c.sheetData != nil, func() app.UI {
											emailAddresses := c.sheetData.GetEmailAddresses()

											return app.FieldSet().
												Body(
													app.Legend().Text("All Responses"),
													blazar.Collapse().
														Open(true).
														Label("Respondants: "+strconv.Itoa(len(c.sheetData.GetEmailAddresses()))).
														Body(
															app.Range(emailAddresses).Slice(func(i int) app.UI {
																emailAddress := emailAddresses[i]
																lastResponse := c.sheetData.GetLatestRecordByEmail(emailAddress)
																if lastResponse == nil {
																	return app.Div().Text(emailAddress)
																}
																return app.Div().
																	Body(
																		app.Text(emailAddress+" (Last response: "+lastResponse.Date+")"),
																		blazar.Button().
																			Icon(tabler.IconOutlineDownload).
																			Label("Load Latest Responses").
																			DataSet("email-address", emailAddress).
																			OnClick(func(ctx app.Context, e htmlevent.MouseEvent) {
																				emailAddress := e.Target.Call("closest", "[data-email-address]").Get("dataset").Get("emailAddress").String()
																				c.loadLatestResponsesForEmail(ctx, emailAddress)
																			}),
																	)
															}),
														),
												)
										}),
										blazar.Button().
											Icon(tabler.IconOutlineRefresh).
											Label("Refresh File Info").
											OnClick(func(ctx app.Context, e htmlevent.MouseEvent) {
												c.loadFileInfo(ctx)
											}),
									)
							}),
						}
					}),
					app.Div().
						Style("display", "flex").
						Style("flex-direction", "row").
						Style("gap", "1em").
						Style("align-items", "center").
						Body(
							app.Div().
								Body(
									app.Div().
										Class("score-total").
										Style("font-size", "1.2em").
										Style("font-weight", "bold").
										Body(
											app.Text("Score: "+strconv.Itoa(c.score.Total)+"%"),
										),
									app.Table().
										Body(
											app.Range(c.model.Categories).Slice(func(i int) app.UI {
												category := c.model.Categories[i]

												return app.Tr().
													Class("score-category").
													Style("font-size", "0.9em").
													Style("font-weight", "bold").
													Body(
														app.Td().
															Text(category.Name+":"),
														app.Td().
															Style("text-align", "right").
															Text(strconv.Itoa(c.score.CategoryScores[category.Name])+" points"),
													)
											}),
										),
								),
							app.Div().
								Style("flex", "1"),
							blazar.Button().
								Icon(tabler.IconOutlineTrash).
								Label("Reset Responses").
								OnClick(func(ctx app.Context, e htmlevent.MouseEvent) {
									c.resetAnswers(ctx)

									ctx.Update()
								}),
						),
					app.Div().
						Body(
							app.Range(c.items).Slice(func(i int) app.UI {
								item := c.items[i]

								return app.Div().
									Class("item").
									DataSet("id", item.ID).
									Body(
										app.Div().
											Class("item-text").
											Body(formatItemText(item.Text)...),
										app.Div().
											Body(
												blazar.Button().
													Outline(!slices.Contains(c.answers, "-"+item.ID)).
													Label("False").
													OnClick(func(ctx app.Context, e htmlevent.MouseEvent) {
														itemID := e.Target.Call("closest", "[data-id]").Get("dataset").Get("id").String()
														slog.InfoContext(ctx.Context, "SurveyPage: Render: OnClick(False): itemID", "itemID", itemID)

														c.addAnswer(ctx, "-"+itemID)

														ctx.Update()
													}),
												blazar.Button().
													Outline(!slices.Contains(c.answers, "+"+item.ID)).
													Label("True").
													OnClick(func(ctx app.Context, e htmlevent.MouseEvent) {
														itemID := e.Target.Call("closest", "[data-id]").Get("dataset").Get("id").String()
														slog.InfoContext(ctx.Context, "SurveyPage: Render: OnClick(True): itemID", "itemID", itemID)

														c.addAnswer(ctx, "+"+itemID)

														ctx.Update()
													}),
											),
									)
							}),
						),
				}
			}),
		)
}

func (c *SurveyPage) login(ctx app.Context, action app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		c.loggingIn = true
	})

	ctx.Async(func() {
		err := func() error {
			slog.InfoContext(ctx.Context, "SurveyPage: login.")

			c.loggingIn = true
			slog.InfoContext(ctx.Context, "SurveyPage: login: Started.", "loggingIn", c.loggingIn, "loggedIn", c.loggedIn)

			defer func() {
				c.loggingIn = false

				slog.InfoContext(ctx.Context, "SurveyPage: login: Complete.", "loggingIn", c.loggingIn, "loggedIn", c.loggedIn)

				ctx.Update()
			}()

			if c.authClient == nil {
				slog.ErrorContext(ctx.Context, "SurveyPage: login: authClient is nil")
				return fmt.Errorf("authClient is nil")
			}

			err := c.authClient.RequestAccessToken()
			if err != nil {
				slog.ErrorContext(ctx.Context, "SurveyPage: login: error requesting access token.", "err", err)
				return fmt.Errorf("failed to refresh token: %w", err)
			}

			err = c.setTokenAsPartOfLoginWorkflow(ctx, c.authClient.Token())
			if err != nil {
				slog.ErrorContext(ctx.Context, "SurveyPage: login: error logging in with token.", "err", err)
				return fmt.Errorf("failed to login with token: %w", err)
			}

			ctx.Defer(func(ctx app.Context) {
				// TODO: DO SOMETHING ON LOGIN.
			})

			return nil
		}()
		if err != nil {
			slog.ErrorContext(ctx.Context, "SurveyPage: login: Render: could not login.", "err", err)
			blazar.SendSnackbarMessage(ctx, "snackbar-message", blazar.SnackbarMessage().Closable(true).Bad().Text("Could not sign in with Google: "+err.Error()))
			return
		}
	})
}

func (c *SurveyPage) setTokenAsPartOfLoginWorkflow(ctx app.Context, token oauth2.Token) error {
	defer func() {
		c.loggedIn = true
		slog.InfoContext(ctx.Context, "SurveyPage: setTokenAsPartOfLoginWorkflow: Complete.", "loggingIn", c.loggingIn, "loggedIn", c.loggedIn)

		ctx.Update()
	}()

	tokenSource := oauth2.StaticTokenSource(&token)
	slog.InfoContext(ctx.Context, "SurveyPage: setTokenAsPartOfLoginWorkflow: tokenSource.", "tokenSource", tokenSource)

	googleOauthClient, err := googleoauth2.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		slog.ErrorContext(ctx.Context, "SurveyPage: setTokenAsPartOfLoginWorkflow: error creating google oauth client.", "err", err)
		return fmt.Errorf("failed to create google oauth client: %w", err)
	}

	slog.InfoContext(ctx.Context, "SurveyPage: setTokenAsPartOfLoginWorkflow: googleOauthClient created.", "googleOauthClient", googleOauthClient)
	userInfo, err := googleOauthClient.Userinfo.Get().Do()
	if err != nil {
		slog.ErrorContext(ctx.Context, "SurveyPage: setTokenAsPartOfLoginWorkflow: could not get user info.", "err", err)

		if retrieveErr, ok := errors.AsType[*googleapi.Error](err); ok {
			slog.ErrorContext(ctx.Context, "SurveyPage: setTokenAsPartOfLoginWorkflow: oauth2 error.", "Code", retrieveErr.Code)
			if retrieveErr.Code == 401 {
				c.logout(ctx)
			}
		}

		return fmt.Errorf("failed to get user info: %w", err)
	}
	slog.InfoContext(ctx.Context, "SurveyPage: setTokenAsPartOfLoginWorkflow: userInfo retrieved.", "userInfo", userInfo)
	c.userInfo = userInfo
	ctx.SetState("user-info", c.userInfo).Persist()

	slog.InfoContext(ctx.Context, "SurveyPage: setTokenAsPartOfLoginWorkflow: Sending snackbar message.", "Email", c.userInfo.Email)
	blazar.SendSnackbarMessage(ctx, "snackbar-message", blazar.SnackbarMessage().Closable(true).Text("You are logged in as "+c.userInfo.Email))

	adminClient, err := admin.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		slog.ErrorContext(ctx.Context, "SurveyPage: setTokenAsPartOfLoginWorkflow: could not create admin client.", "err", err)
		return fmt.Errorf("failed to create admin client: %w", err)
	}
	slog.InfoContext(ctx.Context, "SurveyPage: setTokenAsPartOfLoginWorkflow: adminClient created.", "adminClient", adminClient)
	c.adminClient = adminClient

	return nil
}

// tokenIsExpired returns true if the token is expired.
func (c *SurveyPage) tokenIsExpired(ctx app.Context) bool {
	if c.userToken == nil {
		return true
	}
	if c.userToken.AccessToken != "" {
		if c.userToken.Expiry.Before(time.Now()) {
			return true
		}
	}

	tokenInfo, err := c.authClient.TokenInfo(ctx.Context, c.userToken.AccessToken)
	if err != nil {
		slog.ErrorContext(ctx.Context, "SurveyPage: tokenIsExpired: error", "err", err)
		return true
	}

	if tokenInfo.ExpiresIn < 300*time.Second {
		return true
	}

	return false
}

func (c *SurveyPage) refreshToken(ctx app.Context) error {
	slog.InfoContext(ctx.Context, "SurveyPage: refreshToken")
	if c.authClient == nil {
		slog.ErrorContext(ctx.Context, "SurveyPage: refreshToken: authClient is nil")
		return fmt.Errorf("authClient is nil")
	}
	if c.userInfo == nil {
		slog.ErrorContext(ctx.Context, "SurveyPage: refreshToken: userInfo is nil")
		return fmt.Errorf("userInfo is nil")
	}

	err := c.authClient.RequestAccessToken(
		googleoauthjs.WithOverridablePrompt(googleoauthjs.PromptEmptyString),
		googleoauthjs.WithOverridableLoginHint(c.userInfo.Id),
	)
	if err != nil {
		slog.ErrorContext(ctx.Context, "SurveyPage: refreshToken: error", "err", err)
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	err = c.setTokenAsPartOfLoginWorkflow(ctx, c.authClient.Token())
	if err != nil {
		slog.ErrorContext(ctx.Context, "SurveyPage: refreshToken: error", "err", err)
		return fmt.Errorf("failed to login with token: %w", err)
	}

	return nil
}

func (c *SurveyPage) logout(ctx app.Context) {
	slog.InfoContext(ctx.Context, "SurveyPage: logout")
	c.userInfo = nil
	ctx.SetState("user-info", nil).Persist()

	c.userToken = nil
	ctx.SetState("user-token", nil).Persist()

	c.loggedIn = false
	c.loggingIn = false

	c.pickedFile = false
	c.pickedFileName = ""

	blazar.SendSnackbarMessage(ctx, "snackbar-message", blazar.SnackbarMessage().Closable(true).Text("You have been logged out."))
}

// resetAnswers resets the answers to the default answers.
//
// The answers will be saved in local storage.
func (c *SurveyPage) resetAnswers(ctx app.Context) {
	c.answers = []string{}

	c.addAnswer(ctx)
}

// addAnswer adds an answer to the answers.
//
// The answers will be saved in local storage.
func (c *SurveyPage) addAnswer(ctx app.Context, answers ...string) {
	for _, answer := range answers {
		id := strings.TrimPrefix(answer, "+")
		id = strings.TrimPrefix(id, "-")

		c.answers = slices.DeleteFunc(c.answers, func(a string) bool {
			return a == "+"+id || a == "-"+id
		})

		c.answers = append(c.answers, answer)
	}
	slices.Sort(c.answers)
	c.answers = slices.Compact(c.answers)

	c.items = c.model.AllItems(c.answers)
	c.score = c.model.CalculateScore(c.answers)

	ctx.SetState("answers", c.answers).Persist()
}

// pickFile picks a file from the user's Google Drive.
func (c *SurveyPage) pickFile(ctx app.Context) {
	slog.InfoContext(ctx.Context, "SurveyPage: pickFile")

	pickerBuilder := googlepickerjs.NewPickerBuilder()
	pickerBuilder.SetOAuthToken(c.userToken.AccessToken)
	pickerBuilder.SetDeveloperKey("AIzaSyDCCJeLBokY2TjBVphFgD9YagDGCaxi6NI")
	pickerBuilder.SetAppId("733119922879") // ddm-personal
	if c.fileID != "" {
		docsView := googlepickerjs.NewDocsView(nil)
		docsView.SetFileIds(c.fileID)
		docsView.SetLabel("Preferred File")

		pickerBuilder.AddView(docsView)
	}
	pickerBuilder.AddView(googlepickerjs.ViewIDSpreadsheets)
	pickerBuilder.SetCallback(func(response *googlepickerjs.ResponseObject) {
		slog.InfoContext(ctx.Context, "SurveyPage: pickFile: picker callback", "response", response)

		if len(response.Docs) == 0 {
			return
		}

		doc := response.Docs[0]

		slog.InfoContext(ctx.Context, "SurveyPage: pickFile: picker callback: doc.", "ID", doc.ID, "URL", doc.URL, "Name", doc.Name)
		ctx.Dispatch(func(ctx app.Context) {
			c.fileID = doc.ID
			ctx.SetState("file-id", c.fileID).Persist()

			c.pickedFile = true
			c.pickedFileName = doc.Name

			err := c.loadSheetData(ctx)
			if err != nil {
				slog.ErrorContext(ctx.Context, "SurveyPage: pickFile: error loading sheet data.", "err", err)
				blazar.SendSnackbarMessage(ctx, "snackbar-message", blazar.SnackbarMessage().Closable(true).Bad().Text("Error loading sheet data: "+err.Error()))
				return
			}
		})
	})
	picker := pickerBuilder.Build()
	picker.SetVisible(true)
}

func (c *SurveyPage) saveMyCurrentResponses(ctx app.Context) {
	slog.InfoContext(ctx.Context, "SurveyPage: saveMyCurrentResponses", "fileID", c.fileID)

	err := func() error {
		tokenSource := oauth2.StaticTokenSource(c.userToken)
		sheetsService, err := sheets.NewService(ctx, option.WithTokenSource(tokenSource))
		if err != nil {
			return fmt.Errorf("error creating sheets service: %w", err)
		}

		// Test for a header and create one if there isn't one.
		{
			valueRange, err := sheetsService.Spreadsheets.Values.Get(c.fileID, "1:1").Do()
			if err != nil {
				return fmt.Errorf("error getting value range: %w", err)
			}
			slog.InfoContext(ctx.Context, "SurveyPage: saveMyCurrentResponses: value range.", "valueRange", valueRange)

			slog.InfoContext(ctx.Context, "SurveyPage: saveMyCurrentResponses: value range values.", "values", valueRange.Values)
			if len(valueRange.Values) == 0 {
				slog.InfoContext(ctx.Context, "SurveyPage: saveMyCurrentResponses: value range values is empty.")

				headerRow := []any{"date", "email"}
				for _, item := range c.model.Items {
					headerRow = append(headerRow, item.ID)
				}

				response, err := sheetsService.Spreadsheets.Values.Append(c.fileID, "A1:ZZ1000", &sheets.ValueRange{Range: "A1:ZZ1000", Values: [][]any{headerRow}}).ValueInputOption("RAW").Do()
				if err != nil {
					return fmt.Errorf("error appending initial header: %w", err)
				}
				slog.InfoContext(ctx.Context, "SurveyPage: saveMyCurrentResponses: value range appended.", "response", response)
			}
		}

		// Get the header row.
		valueRange, err := sheetsService.Spreadsheets.Values.Get(c.fileID, "1:1").Do()
		if err != nil {
			return fmt.Errorf("error getting value range: %w", err)
		}
		slog.InfoContext(ctx.Context, "SurveyPage: saveMyCurrentResponses: value range.", "valueRange", valueRange)
		slog.InfoContext(ctx.Context, "SurveyPage: saveMyCurrentResponses: value range values.", "values", valueRange.Values)

		if len(valueRange.Values) == 0 {
			return fmt.Errorf("no header row present")
		}

		header := []string{}
		for _, cell := range valueRange.Values[0] {
			header = append(header, fmt.Sprintf("%v", cell))
		}

		headerNameToIndexMap := map[string]int{}
		for index, name := range header {
			headerNameToIndexMap[name] = index
		}

		newRow := make([]any, len(header))
		if index, ok := headerNameToIndexMap["date"]; ok {
			newRow[index] = time.Now().Format("2006-01-02")
		} else {
			return fmt.Errorf("header not present: date")
		}
		if index, ok := headerNameToIndexMap["email"]; ok {
			newRow[index] = c.userInfo.Email
		} else {
			return fmt.Errorf("header not present: email")
		}
		for _, item := range c.answers {
			value := ""
			var itemID string
			if strings.HasPrefix(item, "+") {
				value = "1"
				itemID = strings.TrimPrefix(item, "+")
			} else if strings.HasPrefix(item, "-") {
				value = "0"
				itemID = strings.TrimPrefix(item, "-")
			}

			if index, ok := headerNameToIndexMap[itemID]; ok {
				newRow[index] = value
			} else {
				return fmt.Errorf("header not present: %s", itemID)
			}
		}

		response, err := sheetsService.Spreadsheets.Values.Append(c.fileID, "A1:ZZ1000", &sheets.ValueRange{Range: "A1:ZZ1000", Values: [][]any{newRow}}).ValueInputOption("RAW").Do()
		if err != nil {
			return fmt.Errorf("error appending value range: %w", err)
		}
		slog.InfoContext(ctx.Context, "SurveyPage: saveMyCurrentResponses: value range appended.", "response", response)

		return nil
	}()
	if err != nil {
		slog.ErrorContext(ctx.Context, "SurveyPage: saveMyCurrentResponses: error.", "err", err)
		blazar.SendSnackbarMessage(ctx, "snackbar-message", blazar.SnackbarMessage().Closable(true).Bad().Text("Error saving file: "+err.Error()))
		return
	}

	blazar.SendSnackbarMessage(ctx, "snackbar-message", blazar.SnackbarMessage().Closable(true).Good().Text("Your responses have been saved."))
}

type SheetData struct {
	Records []SheetDataRecord
}

type SheetDataRecord struct {
	Date    string
	Email   string
	Answers []string
}

func (s *SheetData) GetEmailAddresses() []string {
	emailAddresses := []string{}
	emailAddressMap := map[string]bool{}
	for _, record := range s.Records {
		if !emailAddressMap[record.Email] {
			emailAddresses = append(emailAddresses, record.Email)
			emailAddressMap[record.Email] = true
		}
	}
	slices.Sort(emailAddresses)
	return emailAddresses
}

func (s *SheetData) GetRecordsByEmail(email string) []SheetDataRecord {
	records := []SheetDataRecord{}
	for _, record := range s.Records {
		if record.Email == email {
			records = append(records, record)
		}
	}
	return records
}

func (s *SheetData) GetLatestRecordByEmail(email string) *SheetDataRecord {
	records := s.GetRecordsByEmail(email)
	if len(records) == 0 {
		return nil
	}
	return &records[len(records)-1]
}

func (c *SurveyPage) loadSheetData(ctx app.Context) error {
	sheetData := SheetData{
		Records: []SheetDataRecord{},
	}

	err := func() error {
		tokenSource := oauth2.StaticTokenSource(c.userToken)
		sheetsService, err := sheets.NewService(ctx, option.WithTokenSource(tokenSource))
		if err != nil {
			return fmt.Errorf("error creating sheets service: %w", err)
		}

		// Get the header row.
		valueRange, err := sheetsService.Spreadsheets.Values.Get(c.fileID, "A1:ZZ10000").Do()
		if err != nil {
			return fmt.Errorf("error getting value range: %w", err)
		}
		slog.InfoContext(ctx.Context, "SurveyPage: loadSheetData: value range.", "valueRange", valueRange)
		slog.InfoContext(ctx.Context, "SurveyPage: loadSheetData: value range values.", "values", valueRange.Values)

		if len(valueRange.Values) == 0 {
			return fmt.Errorf("no header row present")
		}

		header := []string{}
		for _, cell := range valueRange.Values[0] {
			header = append(header, fmt.Sprintf("%v", cell))
		}
		rows := [][]string{}
		for _, rawRow := range valueRange.Values[1:] {
			row := []string{}
			for _, cell := range rawRow {
				row = append(row, fmt.Sprintf("%v", cell))
			}
			for len(row) < len(header) {
				row = append(row, "")
			}
			rows = append(rows, row)
		}
		//slog.InfoContext(ctx.Context, "SurveyPage: loadSheetData: Rows.", "rows", rows)

		headerNameToIndexMap := map[string]int{}
		for index, name := range header {
			headerNameToIndexMap[name] = index
		}

		newRow := make([]any, len(header))
		if index, ok := headerNameToIndexMap["date"]; ok {
			newRow[index] = time.Now().Format("2006-01-02")
		} else {
			return fmt.Errorf("header not present: date")
		}
		if index, ok := headerNameToIndexMap["email"]; ok {
			newRow[index] = c.userInfo.Email
		} else {
			return fmt.Errorf("header not present: email")
		}
		for _, item := range c.answers {
			value := ""
			var itemID string
			if strings.HasPrefix(item, "+") {
				value = "1"
				itemID = strings.TrimPrefix(item, "+")
			} else if strings.HasPrefix(item, "-") {
				value = "0"
				itemID = strings.TrimPrefix(item, "-")
			}

			if index, ok := headerNameToIndexMap[itemID]; ok {
				newRow[index] = value
			} else {
				return fmt.Errorf("header not present: %s", itemID)
			}
		}

		validItemIDMap := map[string]bool{}
		for _, item := range c.model.Items {
			validItemIDMap[item.ID] = true
		}

		var answers []string
		for _, row := range rows {
			rowDate := row[headerNameToIndexMap["date"]]
			rowEmail := row[headerNameToIndexMap["email"]]
			answers = []string{}
			for _, name := range header {
				if !validItemIDMap[name] {
					continue
				}

				switch row[headerNameToIndexMap[name]] {
				case "1":
					answers = append(answers, "+"+name)
				case "0":
					answers = append(answers, "-"+name)
				}
			}

			record := SheetDataRecord{
				Date:    rowDate,
				Email:   rowEmail,
				Answers: answers,
			}
			sheetData.Records = append(sheetData.Records, record)
		}

		return nil
	}()
	if err != nil {
		return err
	}

	c.sheetData = &sheetData
	return nil
}

func (c *SurveyPage) loadLatestResponsesForEmail(ctx app.Context, email string) {
	slog.InfoContext(ctx.Context, "SurveyPage: loadLatestResponsesForEmail", "email", email)

	err := c.loadSheetData(ctx)
	if err != nil {
		slog.ErrorContext(ctx.Context, "SurveyPage: loadMyLatestResponses: error.", "err", err)
		blazar.SendSnackbarMessage(ctx, "snackbar-message", blazar.SnackbarMessage().Closable(true).Bad().Text("Error reading previous responses: "+err.Error()))
		return
	}

	previousRecord := c.sheetData.GetLatestRecordByEmail(email)
	if previousRecord == nil {
		blazar.SendSnackbarMessage(ctx, "snackbar-message", blazar.SnackbarMessage().Closable(true).Bad().Text("No responses found for "+email+"."))
		return
	}

	c.answers = []string{}
	c.addAnswer(ctx, previousRecord.Answers...)

	blazar.SendSnackbarMessage(ctx, "snackbar-message", blazar.SnackbarMessage().Closable(true).Good().Text("The previous response from "+previousRecord.Date+" has been loaded from "+email+"."))
}

func (c *SurveyPage) loadMyLatestResponses(ctx app.Context) {
	slog.InfoContext(ctx.Context, "SurveyPage: loadMyLatestResponses", "fileID", c.fileID)

	err := c.loadSheetData(ctx)
	if err != nil {
		slog.ErrorContext(ctx.Context, "SurveyPage: loadMyLatestResponses: error.", "err", err)
		blazar.SendSnackbarMessage(ctx, "snackbar-message", blazar.SnackbarMessage().Closable(true).Bad().Text("Error reading previous responses: "+err.Error()))
		return
	}

	previousRecord := c.sheetData.GetLatestRecordByEmail(c.userInfo.Email)
	if previousRecord == nil {
		blazar.SendSnackbarMessage(ctx, "snackbar-message", blazar.SnackbarMessage().Closable(true).Bad().Text("No previous responses found."))
		return
	}

	c.answers = []string{}
	c.addAnswer(ctx, previousRecord.Answers...)

	blazar.SendSnackbarMessage(ctx, "snackbar-message", blazar.SnackbarMessage().Closable(true).Good().Text("Your previous response from "+previousRecord.Date+" has been loaded."))
}

func (c *SurveyPage) loadFileInfo(ctx app.Context) error {
	tokenSource := oauth2.StaticTokenSource(c.userToken)
	slog.InfoContext(ctx.Context, "SurveyPage: OnMount: tokenSource.", "tokenSource", tokenSource)

	driveService, err := drive.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return fmt.Errorf("error creating drive service: %w", err)
	}

	file, err := driveService.Files.Get(c.fileID).Fields(googleapi.Field("name")).Do()
	if err != nil {
		return fmt.Errorf("error getting file: %w", err)
	}

	c.pickedFileName = file.Name

	slog.InfoContext(ctx.Context, "SurveyPage: OnMount: Loading sheet data for file ID.", "fileID", c.fileID)
	err = c.loadSheetData(ctx)
	if err != nil {
		return fmt.Errorf("error loading sheet data: %w", err)
	}

	c.pickedFile = true

	return nil
}
