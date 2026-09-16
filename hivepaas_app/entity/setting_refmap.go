package entity

import (
	"reflect"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// RemapRefs rewrites every reference a setting holds, from the identifiers it
// was exported with to the identifiers it is being imported as.
//
// It walks by value rather than by type. References are stored in at least four
// different shapes - ObjectID, ObjectValue, ObjectIDSlice, and bespoke structs
// carrying a bare `ID string` such as SystemBackupCloudStorage - and a walk that
// recognized only the declared types would silently skip the last group. What
// every shape has in common is the identifier itself, and GetRefObjectIDs
// already reports exactly which identifiers are references, so that is what the
// walk matches on.
//
// After rewriting, the references are read back and compared. A reference the
// walk could not reach fails here, loudly, instead of importing as a pointer
// into another installation's data.
//
// Unexported fields are skipped, which is what keeps an EncryptedField's
// contents out of reach: a secret whose plaintext happened to equal a reference
// id would otherwise be rewritten beyond recovery.
func RemapRefs(data SettingData, mapping map[string]string) error {
	if data == nil || len(mapping) == 0 {
		return nil
	}

	before := collectRefIDs(data)
	remapStringValues(reflect.ValueOf(data), mapping)

	want := make(map[string]bool, len(before))
	for _, id := range before {
		want[gofn.Coalesce(mapping[id], id)] = true
	}

	for _, id := range collectRefIDs(data) {
		if !want[id] {
			return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog(
				"reference %q was not rewritten by RemapRefs; the setting type holds it "+
					"somewhere the walk cannot reach", id)
		}
	}
	return nil
}

func collectRefIDs(data SettingData) []string {
	refIDs := data.GetRefObjectIDs()
	if refIDs == nil {
		return nil
	}
	all := make([]string, 0,
		len(refIDs.RefSettingIDs)+len(refIDs.RefAppIDs)+len(refIDs.RefProjectIDs)+
			len(refIDs.RefProjectEnvIDs)+len(refIDs.RefUserIDs))
	all = append(all, refIDs.RefSettingIDs...)
	all = append(all, refIDs.RefAppIDs...)
	all = append(all, refIDs.RefProjectIDs...)
	all = append(all, refIDs.RefProjectEnvIDs...)
	all = append(all, refIDs.RefUserIDs...)
	return all
}

// remapStringValues replaces every settable string equal to a mapping key.
//
// The shape of this walk follows reencryptValue in setting_reencrypt.go, which
// solved the same traversal problem for encrypted fields.
func remapStringValues(value reflect.Value, mapping map[string]string) {
	switch value.Kind() { //nolint:exhaustive // reflect.Kind: only containers and strings matter to this walk
	case reflect.Pointer, reflect.Interface:
		if !value.IsNil() {
			remapStringValues(value.Elem(), mapping)
		}

	case reflect.Struct:
		for i := range value.NumField() {
			if !value.Type().Field(i).IsExported() {
				continue // an unexported field cannot hold data we serialize
			}
			remapStringValues(value.Field(i), mapping)
		}

	case reflect.Slice, reflect.Array:
		for i := range value.Len() {
			remapStringValues(value.Index(i), mapping)
		}

	case reflect.Map:
		remapMapValues(value, mapping)

	case reflect.String:
		if !value.CanSet() {
			return
		}
		if replacement, ok := mapping[value.String()]; ok {
			value.SetString(replacement)
		}
	}
}

// remapMapValues handles maps, whose values are not addressable and so have to
// be copied out, rewritten and put back.
func remapMapValues(value reflect.Value, mapping map[string]string) {
	if value.IsNil() {
		return
	}
	for _, key := range value.MapKeys() {
		entry := reflect.New(value.Type().Elem()).Elem()
		entry.Set(value.MapIndex(key))
		remapStringValues(entry, mapping)
		value.SetMapIndex(key, entry)
	}
}
