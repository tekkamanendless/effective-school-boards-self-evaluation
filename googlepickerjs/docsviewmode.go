package googlepickerjs

// DocsViewMode is an enumerated type for displaying data within a DocsView. Use these values in calls to DocsView.setMode.
//
// If using a scope other than https://www.googleapis.com/auth/drive or https://www.googleapis.com/auth/drive.readonly, it is recommended to use DocsViewMode.LIST as the user has not granted access to thumbnails.
type DocsViewMode string

const (
	// Display documents in a thumbnail grid.
	DocsViewModeGrid DocsViewMode = "grid"
	// Display documents in a detailed list.
	DocsViewModeList DocsViewMode = "list"
)
