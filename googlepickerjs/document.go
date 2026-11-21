package googlepickerjs

// Document is an enumerated type used to describe the fields of a DocumentObject.
type Document string

const (
	DocumentAddressLines  Document = "addressLines"
	DocumentAudience      Document = "audience"
	DocumentDescription   Document = "description"
	DocumentDuration      Document = "duration"
	DocumentEmbedUrl      Document = "embedUrl"
	DocumentIconUrl       Document = "iconUrl"
	DocumentId            Document = "id"
	DocumentIsNew         Document = "isNew"
	DocumentLastEditedUTC Document = "lastEditedUtc"
	DocumentLatitude      Document = "latitude"
	DocumentLongitude     Document = "longitude"
	DocumentMimeType      Document = "mimeType"
	DocumentName          Document = "name"
	DocumentNumChildren   Document = "numChildren"
	DocumentParentId      Document = "parentId"
	DocumentPhoneNumbers  Document = "phoneNumbers"
	DocumentReadOnly      Document = "readOnly"
	DocumentServiceId     Document = "serviceId"
	DocumentThumbnails    Document = "thumbnails"
	DocumentType          Document = "type"
	DocumentURL           Document = "url"
	DocumentVersion       Document = "version"
)
