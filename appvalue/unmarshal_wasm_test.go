package appvalue

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/stretchr/testify/require"
)

func TestUnmarshal(t *testing.T) {
	handler :=
		tint.NewHandler(os.Stderr, &tint.Options{
			NoColor:   !isatty.IsTerminal(os.Stderr.Fd()),
			AddSource: true,
			Level:     slog.LevelDebug,
		})
	slog.SetDefault(slog.New(slog.NewMultiHandler(handler)))
	Debug(true)

	var jsObjectValue app.Value
	var jsStringValue app.Value
	var jsIntegerValue app.Value
	var jsBooleanValue app.Value
	var jsFloatValue app.Value
	{
		jsValue := app.Window().Get("Object").New()
		jsValue.Set("name", "John")
		jsValue.Set("age", 30)
		jsValue.Set("isStudent", true)
		jsValue.Set("height", 1.75)
		jsValue.Set("emailAddresses", []any{"john1@example.com", "john2@example.com"})
		jsValue.Set("phoneNumbers", []any{map[string]any{"type": "home", "value": "111-555-1000"}, map[string]any{"type": "work", "value": "111-555-2000"}})
		jsValue.Set("createdAt", app.Window().Get("Date").New(app.Window().Get("Date").Call("UTC", 2026, 0, 1, 12, 30, 0, 0)))

		jsObjectValue = jsValue
		jsStringValue = jsValue.Get("name")
		jsIntegerValue = jsValue.Get("age")
		jsBooleanValue = jsValue.Get("isStudent")
		jsFloatValue = jsValue.Get("height")
	}

	t.Run("String", func(t *testing.T) {
		jsValue := jsStringValue
		require.Equal(t, "John", jsValue.String())

		t.Run("*bool", func(t *testing.T) {
			var target bool
			err := Unmarshal(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, false, target)
		})
		t.Run("*float64", func(t *testing.T) {
			var target float64
			err := Unmarshal(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, 0.0, target)
		})
		t.Run("*int", func(t *testing.T) {
			var target int
			err := Unmarshal(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, 0, target)
		})
		t.Run("*string", func(t *testing.T) {
			var target string
			err := Unmarshal(jsValue, &target)
			require.NoError(t, err)
			require.Equal(t, "John", target)
		})
		t.Run("string", func(t *testing.T) {
			var target string
			err := Unmarshal(jsValue, target)
			require.Error(t, err)
			require.Equal(t, "", target)
		})
	})
	t.Run("Number", func(t *testing.T) {
		jsValue := jsIntegerValue
		require.Equal(t, 30, jsValue.Int())

		t.Run("*bool", func(t *testing.T) {
			var target bool
			err := Unmarshal(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, false, target)
		})
		t.Run("*float64", func(t *testing.T) {
			var target float64
			err := Unmarshal(jsValue, &target)
			require.NoError(t, err)
			require.Equal(t, 30.0, target)
		})
		t.Run("*int", func(t *testing.T) {
			var target int
			err := Unmarshal(jsValue, &target)
			require.NoError(t, err)
			require.Equal(t, 30, target)
		})
		t.Run("int", func(t *testing.T) {
			var target int
			err := Unmarshal(jsValue, target)
			require.Error(t, err)
			require.Equal(t, 0, target)
		})
		t.Run("*string", func(t *testing.T) {
			var target string
			err := Unmarshal(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, "", target)
		})
	})
	t.Run("Boolean", func(t *testing.T) {
		jsValue := jsBooleanValue
		require.True(t, jsValue.Bool())

		t.Run("*bool", func(t *testing.T) {
			var target bool
			err := Unmarshal(jsValue, &target)
			require.NoError(t, err)
			require.True(t, target)
		})
		t.Run("bool", func(t *testing.T) {
			var target bool
			err := Unmarshal(jsValue, target)
			require.Error(t, err)
			require.Equal(t, false, target)
		})
		t.Run("*float64", func(t *testing.T) {
			var target float64
			err := Unmarshal(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, 0.0, target)
		})
		t.Run("*int", func(t *testing.T) {
			var target int
			err := Unmarshal(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, 0, target)
		})
		t.Run("*string", func(t *testing.T) {
			var target string
			err := Unmarshal(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, "", target)
		})
	})
	t.Run("Float64", func(t *testing.T) {
		jsValue := jsFloatValue
		require.Equal(t, 1.75, jsValue.Float())

		t.Run("*bool", func(t *testing.T) {
			var target bool
			err := Unmarshal(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, false, target)
		})
		t.Run("*float64", func(t *testing.T) {
			var target float64
			err := Unmarshal(jsValue, &target)
			require.NoError(t, err)
			require.Equal(t, 1.75, target)
		})
		t.Run("float64", func(t *testing.T) {
			var target float64
			err := Unmarshal(jsValue, target)
			require.Error(t, err)
			require.Equal(t, 0.0, target)
		})
		t.Run("*int", func(t *testing.T) {
			var target int
			err := Unmarshal(jsValue, &target)
			require.NoError(t, err)
			require.Equal(t, 1, target)
		})
		t.Run("*string", func(t *testing.T) {
			var target string
			err := Unmarshal(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, "", target)
		})
	})
	t.Run("Object", func(t *testing.T) {
		t.Run("Perfect struct", func(t *testing.T) {
			type hasRawValue struct {
				RawValue app.Value `javascript:"@this"`
			}
			type base struct {
				Name string `javascript:"name"`
				Age  int    `javascript:"age"`
			}
			var target struct {
				hasRawValue
				base
				IsStudent      bool     `javascript:"isStudent"`
				Height         float64  `javascript:"height"`
				EmailAddresses []string `javascript:"emailAddresses"`
				PhoneNumbers   []struct {
					Type  string `javascript:"type"`
					Value string `javascript:"value"`
				} `javascript:"phoneNumbers"`
				CreatedAt Timestamp `javascript:"createdAt"`
			}
			err := Unmarshal(jsObjectValue, &target)
			require.NoError(t, err)
			require.Equal(t, jsObjectValue, target.RawValue)
			require.Equal(t, "John", target.Name)
			require.Equal(t, 30, target.Age)
			require.True(t, target.IsStudent)
			require.Equal(t, 1.75, target.Height)
			require.Equal(t, []string{"john1@example.com", "john2@example.com"}, target.EmailAddresses)
			require.Equal(t, []struct {
				Type  string `javascript:"type"`
				Value string `javascript:"value"`
			}{
				{Type: "home", Value: "111-555-1000"},
				{Type: "work", Value: "111-555-2000"},
			}, target.PhoneNumbers)
			require.Equal(t, Timestamp(time.Date(2026, time.January, 1, 12, 30, 0, 0, time.UTC)), target.CreatedAt)
		})
		t.Run("Struct with ignored field", func(t *testing.T) {
			var target struct {
				Name           string   `javascript:"-"`
				Age            int      `javascript:"age"`
				IsStudent      bool     `javascript:"isStudent"`
				Height         float64  `javascript:"height"`
				EmailAddresses []string `javascript:"emailAddresses"`
				PhoneNumbers   []struct {
					Type  string `javascript:"type"`
					Value string `javascript:"value"`
				} `javascript:"phoneNumbers"`
				CreatedAt Timestamp `javascript:"createdAt"`
			}
			err := Unmarshal(jsObjectValue, &target)
			require.NoError(t, err)
			require.Equal(t, "", target.Name)
			require.Equal(t, 30, target.Age)
			require.True(t, target.IsStudent)
			require.Equal(t, 1.75, target.Height)
			require.Equal(t, []string{"john1@example.com", "john2@example.com"}, target.EmailAddresses)
			require.Equal(t, []struct {
				Type  string `javascript:"type"`
				Value string `javascript:"value"`
			}{
				{Type: "home", Value: "111-555-1000"},
				{Type: "work", Value: "111-555-2000"},
			}, target.PhoneNumbers)
		})
		t.Run("Struct with missing field", func(t *testing.T) {
			var target struct {
				Name           string
				Age            int      `javascript:"age"`
				IsStudent      bool     `javascript:"isStudent"`
				Height         float64  `javascript:"height"`
				EmailAddresses []string `javascript:"emailAddresses"`
				PhoneNumbers   []struct {
					Type  string `javascript:"type"`
					Value string `javascript:"value"`
				} `javascript:"phoneNumbers"`
				CreatedAt Timestamp `javascript:"createdAt"`
			}
			err := Unmarshal(jsObjectValue, &target)
			require.NoError(t, err)
			require.Equal(t, "", target.Name)
			require.Equal(t, 30, target.Age)
			require.True(t, target.IsStudent)
			require.Equal(t, 1.75, target.Height)
			require.Equal(t, []string{"john1@example.com", "john2@example.com"}, target.EmailAddresses)
			require.Equal(t, []struct {
				Type  string `javascript:"type"`
				Value string `javascript:"value"`
			}{
				{Type: "home", Value: "111-555-1000"},
				{Type: "work", Value: "111-555-2000"},
			}, target.PhoneNumbers)
		})
		t.Run("Map", func(t *testing.T) {
			var target map[string]any
			err := Unmarshal(jsObjectValue, &target)
			require.NoError(t, err)
			require.NotNil(t, target)
			require.Equal(t, "John", target["name"])
			require.Equal(t, 30.0, target["age"])
			require.Equal(t, true, target["isStudent"])
			require.Equal(t, 1.75, target["height"])
			require.Equal(t, []any{"john1@example.com", "john2@example.com"}, target["emailAddresses"])
			require.Equal(t, []any{map[string]any{"type": "home", "value": "111-555-1000"}, map[string]any{"type": "work", "value": "111-555-2000"}}, target["phoneNumbers"])
			require.Equal(t, Timestamp(time.Date(2026, time.January, 1, 12, 30, 0, 0, time.UTC)), target["createdAt"])
		})
	})
}
