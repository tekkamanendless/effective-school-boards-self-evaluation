package googlepickerjs

import "github.com/maxence-charriere/go-app/v11/pkg/app"

type ParentDocumentObject struct {
	jsValue app.Value // The underlying JavaScript value.
}

// JSValue returns the underlying JavaScript value.
func (p *ParentDocumentObject) JSValue() app.Value {
	return p.jsValue
}
