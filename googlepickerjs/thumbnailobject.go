package googlepickerjs

import "github.com/maxence-charriere/go-app/v11/pkg/app"

type ThumbnailObject struct {
	jsValue app.Value // The underlying JavaScript value.

	Height int    `javascript:"height"`
	Type   string `javascript:"type"`
	Width  int    `javascript:"width"`
	URL    string `javascript:"url"`
}

// JSValue returns the underlying JavaScript value.
func (t *ThumbnailObject) JSValue() app.Value {
	return t.jsValue
}
