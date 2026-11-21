package googlepickerjs

import "github.com/maxence-charriere/go-app/v11/pkg/app"

type ViewGroup struct {
	jsValue app.Value // The underlying JavaScript value.
}

// JSValue returns the underlying JavaScript value.
func (v *ViewGroup) JSValue() app.Value {
	return v.jsValue
}

// Add a label to this ViewGroup.
func (v *ViewGroup) AddLabel(label string) *ViewGroup {
	v.jsValue.Call("addLabel", label)
	return v
}

// Add a view to the ViewGroup. The view can be represented by a view-derived object, or simply by ViewId.
func (v *ViewGroup) AddView(view View) *ViewGroup {
	if _, ok := view.(ViewID); ok {
		v.jsValue.Call("addView", view.GetID())
	} else {
		v.jsValue.Call("addView", view.JSValue())
	}
	return v
}

// Nest a ViewGroup within the current ViewGroup.
func (v *ViewGroup) AddViewGroup(viewGroup *ViewGroup) *ViewGroup {
	v.jsValue.Call("addViewGroup", viewGroup.JSValue())
	return v
}
