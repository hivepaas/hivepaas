package fileutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type modeHolder struct {
	Mode FileMode `yaml:"mode"`
}

func TestFileModeMarshalsYAMLAsOctal(t *testing.T) {
	out, err := yaml.Marshal(&modeHolder{Mode: FileMode(0o644)})
	assert.NoError(t, err)
	assert.Equal(t, "mode: \"0644\"\n", string(out))
}

func TestFileModeYAMLRoundTrips(t *testing.T) {
	for _, in := range []FileMode{0, 0o600, 0o644, 0o755} {
		out, err := yaml.Marshal(&modeHolder{Mode: in})
		assert.NoError(t, err)

		var back modeHolder
		assert.NoError(t, yaml.Unmarshal(out, &back))
		assert.Equal(t, in, back.Mode, "round trip of %v via %q", in, out)
	}
}
