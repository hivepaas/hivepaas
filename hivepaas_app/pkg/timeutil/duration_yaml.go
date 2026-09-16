package timeutil

import (
	"strconv"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tracerr"
)

// MarshalYAML writes a duration the way MarshalJSON does - "30s", "1d2h" -
// rather than as a count of nanoseconds.
//
// yaml.v3 does not consult MarshalJSON, and it honors encoding.TextMarshaler
// rather than defining its own text hook. Implementing TextMarshaler here would
// also change how every other text-based encoder treats this type, including
// TOML, so the YAML hook is implemented on its own instead.
func (dur Duration) MarshalYAML() (any, error) {
	return dur.String(), nil
}

// UnmarshalYAML accepts what MarshalYAML writes, and also a bare number of
// nanoseconds so that a hand-edited document carrying an integer still loads.
//
// Both arrive as a string: YAML scalars are untyped, so yaml.v3 will decode
// `30s` and `5000000000` alike into a string. The number case is therefore a
// fallback on the parse, not on the decode.
func (dur *Duration) UnmarshalYAML(unmarshal func(any) error) error {
	var text string
	if err := unmarshal(&text); err != nil {
		return tracerr.Wrap(err)
	}

	parsed, err := ParseDurationWithEmptyIsZero(text)
	if err == nil {
		*dur = parsed
		return nil
	}

	nanos, numErr := strconv.ParseInt(text, 10, 64)
	if numErr != nil {
		return tracerr.Wrap(err) // report the duration error, which is the useful one
	}
	*dur = Duration(nanos)
	return nil
}
