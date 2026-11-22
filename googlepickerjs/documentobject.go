package googlepickerjs

// DocumentObject is an interface describing the attributes of a selected item.
type DocumentObject struct {
	// Deprecated
	AddressLines []string `javascript:"addressLines"`
	// Deprecated
	Audience map[string]string `javascript:"audience"`
	// A user-contributed description of the selected item.
	Description string `javascript:"description"`
	// A URL to download this item.
	DownloadURL string `javascript:"downloadUrl"`
	// The error code for the request to pre-open the document in Google Drive.
	DriveError string `javascript:"driveError"`
	// Whether the request to pre-open the document in Google Drive was successful.
	DriveSuccess bool `javascript:"driveSuccess"`
	// The duration of a selected video.
	Duration int `javascript:"duration"`
	// A URL for this item suitable for embedding in a web page.
	EmbedUrl string `javascript:"embedUrl"`
	// A URL to an icon for this item.
	IconUrl string `javascript:"iconUrl"`
	// The ID for the selected item.
	ID string `javascript:"id"`
	// Returns true if the selected item was just uploaded.
	IsNew bool `javascript:"isNew"`
	// Whether the item is shared.
	IsShared bool `javascript:"isShared"`
	// The timestamp describing when this item was last edited.
	LastEditedUTCNanoseconds uint64 `javascript:"lastEditedUtc"`
	// The latitude of the selected item.
	Latitude float64 `javascript:"latitude"`
	// The longitude of the selected item.
	Longitude float64 `javascript:"longitude"`
	// The MIME type of this item.
	MimeType string `javascript:"mimeType"`
	// The name of this item.
	Name string `javascript:"name"`
	// Display name for the owning organization.
	OrganizationDisplayName string `javascript:"organizationDisplayName"`
	// The parent ID of this item. For example, the folder containing this file.
	ParentID string `javascript:"parentId"`
	// The phone numbers of the selected item.
	PhoneNumbers []struct {
		Type   string `javascript:"type"`
		Number string `javascript:"number"`
	} `javascript:"phoneNumbers"`
	// The resource key for the item, if present. Empty string otherwise
	ResourceKey string `javascript:"resourceKey"`
	// Deprecated
	Rotation float64 `javascript:"rotation"`
	// Deprecated
	RotationDegree float64 `javascript:"rotationDegree"`
	// An id describing the service this item was selected from.
	ServiceID string `javascript:"serviceId"`
	// Size of the picked item in bytes. The value is not returned when the item is uploaded during the Picker session.
	SizeBytes int64 `javascript:"sizeBytes"`
	// An array of Thumbnails which describe the attributes of a photo or video. Thumbnails aren't returned if the selected items belong to Google Drive.
	Thumbnails []ThumbnailObject `javascript:"thumbnails"`
	// The type of the selected item.
	Type string `javascript:"type"`
	// The item order in the upload session.
	UploadID string `javascript:"uploadId"`
	// The state of the upload.
	UploadState string `javascript:"uploadState"`
	// A URL to this item.
	URL string `javascript:"url"`
}
