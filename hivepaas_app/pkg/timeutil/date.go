package timeutil

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tracerr"
)

type Date time.Time

const (
	RFC3339DateIn  = "2006-1-2"   // In format, supports both yyyy-mm-dd and yyyy-m-d
	RFC3339DateOut = "2006-01-02" // Out format, supports yyyy-mm-dd
)

func NewDate(t time.Time) Date {
	return Date(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC))
}

func ParseDate(s string) (Date, error) {
	t, err := time.ParseInLocation(RFC3339DateIn, s, time.UTC)
	if err != nil {
		return Date{}, tracerr.Wrap(err)
	}
	return Date(t), nil
}

func (date Date) ToTime() time.Time {
	return time.Time(date)
}

func (date Date) Equal(d Date) bool {
	return date.ToTime().Equal(d.ToTime())
}

func (date Date) Before(d Date) bool {
	return date.ToTime().Before(d.ToTime())
}

func (date Date) After(d Date) bool {
	return date.ToTime().After(d.ToTime())
}

func (date Date) IsZero() bool {
	return time.Time(date).IsZero()
}

func (date Date) AddDate(years, months, days int) Date {
	return NewDate(date.ToTime().AddDate(years, months, days))
}

func (date Date) Sub(dt Date) time.Duration {
	return date.ToTime().Sub(dt.ToTime())
}

// String implement fmt.Stringer interface
func (date Date) String() string {
	return time.Time(date).Format(RFC3339DateOut)
}

func (date Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + date.String() + `"`), nil
}

func (date *Date) UnmarshalJSON(in []byte) error {
	if bytes.Equal(in, []byte("null")) {
		*date = Date{}
		return nil
	}
	// Remove double quotes covering the str
	if len(in) > 1 {
		in = in[1 : len(in)-1]
	}
	d, err := ParseDate(string(in))
	if err != nil {
		return err
	}
	*date = d
	return nil
}

var _ sql.Scanner = (*Date)(nil)
var _ driver.Valuer = (*Date)(nil)

// Scan scans the date or time value from DB table columns
func (date *Date) Scan(src any) (err error) {
	switch src := src.(type) {
	case time.Time:
		*date = Date(src)
		return nil
	case nil:
		*date = Date(time.Time{})
		return nil
	default:
		return fmt.Errorf("unsupported data type: %T", src) //nolint:err113
	}
}

// Value implements the [driver.Valuer] interface
func (date Date) Value() (driver.Value, error) {
	t := time.Time(date)
	if t.IsZero() {
		return nil, nil
	}
	return t, nil
}
