package timeutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type durHolder struct {
	Interval Duration `yaml:"interval"`
}

// Without a YAML hook this renders as 30000000000, which is unreadable in a
// document meant to be reviewed and hand-edited.
func TestDurationMarshalsYAMLAsText(t *testing.T) {
	out, err := yaml.Marshal(&durHolder{Interval: Duration(30_000_000_000)})
	assert.NoError(t, err)
	assert.Equal(t, "interval: 30s\n", string(out))
}

func TestDurationYAMLRoundTrips(t *testing.T) {
	cases := []Duration{
		0,
		Duration(1500 * 1000 * 1000),
		Duration(Dur2Weeks),
		Duration(-3600 * 1000 * 1000 * 1000),
	}
	for _, in := range cases {
		out, err := yaml.Marshal(&durHolder{Interval: in})
		assert.NoError(t, err)

		var back durHolder
		assert.NoError(t, yaml.Unmarshal(out, &back))
		assert.Equal(t, in, back.Interval, "round trip of %v via %q", in, out)
	}
}

// A hand-edited document with a bare integer still loads.
func TestDurationYAMLAcceptsRawNanoseconds(t *testing.T) {
	var back durHolder
	assert.NoError(t, yaml.Unmarshal([]byte("interval: 5000000000\n"), &back))
	assert.Equal(t, Duration(5_000_000_000), back.Interval)
}

func TestDurationYAMLRejectsNonsense(t *testing.T) {
	var back durHolder
	assert.Error(t, yaml.Unmarshal([]byte("interval: not-a-duration\n"), &back))
}
