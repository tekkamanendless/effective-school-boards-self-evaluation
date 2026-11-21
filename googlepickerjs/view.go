package googlepickerjs

import "github.com/maxence-charriere/go-app/v11/pkg/app"

type ViewID string

func (v ViewID) GetID() string {
	return string(v)
}

func (v ViewID) JSValue() app.Value {
	return app.Window().Get("String").New(string(v))
}

var (
	ViewIDDocs             = ViewID("all")
	ViewIDDocsImages       = ViewID("docs-images")
	ViewIDDocsImagesVideos = ViewID("docs-images-and-videos")
	ViewIDDocsVideos       = ViewID("docs-videos")
	ViewIDDocuments        = ViewID("documents")
	ViewIDDrawings         = ViewID("drawings")
	ViewIDFolders          = ViewID("folders")
	ViewIDForms            = ViewID("forms")
	ViewIDPDFs             = ViewID("pdfs")
	ViewIDPresentations    = ViewID("presentations")
	ViewIDSpreadsheets     = ViewID("spreadsheets")
)

// viewStruct is an abstract class for all views.
type viewStruct struct {
	jsValue app.Value // The underlying JavaScript value.
	id      string    // If this represents an enumerated view, then this is the ID of the view.
}

// JSValue returns the underlying JavaScript value.
func (v *viewStruct) JSValue() app.Value {
	return v.jsValue
}

// GetID returns the ID of the view.
func (v *viewStruct) GetID() string {
	return v.id
}

type View interface {
	JSValue() app.Value

	GetID() string
}
