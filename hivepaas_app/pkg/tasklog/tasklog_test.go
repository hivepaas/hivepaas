package tasklog

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseTimestampFromData(t *testing.T) {
	// RFC3339 format without brackets
	tsStr := "2023-01-02T15:04:05Z"
	frame := &LogFrame{Data: tsStr + " payload"}
	frame.ParseTimestampFromData()
	assert.Equal(t, "payload", frame.Data)
	expected, _ := time.Parse(time.RFC3339, tsStr)
	assert.WithinDuration(t, expected, frame.Ts, time.Second)

	// Bracketed format
	frame2 := &LogFrame{Data: "[" + tsStr + "] other"}
	frame2.ParseTimestampFromData()
	assert.Equal(t, "other", frame2.Data)
	assert.WithinDuration(t, expected, frame2.Ts, time.Second)

	// Invalid timestamp should leave Data unchanged
	frame3 := &LogFrame{Data: "not-a-time data"}
	frame3.ParseTimestampFromData()
	assert.Equal(t, "not-a-time data", frame3.Data)
}

func TestNewFramesWithTimestamp(t *testing.T) {
	custom := time.Date(2022, 12, 31, 23, 59, 0, 0, time.UTC)

	in := NewInFrame("data", &custom)
	assert.Equal(t, LogTypeIn, in.Type)
	assert.Equal(t, "data", in.Data)
	assert.Equal(t, custom, in.Ts)

	out := NewOutFrame("out", &custom)
	assert.Equal(t, LogTypeOut, out.Type)
	assert.Equal(t, "out", out.Data)
	assert.Equal(t, custom, out.Ts)

	err := NewErrFrame("error", &custom)
	assert.Equal(t, LogTypeErr, err.Type)
	assert.Equal(t, "error", err.Data)
	assert.Equal(t, custom, err.Ts)

	warn := NewWarnFrame("warn", &custom)
	assert.Equal(t, LogTypeWarn, warn.Type)
	assert.Equal(t, "warn", warn.Data)
	assert.Equal(t, custom, warn.Ts)
}

func TestNewFramesWithNow(t *testing.T) {
	// Use TsNow to set timestamp to now
	in := NewInFrame("nowdata", TsNow)
	assert.Equal(t, LogTypeIn, in.Type)
	assert.Equal(t, "nowdata", in.Data)
	// Timestamp should be close to now
	assert.WithinDuration(t, time.Now().Add(-2*time.Second), in.Ts, 2*time.Second)
}

func TestMessageParsing(t *testing.T) {
	cmd, data := parseMessage(string(CommandNewData) + "\n" + "payload")
	assert.Equal(t, CommandNewData, cmd)
	assert.Equal(t, "payload", data)

	cmd2, data2 := parseMessage(string(CommandClosed) + "\n" + "closed")
	assert.Equal(t, CommandClosed, cmd2)
	assert.Equal(t, "closed", data2)
}

func TestBuildMessage(t *testing.T) {
	msg := buildMessage(CommandNewData)
	assert.Equal(t, string(CommandNewData), msg)
}
