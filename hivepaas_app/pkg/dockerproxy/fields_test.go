package dockerproxy

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsZeroIsWhatAClientSendsForUnset(t *testing.T) {
	unset := []any{
		nil, false, "", int64(0), json.Number("0"), json.Number("0.0"), []any{}, map[string]any{},
		map[string]any{"Type": "", "Config": map[string]any{}},
	}
	for _, v := range unset {
		assert.True(t, isZero(v), "%#v", v)
	}
	set := []any{
		true, "host", int64(3), json.Number("1"), json.Number("-1"), []any{"SYS_ADMIN"},
		[]any{json.Number("0")}, map[string]any{"Type": "syslog"},
	}
	for _, v := range set {
		assert.False(t, isZero(v), "%#v", v)
	}
}

func TestCheckFieldsRefusesAFieldItDoesNotKnow(t *testing.T) {
	obj := map[string]any{"Image": "alpine", "Privileged": false, "MaskedPaths": nil}
	stop(t, assert.NoError(t, checkFields("HostConfig", obj, []string{"Image"}, []string{"MaskedPaths"})))

	obj["Privileged"] = true
	err := checkFields("HostConfig", obj, []string{"Image"}, []string{"MaskedPaths"})
	var refusal *refusalError
	stop(t, assert.ErrorAs(t, err, &refusal))
	assert.EqualError(t, err, "HostConfig.Privileged is not allowed")
}

func TestCheckFieldsTakesOnlyNullWhereAnEmptyListMeansSomething(t *testing.T) {
	// An empty MaskedPaths is how --security-opt systempaths=unconfined unmasks /proc.
	err := checkFields("HostConfig", map[string]any{"MaskedPaths": []any{}}, nil, []string{"MaskedPaths"})
	assert.EqualError(t, err, "HostConfig.MaskedPaths must not be set")
}

func TestCheckFieldsKnowsTheCLIsUnsetSwappiness(t *testing.T) {
	assert.NoError(t, checkFields("HostConfig", map[string]any{"MemorySwappiness": json.Number("-1")}, nil, nil))
	assert.EqualError(t,
		checkFields("HostConfig", map[string]any{"MemorySwappiness": json.Number("60")}, nil, nil),
		"HostConfig.MemorySwappiness is not allowed")
}
