// Package appvalue provides a way to unmarshal app.Value objects into Go structs, similar to how json.Unmarshal and xml.Unmarshal work.
//
// # Struct Tags
//
// The following struct tags are supported:
//
// - javascript:"<name>": The name of the property in the javascript object (such as `javascript:"favoriteColor"`).
// - javascript:"-" : Skip the field.
// - javascript:"@this": Put the raw app.Value into this field.
//
// # Special Support
//
// The following special types are supported:
//
// - app.Value: This is a raw app.Value; no processing will be done on it.
// - appvalue.Timestamp: This is a `time.Time` value; you may cast it via `time.Time(value)`.
//
// # Example
//
// ```go
//
//	type User struct {
//		Name string `javascript:"name"`
//		Age  int    `javascript:"age"`
//	}
//
//	var jsValue app.Value
//	jsValue.Set("name", "John")
//	jsValue.Set("age", 30)
//
//	var user User
//	err := Unmarshal(jsValue, &user)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// ```
package appvalue
