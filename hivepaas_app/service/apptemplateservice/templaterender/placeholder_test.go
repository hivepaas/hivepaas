package templaterender

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

var errTestUndefined = errors.New("undefined")

func mapResolver(values map[string]any) Resolver {
	return func(ref string) (any, bool, error) {
		value, ok := values[ref]
		if !ok {
			return nil, false, errTestUndefined
		}
		return value, false, nil
	}
}

func TestSubstitute(t *testing.T) {
	resolve := mapResolver(map[string]any{
		"params.memory":  unit.DataSize(512 << 20),
		"params.user":    "app",
		"params.port":    int64(5432),
		"params.enabled": true,
		"version.name":   "17",
	})
	tree := map[string]any{
		"memory":  "${{ params.memory }}",
		"port":    "${{params.port}}",
		"enabled": "${{ params.enabled }}",
		"cmd":     "pg_isready -U ${{ params.user }} -p ${{ params.port }}",
		"nested":  []any{map[string]any{"v": "${{ version.name }}"}},
		"escaped": "echo $${{ params.user }}",
		"envRef":  "${HIVEPAAS_PASSWORD}",
		"number":  3,
	}

	got, err := Substitute(tree, resolve)

	assert.NoError(t, err)
	assert.Equal(t, map[string]any{
		"memory":  unit.DataSize(512 << 20), // a whole-string placeholder takes the value's type
		"port":    int64(5432),
		"enabled": true,
		"cmd":     "pg_isready -U app -p 5432",
		"nested":  []any{map[string]any{"v": "17"}},
		"escaped": "echo ${{ params.user }}",
		"envRef":  "${HIVEPAAS_PASSWORD}", // HivePaaS's own references are not placeholders
		"number":  3,
	}, got)
	assert.Equal(t, "${{ params.memory }}", tree["memory"], "the input tree is not modified")
}

func TestSubstituteKeepsWhatTheResolverKeeps(t *testing.T) {
	resolve := func(ref string) (any, bool, error) {
		if ref == "params.password" {
			return nil, true, nil
		}
		return "app", false, nil
	}
	got, err := Substitute(map[string]any{
		"whole":    "${{params.password}}",
		"embedded": "postgres://${{ params.user }}:${{  params.password }}@db",
	}, resolve)

	assert.NoError(t, err)
	assert.Equal(t, map[string]any{
		"whole":    "${{ params.password }}",
		"embedded": "postgres://app:${{ params.password }}@db",
	}, got, "a kept placeholder is written in one canonical spelling")
}

func TestSubstituteReportsTheResolverError(t *testing.T) {
	_, err := Substitute([]any{"ok", "${{ params.nope }}"}, mapResolver(map[string]any{}))
	assert.ErrorIs(t, err, errTestUndefined)
}
