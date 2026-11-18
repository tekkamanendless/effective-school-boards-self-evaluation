package demo

import (
	"log/slog"
	"strings"

	"github.com/go-app-blazar/blazar/blazar"
	"github.com/go-app-blazar/tabler-font"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type IndexPage struct {
	app.Compo
}

func (c *IndexPage) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "IndexPage: OnMount")
}

func (c *IndexPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "IndexPage: OnNav")
}

func (c *IndexPage) Render() app.UI {
	return blazar.Page().
		Body(
			app.P().Text("This tool allows you to self-evaluate your school board's effectiveness in implementing the Effective School Boards framework."),
			app.P().Body(
				blazar.Button().
					Style("font-size", "1.2em").
					Style("padding", "0.5em").
					To("/survey").
					Icon(tabler.IconFilledArrowBigRight).
					Label("Get Started"),
			),
			app.Br(),
			app.H2().Text("Additional Information"),
			app.Div().Body(
				app.P().Body(
					app.Text("For additional information, see below."),
				),
				app.Ul().Body(
					app.Li().Body(
						app.A().
							Href("/terms").
							Text("Terms of Service"),
					),
					app.Li().Body(
						app.A().
							Href("/privacy").
							Text("Privacy Policy"),
					),
				),
			),
		)
}

type googleWorkspacePermission struct {
	app.Compo

	IParts       []string
	IRequired    bool
	IDescription string
}

func GoogleWorkspacePermission() *googleWorkspacePermission {
	return &googleWorkspacePermission{}
}

func (p *googleWorkspacePermission) Parts(parts ...string) *googleWorkspacePermission {
	p.IParts = parts
	return p
}

func (p *googleWorkspacePermission) Required(required bool) *googleWorkspacePermission {
	p.IRequired = required
	return p
}

func (p *googleWorkspacePermission) Description(description string) *googleWorkspacePermission {
	p.IDescription = description
	return p
}

func (p *googleWorkspacePermission) Render() app.UI {
	var pathParts []string
	var name string
	if len(p.IParts) > 0 {
		pathParts = p.IParts[:len(p.IParts)-1]
		name = p.IParts[len(p.IParts)-1]
	}

	return app.Div().
		Style("display", "flex").
		Style("flex-direction", "column").
		Style("gap", "0.2em").
		Style("margin-bottom", "0.5em").
		Body(
			app.Div().
				Style("display", "flex").
				Style("flex-direction", "column").
				Style("border", "1px solid #ccc").
				Style("padding", "0.5em").
				Style("border-radius", "0.5em").
				Body(
					app.Div().
						Class("google-workspace-permission__path").
						Style("font-size", "0.7em").
						Text(strings.Join(pathParts, " > ")),
					app.Div().
						Class("google-workspace-permission__name").
						Style("font-weight", "bold").
						Text(name),
				),
			app.Div().
				Class("google-workspace-permission__description").
				Style("font-style", "italic").
				Text(p.IDescription),
		)
}
