package googlepickerjs

// Feature is an enumerated type, for turning on/off features for various views. Use these values in calls to PickerBuilder.EnableFeature and PickerBuilder.DisableFeature.
type Feature string

const (
	// Show only documents owned by the user when showing items from Google Drive.
	FeatureMineOnly Feature = "mineOnly"
	// Allow user to choose more than one item.
	FeatureMultiselectEnabled Feature = "multiselectEnabled"
	// Hide the navigation pane. If the navigation pane is hidden, users can only select from the first view chosen.
	FeatureNavHidden Feature = "navHidden"
	// For photo uploads, controls whether per-photo DocumentObject (as opposed to per-album) DocumentObject is enabled.
	FeatureSimpleUploadEnabled Feature = "simpleUploadEnabled"
	// Whether shared drive items are included in results.
	//
	// Deprecated.
	// WARNING: Shared drive items are now included by default.
	FeatureSupportDrives Feature = "sdr"
)
