package appvalue

import (
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"time"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

// Unmarshaler is the interface implemented by objects that can unmarshal a Javascript object.
type Unmarshaler interface {
	UnmarshalJavascript(javascriptValue app.Value) error
}

// Unmarshal an app.Value into the target struct.
func Unmarshal(appValue app.Value, target any) error {
	if debug {
		slog.Debug("Unmarshal: Parsing javascript value", "appValue", appValue, "target", target, "targetType", fmt.Sprintf("%T", target))
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Pointer {
		return fmt.Errorf("target must be a pointer: %s", targetValue.Kind())
	}
	targetValue = targetValue.Elem()

	return parseValue(appValue, targetValue)
}

// parseValue parses a javascript value into a target value.
func parseValue(javascriptValue app.Value, targetValue reflect.Value) error {
	if debug {
		slog.Debug("parseValue: Parsing javascript value", "javascriptValue", javascriptValue, "targetType", targetValue.Type().String())
	}

	if targetValue.Kind() != reflect.Pointer && targetValue.CanAddr() {
		if targetValue.Addr().Type().Implements(reflect.TypeFor[Unmarshaler]()) {
			if debug {
				slog.Debug("parseValue: Unmarshalling javascript value using Unmarshaler.", "javascriptValue", javascriptValue, "targetValue", targetValue, "targetType", targetValue.Type())
			}

			return targetValue.Addr().Interface().(Unmarshaler).UnmarshalJavascript(javascriptValue)
		}
	}

	if targetValue.Type().Implements(reflect.TypeFor[Unmarshaler]()) {
		if debug {
			slog.Debug("parseValue: Unmarshalling javascript value using Unmarshaler.", "javascriptValue", javascriptValue, "targetValue", targetValue, "targetType", targetValue.Type())
		}

		return targetValue.Interface().(Unmarshaler).UnmarshalJavascript(javascriptValue)
	}

	// If the field is an app.Value, we can just set it directly.
	if targetValue.Type().Implements(reflect.TypeFor[app.Value]()) {
		if debug {
			slog.Debug("parseValue: Setting javascript value directly.", "javascriptValue", javascriptValue, "targetValue", targetValue, "targetType", targetValue.Type())
		}

		targetValue.Set(reflect.ValueOf(javascriptValue))
		return nil
	}

	if debug {
		slog.Debug("parseValue: javascriptValue.", "type", javascriptValue.Type())
	}

	// If the value is null, we can just skip it.
	if javascriptValue.IsNull() {
		if debug {
			slog.Debug("parseValue: Skipping null javascript value.", "javascriptValue", javascriptValue, "targetValue", targetValue, "targetType", targetValue.Type())
		}

		return nil
	}
	if javascriptValue.IsUndefined() {
		if debug {
			slog.Debug("parseValue: Skipping undefined javascript value.", "javascriptValue", javascriptValue, "targetValue", targetValue, "targetType", targetValue.Type())
		}

		return nil
	}

	if debug {
		slog.Debug("parseValue: targetValue.", "kind", targetValue.Kind())
	}

	switch targetValue.Kind() {
	case reflect.String:
		if javascriptValue.Type() != app.TypeString {
			return fmt.Errorf("javascript value is not a string: %s", javascriptValue.Type())
		}
		targetValue.SetString(javascriptValue.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if javascriptValue.Type() != app.TypeNumber {
			return fmt.Errorf("javascript value is not an number: %s", javascriptValue.Type())
		}
		targetValue.SetInt(int64(javascriptValue.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if javascriptValue.Type() != app.TypeNumber {
			return fmt.Errorf("javascript value is not a number: %s", javascriptValue.Type())
		}
		targetValue.SetUint(uint64(javascriptValue.Int()))
	case reflect.Float32, reflect.Float64:
		if javascriptValue.Type() != app.TypeNumber {
			return fmt.Errorf("javascript value is not a number: %s", javascriptValue.Type())
		}
		targetValue.SetFloat(javascriptValue.Float())
	case reflect.Bool:
		if javascriptValue.Type() != app.TypeBoolean {
			return fmt.Errorf("javascript value is not a boolean: %s", javascriptValue.Type())
		}
		targetValue.SetBool(javascriptValue.Bool())
	case reflect.Struct:
		if javascriptValue.Type().String() != "object" {
			return fmt.Errorf("javascript value is not an object: %s", javascriptValue.Type().String())
		}

		targetType := targetValue.Type()
		for _, field := range reflect.VisibleFields(targetType) {
			javascriptTag := field.Tag.Get("javascript")
			if javascriptTag == "" {
				continue
			}

			javascriptPropertyName := javascriptTag
			if debug {
				slog.Debug("parseValue: Getting javascript property.", "javascriptPropertyName", javascriptPropertyName)
			}

			if javascriptPropertyName == "-" {
				if debug {
					slog.Debug("parseValue: Skipping field.", "field", field)
				}
				continue
			}

			fieldValue := targetValue.FieldByName(field.Name)
			if !fieldValue.IsValid() {
				if debug {
					slog.Debug("parseValue: Field is not valid.", "field", field)
				}
				continue
			}
			if !fieldValue.CanSet() {
				if debug {
					slog.Debug("parseValue: Field is not settable.", "field", field)
				}
				continue
			}

			var jsPropertyValue app.Value
			if javascriptPropertyName == "@this" {
				if debug {
					slog.Debug("parseValue: Using javascript value as the target value.", "javascriptValue", javascriptValue, "targetValue", targetValue, "targetType", targetValue.Type())
				}

				jsPropertyValue = javascriptValue
			} else {
				jsPropertyValue = javascriptValue.Get(javascriptPropertyName)
			}

			err := parseValue(jsPropertyValue, fieldValue)
			if err != nil {
				return fmt.Errorf("failed to parse field %s: %w", field.Name, err)
			}
		}
	case reflect.Slice:
		if javascriptValue.Type().String() != "object" {
			return fmt.Errorf("javascript value is not an object: %s", javascriptValue.Type().String())
		}

		targetValue.Set(reflect.MakeSlice(targetValue.Type(), javascriptValue.Length(), javascriptValue.Length()))
		for i := 0; i < javascriptValue.Length(); i++ {
			err := Unmarshal(javascriptValue.Index(i), targetValue.Index(i).Addr().Interface())
			if err != nil {
				return fmt.Errorf("failed to parse slice at index %d: %w", i, err)
			}
		}
	case reflect.Map:
		if javascriptValue.Type().String() != "object" {
			return fmt.Errorf("javascript value is not an object: %s", javascriptValue.Type().String())
		}

		if debug {
			slog.Debug("parseValue: Target is a map.", "javascriptValue", javascriptValue, "targetValue", targetValue, "targetType", targetValue.Type())
		}

		targetValue.Set(reflect.MakeMap(targetValue.Type()))

		keys := getObjectKeys(javascriptValue)
		if debug {
			slog.Debug("parseValue: Keys", "keys", keys)
		}

		for _, key := range keys {
			jsMapPropertyValue := javascriptValue.Get(key)

			if targetValue.Type().Elem() == reflect.TypeFor[any]() {
				if debug {
					slog.Debug("parseValue: Target is an any.", "javascriptValue", jsMapPropertyValue, "targetValue", targetValue, "targetType", targetValue.Type())
				}

				nativePropertyValue, err := valueToNative(jsMapPropertyValue)
				if err != nil {
					return fmt.Errorf("failed to convert javascript value to native: %w", err)
				}

				targetValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(nativePropertyValue))
				continue
			}

			err := Unmarshal(jsMapPropertyValue, targetValue.MapIndex(reflect.ValueOf(key)).Addr().Interface())
			if err != nil {
				return fmt.Errorf("failed to parse map at key %s: %w", key, err)
			}
		}
	default:
		return fmt.Errorf("unsupported target value kind: %s", targetValue.Kind())
	}
	return nil
}

// valueToNative converts a javascript value to a native Go value.
func valueToNative(javascriptValue app.Value) (any, error) {
	switch javascriptValue.Type() {
	case app.TypeBoolean:
		return javascriptValue.Bool(), nil
	case app.TypeFunction:
		// TODO: Implement function to native conversion.
		return javascriptValue, nil
	case app.TypeNumber:
		return javascriptValue.Float(), nil
	case app.TypeObject:
		switch javascriptValue.Get("constructor").Get("name").String() {
		case "Array":
			if debug {
				slog.Debug("valueToNative: Javascript value is an Array object")
			}

			newValue := reflect.MakeSlice(reflect.TypeFor[[]any](), javascriptValue.Length(), javascriptValue.Length())
			for i := 0; i < javascriptValue.Length(); i++ {
				nativeValue, err := valueToNative(javascriptValue.Index(i))
				if err != nil {
					return nil, fmt.Errorf("failed to convert javascript value to native: %w", err)
				}
				newValue.Index(i).Set(reflect.ValueOf(nativeValue))
			}
			return newValue.Interface(), nil
		case "Date":
			if debug {
				slog.Debug("valueToNative: Javascript value is a Date object")
			}

			milliseconds := javascriptValue.Call("getTime").Int()
			return Timestamp(time.Unix(0, int64(milliseconds)*int64(time.Millisecond)).UTC()), nil
		case "Object":
			if debug {
				slog.Debug("valueToNative: Javascript value is an Object object")
			}

			newValue := reflect.MakeMap(reflect.TypeFor[map[string]any]())

			keys := getObjectKeys(javascriptValue)
			if debug {
				slog.Debug("valueToNative: Keys", "keys", keys)
			}

			for _, key := range keys {
				nativeValue, err := valueToNative(javascriptValue.Get(key))
				if err != nil {
					return nil, fmt.Errorf("failed to convert javascript value to native: %w", err)
				}
				newValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(nativeValue))
			}
			return newValue.Interface(), nil
		}
		return javascriptValue, nil
	case app.TypeNull:
		return nil, nil
	case app.TypeString:
		return javascriptValue.String(), nil
	case app.TypeUndefined:
		return nil, nil
	}
	return nil, fmt.Errorf("unsupported javascript value type: %s", javascriptValue.Type())
}

// getObjectKeys returns the keys of an object.
func getObjectKeys(javascriptValue app.Value) []string {
	var keys []string
	{
		jsKeys := app.Window().Get("Object").Call("keys", javascriptValue)
		for i := 0; i < jsKeys.Length(); i++ {
			key := jsKeys.Index(i).String()
			keys = append(keys, key)
		}
		slices.Sort(keys)
	}
	return keys
}
