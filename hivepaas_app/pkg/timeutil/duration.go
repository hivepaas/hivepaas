package timeutil

import (
	"bytes"
	"strconv"
	"time"

	"github.com/xhit/go-str2duration/v2"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tracerr"
)

const (
	Day        = 24 * time.Hour
	Dur7Days   = 7 * Day
	Dur10Days  = 10 * Day
	Dur15Days  = 15 * Day
	Dur30Days  = 30 * Day
	Dur60Days  = 60 * Day
	Dur90Days  = 90 * Day
	Dur365Days = 365 * Day

	Week      = Dur7Days
	Dur2Weeks = 2 * Week
	Dur4Weeks = 4 * Week
)

type Duration time.Duration

func ParseDuration(s string) (Duration, error) {
	v, err := str2duration.ParseDuration(s)
	if err != nil {
		return 0, tracerr.Wrap(err)
	}
	return Duration(v), nil
}

//nolint:gosec
func (dur Duration) String() string {
	u := uint64(dur)
	if dur < 0 {
		u = -u
	}

	// Less than a day use the default format func
	if u < uint64(Day) {
		return str2duration.String(time.Duration(dur))
	}

	// Bigger than a day, display at `day` fraction (not use `week`)
	days := u / uint64(Day)
	u -= days * uint64(Day)
	res := strconv.FormatInt(int64(days), 10) + "d"
	if u > 0 {
		res += str2duration.String(time.Duration(u))
	}
	if dur < 0 {
		return "-" + res
	}
	return res
}

func (dur Duration) ToDuration() time.Duration {
	return time.Duration(dur)
}

func (dur Duration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + dur.String() + `"`), nil
}

func (dur *Duration) UnmarshalJSON(in []byte) error {
	if bytes.Equal(in, []byte("null")) {
		*dur = 0
		return nil
	}

	// Remove double quotes covering the str
	if len(in) > 1 && in[0] == '"' {
		in = in[1 : len(in)-1]
		d, err := ParseDuration(reflectutil.UnsafeBytesToStr(in))
		if err != nil {
			return tracerr.Wrap(err)
		}
		*dur = d
		return nil
	}

	// Parse duration as integer
	v, err := strconv.ParseInt(reflectutil.UnsafeBytesToStr(in), 10, 64)
	if err != nil {
		return tracerr.Wrap(err)
	}
	*dur = Duration(v)
	return nil
}
