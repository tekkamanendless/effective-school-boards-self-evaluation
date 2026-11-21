package parsejavascript

import (
	"fmt"
	"reflect"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

// Parse an app.Value into the target struct.
func Parse(appValue app.Value, target any) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Pointer {
		return fmt.Errorf("target must be a pointer: %s", targetValue.Kind())
	}
	targetValue = targetValue.Elem()
	if targetValue.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to a struct: %s", targetValue.Kind())
	}

	targetType := targetValue.Type()
	for _, field := range reflect.VisibleFields(targetType) {
		javascriptTag := field.Tag.Get("javascript")
		if javascriptTag == "" {
			continue
		}

		javascriptValue := appValue.Get(javascriptTag)
		if javascriptValue.IsUndefined() {
			continue
		}
		if javascriptValue.IsNull() {
			continue
		}

		fieldValue := targetValue.FieldByName(field.Name)
		if !fieldValue.IsValid() {
			continue
		}
		if !fieldValue.CanSet() {
			continue
		}

		// If the field is an app.Value, we can just set it directly.
		if fieldValue.Type().Implements(reflect.TypeFor[app.Value]()) {
			fieldValue.Set(reflect.ValueOf(javascriptValue))
			continue
		}

		// If the value is null, we can just skip it.
		if javascriptValue.IsNull() {
			continue
		}

		switch fieldValue.Kind() {
		case reflect.String:
			if javascriptValue.Type() != app.TypeString {
				return fmt.Errorf("field %s is a string, but the event value is not a string: %s", field.Name, javascriptValue.Type())
			}
			fieldValue.SetString(javascriptValue.String())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if javascriptValue.Type() != app.TypeNumber {
				return fmt.Errorf("field %s is a number, but the event value is not a number: %s", field.Name, javascriptValue.Type())
			}
			fieldValue.SetInt(int64(javascriptValue.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if javascriptValue.Type() != app.TypeNumber {
				return fmt.Errorf("field %s is a number, but the event value is not a number: %s", field.Name, javascriptValue.Type())
			}
			fieldValue.SetUint(uint64(javascriptValue.Int()))
		case reflect.Float32, reflect.Float64:
			if javascriptValue.Type() != app.TypeNumber {
				return fmt.Errorf("field %s is a number, but the event value is not a number: %s", field.Name, javascriptValue.Type())
			}
			fieldValue.SetFloat(javascriptValue.Float())
		case reflect.Bool:
			if javascriptValue.Type() != app.TypeBoolean {
				return fmt.Errorf("field %s is a boolean, but the event value is not a boolean: %s", field.Name, javascriptValue.Type())
			}
			fieldValue.SetBool(javascriptValue.Bool())
		case reflect.Struct:
			err := Parse(javascriptValue, fieldValue.Addr().Interface())
			if err != nil {
				return fmt.Errorf("failed to parse struct field %s: %w", field.Name, err)
			}
		case reflect.Slice:
			fieldValue.Set(reflect.MakeSlice(fieldValue.Type(), javascriptValue.Length(), javascriptValue.Length()))
			for i := 0; i < javascriptValue.Length(); i++ {
				err := Parse(javascriptValue.Index(i), fieldValue.Index(i).Addr().Interface())
				if err != nil {
					return fmt.Errorf("failed to parse slice field %s at index %d: %w", field.Name, i, err)
				}
			}
		case reflect.Map:
			// TODO: fieldValue.Set(reflect.MakeMap(fieldValue.Type()))
			return fmt.Errorf("map fields are not supported")
		}
	}
	return nil
}
