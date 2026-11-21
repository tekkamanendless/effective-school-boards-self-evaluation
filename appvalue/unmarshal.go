package appvalue

import (
	"fmt"
	"log/slog"
	"reflect"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

// Unmarshaler is the interface implemented by objects that can unmarshal a Javascript object.
type Unmarshaler interface {
	UnmarshalJavascript(javascriptValue app.Value) error
}

// Unmarshal an app.Value into the target struct.
func Unmarshal(appValue app.Value, target any) error {
	slog.Debug("Unmarshal: Parsing javascript value", "appValue", appValue, "target", target, "targetType", fmt.Sprintf("%T", target))

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Pointer {
		return fmt.Errorf("target must be a pointer: %s", targetValue.Kind())
	}
	targetValue = targetValue.Elem()

	return parseValue(appValue, targetValue)
}

func parseValue(javascriptValue app.Value, targetValue reflect.Value) error {
	slog.Debug("parseValue: Parsing javascript value", "javascriptValue", javascriptValue, "targetType", targetValue.Type().String())

	if targetValue.Kind() != reflect.Pointer && targetValue.CanAddr() {
		if targetValue.Addr().Type().Implements(reflect.TypeFor[Unmarshaler]()) {
			slog.Debug("parseValue: Unmarshalling javascript value using Unmarshaler.", "javascriptValue", javascriptValue, "targetValue", targetValue, "targetType", targetValue.Type())

			return targetValue.Addr().Interface().(Unmarshaler).UnmarshalJavascript(javascriptValue)
		}
	}

	if targetValue.Type().Implements(reflect.TypeFor[Unmarshaler]()) {
		slog.Debug("parseValue: Unmarshalling javascript value using Unmarshaler.", "javascriptValue", javascriptValue, "targetValue", targetValue, "targetType", targetValue.Type())

		return targetValue.Interface().(Unmarshaler).UnmarshalJavascript(javascriptValue)
	}

	// If the field is an app.Value, we can just set it directly.
	if targetValue.Type().Implements(reflect.TypeFor[app.Value]()) {
		slog.Debug("parseValue: Setting javascript value directly.", "javascriptValue", javascriptValue, "targetValue", targetValue, "targetType", targetValue.Type())

		targetValue.Set(reflect.ValueOf(javascriptValue))
		return nil
	}

	slog.Debug("parseValue: javascriptValue.", "type", javascriptValue.Type())

	// If the value is null, we can just skip it.
	if javascriptValue.IsNull() {
		slog.Debug("parseValue: Skipping null javascript value.", "javascriptValue", javascriptValue, "targetValue", targetValue, "targetType", targetValue.Type())

		return nil
	}
	if javascriptValue.IsUndefined() {
		slog.Debug("parseValue: Skipping undefined javascript value.", "javascriptValue", javascriptValue, "targetValue", targetValue, "targetType", targetValue.Type())

		return nil
	}

	slog.Debug("parseValue: targetValue.", "kind", targetValue.Kind())

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
		targetType := targetValue.Type()
		for _, field := range reflect.VisibleFields(targetType) {
			javascriptTag := field.Tag.Get("javascript")
			if javascriptTag == "" {
				continue
			}

			javascriptPropertyName := javascriptTag
			slog.Debug("parseValue: Getting javascript property.", "javascriptPropertyName", javascriptPropertyName)

			if javascriptPropertyName == "-" {
				slog.Debug("parseValue: Skipping field.", "field", field)
				continue
			}

			javascriptValue := javascriptValue.Get(javascriptPropertyName)

			fieldValue := targetValue.FieldByName(field.Name)
			if !fieldValue.IsValid() {
				slog.Debug("parseValue: Field is not valid.", "field", field)
				continue
			}
			if !fieldValue.CanSet() {
				slog.Debug("parseValue: Field is not settable.", "field", field)
				continue
			}

			err := parseValue(javascriptValue, fieldValue)
			if err != nil {
				return fmt.Errorf("failed to parse field %s: %w", field.Name, err)
			}
		}
	case reflect.Slice:
		targetValue.Set(reflect.MakeSlice(targetValue.Type(), javascriptValue.Length(), javascriptValue.Length()))
		for i := 0; i < javascriptValue.Length(); i++ {
			err := Unmarshal(javascriptValue.Index(i), targetValue.Index(i).Addr().Interface())
			if err != nil {
				return fmt.Errorf("failed to parse slice at index %d: %w", i, err)
			}
		}
	case reflect.Map:
		// TODO: fieldValue.Set(reflect.MakeMap(fieldValue.Type()))
		return fmt.Errorf("map fields are not supported")
	}
	return nil
}
