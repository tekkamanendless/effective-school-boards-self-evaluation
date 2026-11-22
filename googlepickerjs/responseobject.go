package googlepickerjs

// The response object passed to the callback method.
type ResponseObject struct {
	// A type representing the action taken by the user to dismiss the dialog.
	Action Action `javascript:"action"`
	// An array of DocumentObjects selected by the user.
	Docs []DocumentObject `javascript:"docs"`
	// The parent folders for the selected items.
	Parents []ParentDocumentObject `javascript:"parents"`
	// The view the user selected these items from.
	ViewToken []any `javascript:"viewToken"` // [0] is the view ID, [1] is the view label (or null), [2] is the a map of view options.
}
