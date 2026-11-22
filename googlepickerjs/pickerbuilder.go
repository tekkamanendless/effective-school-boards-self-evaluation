package googlepickerjs

import (
	"context"
	"log/slog"
	"strings"

	"github.com/go-app-blazar/appvalue"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

// PickerBuilder is a wrapper around the Google PickerBuilder API.
type PickerBuilder struct {
	jsValue app.Value // The underlying JavaScript value.
}

// NewPickerBuilder creates a new PickerBuilder.
func NewPickerBuilder() *PickerBuilder {
	jsValue := app.Window().Get("google").Get("picker").Get("PickerBuilder").New()

	p := &PickerBuilder{
		jsValue: jsValue,
	}
	return p
}

// JSValue returns the underlying JavaScript value.
func (p *PickerBuilder) JSValue() app.Value {
	return p.jsValue
}

// Add a view to the navigation pane.
func (p *PickerBuilder) AddView(view View) *PickerBuilder {
	if _, ok := view.(ViewID); ok {
		p.jsValue.Call("addView", view.GetID())
	} else {
		p.jsValue.Call("addView", view.JSValue())
	}
	return p
}

// Add a ViewGroup to the top-level navigation pane.
func (p *PickerBuilder) AddViewGroup(viewGroup *ViewGroup) *PickerBuilder {
	p.jsValue.Call("addViewGroup", viewGroup.JSValue())
	return p
}

// Construct the Picker object.
func (p *PickerBuilder) Build() *Picker {
	jsPicker := p.jsValue.Call("build")
	picker := &Picker{
		jsValue: jsPicker,
	}
	return picker
}

// Disable a picker feature.
func (p *PickerBuilder) DisableFeature(feature Feature) *PickerBuilder {
	p.jsValue.Call("disableFeature", feature)
	return p
}

// Enable a picker feature.
func (p *PickerBuilder) EnableFeature(feature Feature) *PickerBuilder {
	p.jsValue.Call("enableFeature", feature)
	return p
}

// Get the relay URL, used for gadgets.rpc.
func (p *PickerBuilder) GetRelayUrl() string {
	return p.jsValue.Call("getRelayUrl").String()
}

// Get the dialog title.
func (p *PickerBuilder) GetTitle() string {
	return p.jsValue.Call("getTitle").String()
}

// Disable the title bar from being shown. To re-enable, call setTitle with a non-empty title or undefined.
func (p *PickerBuilder) HideTitleBar() *PickerBuilder {
	p.jsValue.Call("hideTitleBar")
	return p
}

// Check if a picker Feature is enabled.
func (p *PickerBuilder) IsFeatureEnabled(feature Feature) bool {
	return p.jsValue.Call("isFeatureEnabled", feature).Bool()
}

// Sets the ID of the application needing to access the user's files via the Drive API.
func (p *PickerBuilder) SetAppId(appID string) *PickerBuilder {
	p.jsValue.Call("setAppId", appID)
	return p
}

// Set the callback method. This method is called when the user selects items or cancels. The callback method receives a single callback object. The structure of the callback object is described in the JSON Guide.
func (p *PickerBuilder) SetCallback(callback func(response *ResponseObject)) *PickerBuilder {
	p.jsValue.Call("setCallback", app.FuncOf(func(this app.Value, args []app.Value) any {
		jsRawValue := args[0]

		var response ResponseObject
		//app.Window().Get("console").Call("log", "ResponseObject", response.jsValue)
		err := appvalue.Unmarshal(jsRawValue, &response)
		if err != nil {
			slog.ErrorContext(context.TODO(), "failed to parse response object.", "err", err)
			return nil
		}

		callback(&response)
		return nil
	}))
	return p
}

// Sets the Browser API key obtained from Google Developers Console. See the Developer's Guide for details on how to obtain the Browser API key.
func (p *PickerBuilder) SetDeveloperKey(key string) *PickerBuilder {
	p.jsValue.Call("setDeveloperKey", key)
	return p
}

// Set the document.
func (p *PickerBuilder) SetDocument(document Document) *PickerBuilder {
	p.jsValue.Call("setDocument", document)
	return p
}

// Set the locale for the picker. The locale is an ISO 639 language code. If the language is not supported, en-US is used.
func (p *PickerBuilder) SetLocale(locale string) *PickerBuilder {
	p.jsValue.Call("setLocale", locale)
	return p
}

// Sets the maximum number of items a user can pick.
func (p *PickerBuilder) SetMaxItems(max int) *PickerBuilder {
	p.jsValue.Call("setMaxItems", max)
	return p
}

// Sets an OAuth token to use for authenticating the current user.
func (p *PickerBuilder) SetOAuthToken(token string) *PickerBuilder {
	p.jsValue.Call("setOAuthToken", token)
	return p
}

// Sets the origin of picker dialog. The origin should be set to the window.location.protocol + '//' + window.location.host of the top-most page, if your application is running in an iframe.
func (p *PickerBuilder) SetOrigin(origin string) *PickerBuilder {
	p.jsValue.Call("setOrigin", origin)
	return p
}

// Set the relay URL, used for gadgets.rpc.
func (p *PickerBuilder) SetRelayURL(url string) *PickerBuilder {
	p.jsValue.Call("setRelayUrl", url)
	return p
}

// Set the list of MIME types which will be selectable. Use commas to separate MIME types if more than one is required. If you don't set MIME types, files of all MIME types are displayed in the view.
func (p *PickerBuilder) SetSelectableMimeTypes(mimeTypes ...string) *PickerBuilder {
	p.jsValue.Call("setSelectableMimeTypes", strings.Join(mimeTypes, ","))
	return p
}

// Set the preferred dialog size. The dialog will be auto-centered. It has a minimum size of (566,350) and a maximum size of (1051,650).
func (p *PickerBuilder) SetSize(width, height int) *PickerBuilder {
	p.jsValue.Call("setSize", width, height)
	return p
}

// Set the dialog title.
func (p *PickerBuilder) SetTitle(title string) *PickerBuilder {
	p.jsValue.Call("setTitle", title)
	return p
}

// Returns the URI generated by this builder.
func (p *PickerBuilder) ToURI() string {
	return p.jsValue.Call("toUri").String()
}
