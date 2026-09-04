package entity

import (
	"reflect"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

var encryptedFieldType = reflect.TypeFor[EncryptedField]()

// ReencryptData re-encrypts every secret a setting holds with the current app
// secret, and reports whether anything changed.
//
// It walks the parsed data rather than the raw JSON on purpose. HashField stores
// its value behind the same `hpsalt:` prefix as an encrypted one, but it is a
// hash and cannot be decrypted; a textual pass over the JSON would try to
// re-encrypt it and destroy every API key. Walking typed values touches only the
// fields that really are encrypted.
func (s *Setting) ReencryptData() (changed bool, err error) {
	if s.Data == "" {
		return false, nil
	}

	data, err := s.Parse()
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	if data == nil {
		return false, nil
	}

	count, err := reencryptValue(reflect.ValueOf(data))
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	if count == 0 {
		return false, nil
	}

	if err := s.SetData(data); err != nil {
		return false, hperrors.Wrap(err)
	}
	return true, nil
}

// reencryptValue re-encrypts every EncryptedField reachable from v, returning how
// many it touched.
func reencryptValue(value reflect.Value) (int, error) {
	switch value.Kind() { //nolint:exhaustive
	case reflect.Pointer, reflect.Interface:
		if value.IsNil() {
			return 0, nil
		}
		return reencryptValue(value.Elem())

	case reflect.Struct:
		if value.Type() == encryptedFieldType {
			return reencryptField(value)
		}
		return reencryptStructFields(value)

	case reflect.Slice, reflect.Array:
		count := 0
		for i := range value.Len() {
			n, err := reencryptValue(value.Index(i))
			if err != nil {
				return 0, err
			}
			count += n
		}
		return count, nil

	case reflect.Map:
		return reencryptMapValues(value)

	default:
		return 0, nil
	}
}

func reencryptStructFields(value reflect.Value) (int, error) {
	count := 0
	for i := range value.NumField() {
		if !value.Type().Field(i).IsExported() {
			continue // an unexported field cannot hold data we serialize
		}
		n, err := reencryptValue(value.Field(i))
		if err != nil {
			return 0, err
		}
		count += n
	}
	return count, nil
}

// reencryptMapValues handles a map whose values contain encrypted fields. Map
// values are not addressable, so each one is copied out, rewritten and put back.
func reencryptMapValues(value reflect.Value) (int, error) {
	if value.IsNil() || !typeHoldsEncryptedField(value.Type().Elem()) {
		return 0, nil
	}

	count := 0
	for _, key := range value.MapKeys() {
		entry := reflect.New(value.Type().Elem()).Elem()
		entry.Set(value.MapIndex(key))

		n, err := reencryptValue(entry)
		if err != nil {
			return 0, err
		}
		if n > 0 {
			value.SetMapIndex(key, entry)
			count += n
		}
	}
	return count, nil
}

// reencryptField rewrites one field. It has to be addressable, which every field
// reached from the pointer Parse returns is.
func reencryptField(value reflect.Value) (int, error) {
	if !value.CanAddr() {
		return 0, hperrors.Wrap(hperrors.ErrInternal).
			WithMsgLog("encrypted field is not addressable, cannot re-encrypt it")
	}

	field, _ := value.Addr().Interface().(*EncryptedField)
	if field.IsEmpty() {
		return 0, nil
	}
	if err := field.Reencrypt(); err != nil {
		return 0, hperrors.Wrap(err)
	}
	return 1, nil
}

// typeHoldsEncryptedField reports whether a type can contain an EncryptedField,
// so map entries that cannot are left untouched.
func typeHoldsEncryptedField(typ reflect.Type) bool {
	switch typ.Kind() { //nolint:exhaustive
	case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Map:
		return typeHoldsEncryptedField(typ.Elem())
	case reflect.Interface:
		return true // cannot tell without a value
	case reflect.Struct:
		if typ == encryptedFieldType {
			return true
		}
		for i := range typ.NumField() {
			if typ.Field(i).IsExported() && typeHoldsEncryptedField(typ.Field(i).Type) {
				return true
			}
		}
		return false
	default:
		return false
	}
}
