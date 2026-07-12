package version

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

var (
	ErrInvalidFormat = errors.New("invalid version format")

	versionRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z]+)(\d*))?$`)
)

type Version struct {
	Major        int
	Minor        int
	Patch        int
	Suffix       string
	SuffixNumber int
}

func Parse(s string) (*Version, error) {
	matches := versionRegex.FindStringSubmatch(s)
	if matches == nil {
		return nil, ErrInvalidFormat
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	v := &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}

	v.Suffix = matches[4]
	if matches[5] != "" {
		suffixNum, _ := strconv.Atoi(matches[5])
		v.SuffixNumber = suffixNum
	}

	return v, nil
}

func Cmp(v1, v2 *Version) int {
	if v1.Major != v2.Major {
		return v1.Major - v2.Major
	}
	if v1.Minor != v2.Minor {
		return v1.Minor - v2.Minor
	}
	if v1.Patch != v2.Patch {
		return v1.Patch - v2.Patch
	}
	if v1.Suffix != v2.Suffix {
		return strings.Compare(v1.Suffix, v2.Suffix)
	}
	if v1.SuffixNumber != v2.SuffixNumber {
		return v1.SuffixNumber - v2.SuffixNumber
	}
	return 0
}

func CmpStr(s1, s2 string) (int, error) {
	v1, err := Parse(s1)
	if err != nil {
		return 0, apperrors.Wrap(err)
	}
	v2, err := Parse(s2)
	if err != nil {
		return 0, apperrors.Wrap(err)
	}
	return Cmp(v1, v2), nil
}
