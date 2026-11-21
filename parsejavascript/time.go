package parsejavascript

import (
	"fmt"
	"time"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type Timestamp time.Time

func (t *Timestamp) UnmarshalJavascript(javascriptValue app.Value) error {
	{
		var milliseconds int64
		err := Unmarshal(javascriptValue, &milliseconds)
		if err == nil {
			*t = Timestamp(time.Unix(0, milliseconds*int64(time.Millisecond)))
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
					*t = Timestamp(parsedTime)
					return nil
				}
			}
		}
	}
	return fmt.Errorf("failed to unmarshal timestamp")
}
