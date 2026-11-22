package googlepickerjs

import (
	"context"
	"log/slog"

	"github.com/go-app-blazar/appvalue"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type Picker struct {
	jsValue app.Value // The underlying JavaScript value.
}

// JSValue returns the underlying JavaScript value.
func (p *Picker) JSValue() app.Value {
	return p.jsValue
}

// Disposes the Picker object.
func (p *Picker) Dispose() {
	p.jsValue.Call("dispose")
}

// Get a boolean indicating the current Picker visibility.
func (p *Picker) IsVisible() bool {
	return p.jsValue.Call("isVisible").Bool()
}

// Specify the callback method called whenever the user has selected an item (or canceled.)
func (p *Picker) SetCallback(callback func(response *ResponseObject)) {
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
}

// Specify a relay URL to circumvent cross-domain issues.
func (p *Picker) SetRelayUrl(url string) {
	p.jsValue.Call("setRelayUrl", url)
}

// Control the visibility of the Picker object.
func (p *Picker) SetVisible(visible bool) {
	p.jsValue.Call("setVisible", visible)
}
