package fileutil

import "github.com/hivepaas/hivepaas/hivepaas_app/pkg/tracerr"

// MarshalYAML writes a file mode as octal - "0644" - the way MarshalJSON does,
// rather than as the decimal integer yaml.v3 would otherwise produce.
func (fm FileMode) MarshalYAML() (any, error) {
	return fm.String(), nil
}

// UnmarshalYAML accepts the octal string MarshalYAML writes.
func (fm *FileMode) UnmarshalYAML(unmarshal func(any) error) error {
	var text string
	if err := unmarshal(&text); err != nil {
		return tracerr.Wrap(err)
	}
	parsed, err := ParseFileMode(text)
	if err != nil {
		return tracerr.Wrap(err)
	}
	*fm = parsed
	return nil
}
