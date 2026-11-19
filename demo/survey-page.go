package demo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-app-blazar/blazar/blazar"
	"github.com/go-app-blazar/blazar/htmlevent"
	"github.com/go-app-blazar/googleoauthjs"
	"github.com/go-app-blazar/tabler-font"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/tekkamanendless/effective-school-boards-self-evaluation/effectiveschoolboards"
	"golang.org/x/oauth2"
	admin "google.golang.org/api/admin/reports/v1"
	"google.golang.org/api/googleapi"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
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

	model   *effectiveschoolboards.Model
	items   []effectiveschoolboards.Item
	answers []string
	score   int
}

func (c *SurveyPage) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "SurveyPage: OnMount")

	c.ready = false
	c.loggedIn = false
	c.loggingIn = false

	ctx.GetState("answers", &c.answers)

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

			if c.tokenIsExpired(ctx) {
				slog.InfoContext(ctx.Context, "SurveyPage: OnMount: Token is expired, refreshing token.")

				err := c.refreshToken(ctx)
				if err != nil {
					slog.ErrorContext(ctx.Context, "SurveyPage: OnMount: Could not refresh token.", "err", err)
				} else {
					// TODO: DO SOMETHING ON LOGIN.
				}
			} else {
				slog.InfoContext(ctx.Context, "SurveyPage: OnMount: Token is not expired, logging in with saved token.")

				err := c.setTokenAsPartOfLoginWorkflow(ctx, *c.userToken)
				if err != nil {
					slog.ErrorContext(ctx.Context, "SurveyPage: OnMount: Could not log in with saved token.", "err", err)
				} else {
					// TODO: DO SOMETHING ON LOGIN.
				}
			}
		}

		c.ready = true
		slog.InfoContext(ctx.Context, "SurveyPage: OnMount: Ready.")
		ctx.Update()
	})

	model, err := effectiveschoolboards.LoadModel()
	if err != nil {
		slog.ErrorContext(ctx.Context, "SurveyPage: OnMount: Could not load model.", "err", err)
	} else {
		c.model = model
		c.items = model.AllItems(c.answers)
		c.addAnswer(ctx)
	}
}

func (c *SurveyPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "SurveyPage: OnNav")
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
										Text("Score: " + strconv.Itoa(c.score)),
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

	blazar.SendSnackbarMessage(ctx, "snackbar-message", blazar.SnackbarMessage().Closable(true).Text("You have been logged out."))
}

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
