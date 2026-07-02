package tasklog

import (
	"strings"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

type LogType string

const (
	LogTypeIn    LogType = "in"
	LogTypeOut   LogType = "out"
	LogTypeErr   LogType = "err"
	LogTypeWarn  LogType = "warn"
	LogTypeDebug LogType = "debug"
)

var (
	TsNow   = (*time.Time)(nil)
	TsParse = &time.Time{}
)

type LogFrame struct {
	Type LogType   `json:"type"`
	Data string    `json:"data"`
	Ts   time.Time `json:"ts"`
}

// ParseTimestampFromData parses timestamp which may be prefixed to the data.
// Supported formats: `<RFC-3339> <data>` or `[<RFC-3339>] <data>`
func (f *LogFrame) ParseTimestampFromData() {
	tsPart, dataPart, found := strings.Cut(f.Data, " ")
	if !found {
		return
	}
	ts := parseTimestamp(tsPart)
	if ts == nil {
		return
	}
	f.Ts = *ts
	f.Data = dataPart
}

func (f *LogFrame) calcTimestamp(ts *time.Time) {
	if ts == TsNow {
		f.Ts = timeutil.NowUTC()
		return
	}
	if ts == TsParse {
		f.ParseTimestampFromData()
		return
	}
	f.Ts = *ts
}

func NewInFrame(data string, ts *time.Time) *LogFrame {
	f := &LogFrame{
		Type: LogTypeIn,
		Data: data,
	}
	f.calcTimestamp(ts)
	return f
}

func NewOutFrame(data string, ts *time.Time) *LogFrame {
	f := &LogFrame{
		Type: LogTypeOut,
		Data: data,
	}
	f.calcTimestamp(ts)
	return f
}

func NewErrFrame(data string, ts *time.Time) *LogFrame {
	f := &LogFrame{
		Type: LogTypeErr,
		Data: data,
	}
	f.calcTimestamp(ts)
	return f
}

func NewWarnFrame(data string, ts *time.Time) *LogFrame {
	f := &LogFrame{
		Type: LogTypeWarn,
		Data: data,
	}
	f.calcTimestamp(ts)
	return f
}

func NewDebugFrame(data string, ts *time.Time) *LogFrame {
	f := &LogFrame{
		Type: LogTypeDebug,
		Data: data,
	}
	f.calcTimestamp(ts)
	return f
}

func parseTimestamp(s string) *time.Time {
	// Unwrap [] if there is
	s = gofn.StringUnwrapLR(s, "[", "]")

	// Try RFC3339Nano first
	ts, err := time.Parse(time.RFC3339Nano, s)
	if err == nil {
		return &ts
	}
	// Try RFC3339
	ts, err = time.Parse(time.RFC3339, s)
	if err == nil {
		return &ts
	}
	return nil
}
