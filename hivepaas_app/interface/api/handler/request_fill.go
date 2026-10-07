package handler

import "reflect"

// fillNilEmbedded gives a request the structs it embeds by pointer that its
// body left nil, at every depth. Request types embed a struct of their fields
// that way; JSON allocates it only when the body names one of them, and what
// reads the request next - ModifyRequest, Validate - takes it as there. Filled,
// the fields are empty, and refused as such: a body of `{}` is a 400 naming
// what is missing, not a panic.
func fillNilEmbedded(v reflect.Value) {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}
	t := v.Type()
	for i := range t.NumField() {
		field, value := t.Field(i), v.Field(i)
		if !field.Anonymous || !value.CanSet() {
			continue
		}
		switch {
		case field.Type.Kind() == reflect.Pointer && field.Type.Elem().Kind() == reflect.Struct:
			if value.IsNil() {
				value.Set(reflect.New(field.Type.Elem()))
			}
			fillNilEmbedded(value)
		case field.Type.Kind() == reflect.Struct:
			fillNilEmbedded(value)
		}
	}
}
