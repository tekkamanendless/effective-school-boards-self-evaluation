package appvalue

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type Timestamp time.Time

func (t *Timestamp) UnmarshalJavascript(javascriptValue app.Value) error {
	if debug {
		slog.Debug("Timestamp.UnmarshalJavascript: Unmarshalling javascript value", "javascriptValue", javascriptValue, "javascriptValueType", javascriptValue.Type().String())
	}

	if javascriptValue.Type().String() == "object" && javascriptValue.Get("constructor").Get("name").String() == "Date" {
		if debug {
			slog.Debug("Timestamp.UnmarshalJavascript: Javascript value is a Date object")
		}

		milliseconds := javascriptValue.Call("getTime").Int()
		*t = Timestamp(time.Unix(0, int64(milliseconds)*int64(time.Millisecond)).UTC())
		return nil
	}
	{
		var milliseconds int64
		err := Unmarshal(javascriptValue, &milliseconds)
		if err == nil {
			if debug {
				slog.Debug("Timestamp.UnmarshalJavascript: Javascript value is a number", "milliseconds", milliseconds)
			}

			*t = Timestamp(time.Unix(0, milliseconds*int64(time.Millisecond)).UTC())
			return nil
		}
	}
	{
		var stringValue string
		err := Unmarshal(javascriptValue, &stringValue)
		if err == nil {
			formats := []string{
				time.RFC3339Nano,
				time.RFC3339,
			}
			for _, format := range formats {
				parsedTime, err := time.Parse(format, stringValue)
				if err == nil {
					if debug {
						slog.Debug("Timestamp.UnmarshalJavascript: Javascript value is a string", "stringValue", stringValue, "parsedTime", parsedTime)
					}

					*t = Timestamp(parsedTime.UTC())
					return nil
				}
			}
		}
	}
	return fmt.Errorf("failed to unmarshal timestamp")
}
