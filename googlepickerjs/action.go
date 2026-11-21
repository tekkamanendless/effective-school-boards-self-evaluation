package googlepickerjs

// The action type for the ResponseObject.
type Action string

const (
	// User canceled the Google Picker dialog.
	ActionCancel Action = "cancel"
	// User has chosen at least one item.
	ActionPicked Action = "picked"
	// The Google Picker dialog has encountered an error.
	ActionError Action = "error"
)
