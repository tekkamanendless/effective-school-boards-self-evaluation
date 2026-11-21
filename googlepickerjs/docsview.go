package googlepickerjs

import (
	"strings"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

// Use DocsView to select files from Google Drive.
type DocsView struct {
	jsValue app.Value // The underlying JavaScript value.

	id string
}

func NewDocsView(viewID View) *DocsView {
	argumentViewID := ViewIDDocs.GetID()
	if viewID != nil {
		argumentViewID = viewID.GetID()
	}

	return &DocsView{
		jsValue: app.Window().Get("google").Get("picker").Get("DocsView").New(argumentViewID),
		id:      argumentViewID,
	}
}

// JSValue returns the underlying JavaScript value.
func (d *DocsView) JSValue() app.Value {
	return d.jsValue
}

// GetID returns the ID of the DocsView.
func (d *DocsView) GetID() string {
	return d.id
}

// GetLabel returns the label of the DocsView.
func (d *DocsView) GetLabel() string {
	return d.jsValue.Call("getLabel").String()
}

// Shows shared drives and the files they contain. Before enabling, refer to GoogleDrive API documentation for enabling shared drives.
//
// If true, only shared drives are included in the view.
//
// Don't combine this setting with setParent or setFileIds. Calls to this function override previous calls to setParent or setFileIds.
//
// Don't combine this setting with setOwnedByMe. When setEnableDrives(true) and setOwnedByMe(true) are set, there are no results.
//
// Don't combine this setting with setStarred. When setEnableDrives(true) is set, setStarred is ignored.
func (d *DocsView) SetEnableDrives(enable bool) *DocsView {
	d.jsValue.Call("setEnableDrives", enable)
	return d
}

// Sets the file IDs included in the view.
//
// Don't combine this setting with setEnableDrives or setParent. Calls to this function override previous calls to setEnableDrives or setParent.
func (d *DocsView) SetFileIds(fileIds ...string) *DocsView {
	d.jsValue.Call("setFileIds", strings.Join(fileIds, ","))
	return d
}

// Show folders in the view items.
//
// If you don't set this option, folders aren't displayed in the view.
func (d *DocsView) SetIncludeFolders(included bool) *DocsView {
	d.jsValue.Call("setIncludeFolders", included)
	return d
}

// SetLabel sets the label of the DocsView.
func (d *DocsView) SetLabel(label string) *DocsView {
	d.jsValue.Call("setLabel", label)
	return d
}

// Selects which mode the view will use to display the documents.
//
// If using a scope other than https://www.googleapis.com/auth/drive or https://www.googleapis.com/auth/drive.readonly, it is recommended to use DocsViewMode.LIST as the user has not granted access to thumbnails.
func (d *DocsView) SetMode(mode DocsViewMode) *DocsView {
	d.jsValue.Call("setMode", mode)
	return d
}

// Filters the documents based on whether they are owned by the user, or shared with the user.
//
// Don't combine this setting with setEnableDrives. When setEnableDrives(true) and setOwnedByMe(true) are set, there are no results.
//
// If you don't set this option, all documents, including shared documents, are displayed in the view.
func (d *DocsView) SetOwnedByMe(ownedByMe bool) *DocsView {
	d.jsValue.Call("setOwnedByMe", ownedByMe)
	return d
}

// Sets the initial parent folder to display.
//
// Don't combine this setting with setEnableDrives or setFileIds. Calls to this function override previous calls to setEnableDrives or setFileIds.
func (d *DocsView) SetParent(parentId string) *DocsView {
	d.jsValue.Call("setParent", parentId)
	return d
}

// Allows the user to select a folder in Google Drive.
func (d *DocsView) SetSelectFolderEnabled(enabled bool) *DocsView {
	d.jsValue.Call("setSelectFolderEnabled", enabled)
	return d
}

// Filters the documents based on whether they are starred by the user.
//
// If true, only starred documents are displayed in the view. If false, all documents are displayed in the view.
//
// Don't combine this setting with setEnableDrives. When setEnableDrives(true) is set, setStarred is ignored.
func (d *DocsView) SetStarred(starred bool) *DocsView {
	d.jsValue.Call("setStarred", starred)
	return d
}
