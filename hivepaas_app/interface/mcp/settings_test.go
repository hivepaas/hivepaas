package mcp

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// fill sets every exported field reachable from v - a pointer - to a value:
// pointers allocated, lists of one, strings "x", numbers 1. It is what a GET
// could answer at its fullest, to push through the PUT's type.
func fill(v reflect.Value, depth int) {
	if depth > 8 { //nolint:mnd
		return
	}
	switch v.Kind() { //nolint:exhaustive // the kinds the API's types are made of
	case reflect.Pointer:
		if v.IsNil() && v.CanSet() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		if !v.IsNil() {
			fill(v.Elem(), depth+1)
		}
	case reflect.Struct:
		if v.Type() == reflect.TypeFor[time.Time]() {
			v.Set(reflect.ValueOf(time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC)))
			return
		}
		for i := range v.NumField() {
			if field := v.Field(i); field.CanSet() {
				fill(field, depth+1)
			}
		}
	case reflect.Slice:
		slice := reflect.MakeSlice(v.Type(), 1, 1)
		fill(slice.Index(0), depth+1)
		v.Set(slice)
	case reflect.Map:
		if v.Type().Key().Kind() == reflect.String {
			m := reflect.MakeMap(v.Type())
			elem := reflect.New(v.Type().Elem()).Elem()
			fill(elem, depth+1)
			m.SetMapIndex(reflect.ValueOf("x").Convert(v.Type().Key()), elem)
			v.Set(m)
		}
	case reflect.String:
		v.SetString("x")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(1)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(1)
	case reflect.Float32, reflect.Float64:
		v.SetFloat(1)
	case reflect.Bool:
		v.SetBool(true)
	default:
	}
}

// What each kind's GET answers must go through its PUT's request, or the plan
// of a change to it fails for every app - a shape the two endpoints disagree
// on is found here, from the API's own types, rather than by a person.
func TestEverySettingsKindsGetFitsItsUpdate(t *testing.T) {
	for _, kind := range appSettingsKinds {
		get := kind.newGet()
		fill(reflect.ValueOf(get), 0)
		raw, err := json.Marshal(get)
		if !assert.NoError(t, err, kind.name) {
			continue
		}
		answered, err := through(raw, kind.newGet())
		if !assert.NoError(t, err, "%s: the GET's own type reads it", kind.name) {
			continue
		}
		answeredRaw, _ := json.Marshal(answered)
		body, err := through(answeredRaw, kind.newUpdate())
		if assert.NoError(t, err, "%s: the PUT's type takes what the GET answers", kind.name) {
			assert.Contains(t, body, updateVerField, "%s: the plan's updateVer reaches the PUT", kind.name)
		}
	}
}

func TestSettingsKindsAreNamedOnce(t *testing.T) {
	seen := map[string]bool{}
	for _, kind := range appSettingsKinds {
		assert.False(t, seen[kind.name], kind.name)
		seen[kind.name] = true
		assert.Contains(t, settingsKindNames(), kind.name)
	}
	_, err := findSettingsKind("secrets")
	assert.ErrorContains(t, err, "the kinds are env-vars")
}
