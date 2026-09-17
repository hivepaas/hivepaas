package templaterender

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergePatch(t *testing.T) {
	target := map[string]any{
		"a":    1,
		"b":    map[string]any{"c": 2, "d": 3},
		"list": []any{1, 2},
	}
	patch := map[string]any{
		"b":    map[string]any{"c": nil, "e": 4},
		"list": []any{9},
		"f":    "new",
	}

	got := MergePatch(target, patch)

	assert.Equal(t, map[string]any{
		"a":    1,
		"b":    map[string]any{"d": 3, "e": 4},
		"list": []any{9},
		"f":    "new",
	}, got, "maps merge, null deletes, lists replace")
	assert.Equal(t, map[string]any{"c": 2, "d": 3}, target["b"], "the target is not modified")
}

func TestMergePatchNonMaps(t *testing.T) {
	assert.Equal(t, map[string]any{"a": 1}, MergePatch("scalar", map[string]any{"a": 1}))
	assert.Equal(t, "scalar", MergePatch(map[string]any{"a": 1}, "scalar"))
}

func TestMergePatchCopiesWhatItKeeps(t *testing.T) {
	inner := map[string]any{"x": 1}
	got, ok := MergePatch(map[string]any{"inner": inner}, map[string]any{"other": 2}).(map[string]any)
	assert.True(t, ok)
	got["inner"].(map[string]any)["x"] = 99
	assert.Equal(t, 1, inner["x"], "the result shares nothing with its inputs")
}
