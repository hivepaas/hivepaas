package jsonutil

import (
	"reflect"
)

// Contains checks if the 'expected' object is a subset of the 'actual' object.
// For maps, it checks if all keys in 'expected' exist in 'actual' and their values match recursively.
// For all other types (slices, primitives), it falls back to reflect.DeepEqual for exact matching.
func Contains(actual, expected any) bool {
	if expected == nil {
		return true // an empty requirement is technically satisfied by anything
	}
	if actual == nil {
		return false
	}

	actualVal := reflect.ValueOf(actual)
	expectedVal := reflect.ValueOf(expected)

	// If they are maps, perform recursive subset checking
	if actualVal.Kind() == reflect.Map && expectedVal.Kind() == reflect.Map {
		iter := expectedVal.MapRange()
		for iter.Next() {
			k := iter.Key()
			expectedV := iter.Value()
			actualV := actualVal.MapIndex(k)

			if !actualV.IsValid() {
				// Key does not exist in actual
				return false
			}

			if !Contains(actualV.Interface(), expectedV.Interface()) {
				return false
			}
		}
		return true
	}

	// For all other types, rely on DeepEqual
	return reflect.DeepEqual(actual, expected)
}
