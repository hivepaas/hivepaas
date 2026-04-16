package osutil

import (
	"bytes"
	"os"
	"strconv"
	"strings"

	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tracerr"
)

type FileMode os.FileMode

func ParseFileMode(s string) (FileMode, error) {
	// Parse duration as octal integer
	s = strings.TrimPrefix(s, "0")
	if s == "" {
		return 0, nil
	}
	v, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return 0, tracerr.Wrap(err)
	}
	return FileMode(v), nil //nolint:gosec
}

func (fm FileMode) String() string {
	return "0" + strconv.FormatUint(uint64(fm), 8)
}

func (fm FileMode) ToFileMode() os.FileMode {
	return os.FileMode(fm)
}

func (fm FileMode) MarshalJSON() ([]byte, error) {
	return []byte(`"` + fm.String() + `"`), nil
}

func (fm *FileMode) UnmarshalJSON(in []byte) error {
	if bytes.Equal(in, []byte("null")) {
		*fm = 0
		return nil
	}

	// Remove double quotes covering the str if there are
	if len(in) > 1 && in[0] == '"' {
		in = in[1 : len(in)-1]
	}

	d, err := ParseFileMode(reflectutil.UnsafeBytesToStr(in))
	if err != nil {
		return tracerr.Wrap(err)
	}
	*fm = d
	return nil
}
