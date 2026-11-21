package parsejavascript

import (
	"log/slog"
	"os"
	"testing"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	handler :=
		tint.NewHandler(os.Stderr, &tint.Options{
			NoColor:   !isatty.IsTerminal(os.Stderr.Fd()),
			AddSource: true,
			Level:     slog.LevelDebug,
		})
	slog.SetDefault(slog.New(slog.NewMultiHandler(handler)))

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
			err := Parse(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, false, target)
		})
		t.Run("*float64", func(t *testing.T) {
			var target float64
			err := Parse(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, 0.0, target)
		})
		t.Run("*int", func(t *testing.T) {
			var target int
			err := Parse(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, 0, target)
		})
		t.Run("*string", func(t *testing.T) {
			var target string
			err := Parse(jsValue, &target)
			require.NoError(t, err)
			require.Equal(t, "John", target)
		})
		t.Run("string", func(t *testing.T) {
			var target string
			err := Parse(jsValue, target)
			require.Error(t, err)
			require.Equal(t, "", target)
		})
	})
	t.Run("Number", func(t *testing.T) {
		jsValue := jsIntegerValue
		require.Equal(t, 30, jsValue.Int())

		t.Run("*bool", func(t *testing.T) {
			var target bool
			err := Parse(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, false, target)
		})
		t.Run("*float64", func(t *testing.T) {
			var target float64
			err := Parse(jsValue, &target)
			require.NoError(t, err)
			require.Equal(t, 30.0, target)
		})
		t.Run("*int", func(t *testing.T) {
			var target int
			err := Parse(jsValue, &target)
			require.NoError(t, err)
			require.Equal(t, 30, target)
		})
		t.Run("int", func(t *testing.T) {
			var target int
			err := Parse(jsValue, target)
			require.Error(t, err)
			require.Equal(t, 0, target)
		})
		t.Run("*string", func(t *testing.T) {
			var target string
			err := Parse(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, "", target)
		})
	})
	t.Run("Boolean", func(t *testing.T) {
		jsValue := jsBooleanValue
		require.True(t, jsValue.Bool())

		t.Run("*bool", func(t *testing.T) {
			var target bool
			err := Parse(jsValue, &target)
			require.NoError(t, err)
			require.True(t, target)
		})
		t.Run("bool", func(t *testing.T) {
			var target bool
			err := Parse(jsValue, target)
			require.Error(t, err)
			require.Equal(t, false, target)
		})
		t.Run("*float64", func(t *testing.T) {
			var target float64
			err := Parse(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, 0.0, target)
		})
		t.Run("*int", func(t *testing.T) {
			var target int
			err := Parse(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, 0, target)
		})
		t.Run("*string", func(t *testing.T) {
			var target string
			err := Parse(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, "", target)
		})
	})
	t.Run("Float64", func(t *testing.T) {
		jsValue := jsFloatValue
		require.Equal(t, 1.75, jsValue.Float())

		t.Run("*bool", func(t *testing.T) {
			var target bool
			err := Parse(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, false, target)
		})
		t.Run("*float64", func(t *testing.T) {
			var target float64
			err := Parse(jsValue, &target)
			require.NoError(t, err)
			require.Equal(t, 1.75, target)
		})
		t.Run("float64", func(t *testing.T) {
			var target float64
			err := Parse(jsValue, target)
			require.Error(t, err)
			require.Equal(t, 0.0, target)
		})
		t.Run("*int", func(t *testing.T) {
			var target int
			err := Parse(jsValue, &target)
			require.NoError(t, err)
			require.Equal(t, 1, target)
		})
		t.Run("*string", func(t *testing.T) {
			var target string
			err := Parse(jsValue, &target)
			require.Error(t, err)
			require.Equal(t, "", target)
		})
	})
	t.Run("Object", func(t *testing.T) {
		jsValue := app.Window().Get("Object").New()
		jsValue.Set("name", "John")
		jsValue.Set("age", 30)
		jsValue.Set("isStudent", true)
		jsValue.Set("height", 1.75)

		t.Run("Perfect struct", func(t *testing.T) {
			var target struct {
				Name      string  `javascript:"name"`
				Age       int     `javascript:"age"`
				IsStudent bool    `javascript:"isStudent"`
				Height    float64 `javascript:"height"`
			}
			err := Parse(jsValue, &target)
			require.NoError(t, err)
			require.Equal(t, "John", target.Name)
			require.Equal(t, 30, target.Age)
			require.True(t, target.IsStudent)
			require.Equal(t, 1.75, target.Height)
		})
		t.Run("Struct with ignored field", func(t *testing.T) {
			var target struct {
				Name      string  `javascript:"-"`
				Age       int     `javascript:"age"`
				IsStudent bool    `javascript:"isStudent"`
				Height    float64 `javascript:"height"`
			}
			err := Parse(jsValue, &target)
			require.NoError(t, err)
			require.Equal(t, "", target.Name)
			require.Equal(t, 30, target.Age)
			require.True(t, target.IsStudent)
			require.Equal(t, 1.75, target.Height)
		})
		t.Run("Struct with missing field", func(t *testing.T) {
			var target struct {
				Name      string
				Age       int     `javascript:"age"`
				IsStudent bool    `javascript:"isStudent"`
				Height    float64 `javascript:"height"`
			}
			err := Parse(jsValue, &target)
			require.NoError(t, err)
			require.Equal(t, "", target.Name)
			require.Equal(t, 30, target.Age)
			require.True(t, target.IsStudent)
			require.Equal(t, 1.75, target.Height)
		})
	})
}
