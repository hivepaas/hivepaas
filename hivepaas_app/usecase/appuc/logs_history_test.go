package appuc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/services/logging"
)

func TestHistoryFrameType(t *testing.T) {
	cases := []struct {
		entry logging.LogEntry
		want  tasklog.LogType
	}{
		{logging.LogEntry{Stream: "stdout"}, tasklog.LogTypeOut},
		{logging.LogEntry{Stream: "stderr"}, tasklog.LogTypeErr},
		{logging.LogEntry{Stream: "stdout", Level: "ERROR"}, tasklog.LogTypeErr},
		{logging.LogEntry{Stream: "stdout", Level: "fatal"}, tasklog.LogTypeErr},
		{logging.LogEntry{Stream: "stderr", Level: "warning"}, tasklog.LogTypeWarn},
		{logging.LogEntry{Stream: "stderr", Level: "debug"}, tasklog.LogTypeDebug},
		{logging.LogEntry{Stream: "stderr", Level: "info"}, tasklog.LogTypeOut},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, historyFrameType(&tc.entry), "%+v", tc.entry)
	}
}

func TestHistoryPage(t *testing.T) {
	t1 := time.Date(2026, 9, 12, 8, 0, 1, 0, time.UTC)
	t2 := t1.Add(time.Second)
	data := toHistoryData(&logging.QueryResp{
		Entries:   []logging.LogEntry{{Time: t1, Message: "a"}, {Time: t2, Message: "b"}},
		Truncated: true,
	})
	if assert.Len(t, data.Logs, 2) {
		assert.Equal(t, "a", data.Logs[0].Data)
		assert.Equal(t, t1, data.Logs[0].Ts)
	}
	if assert.NotNil(t, data.NextEnd) {
		assert.Equal(t, t1.Add(-time.Nanosecond), *data.NextEnd, "strictly before the oldest line shown")
	}

	data = toHistoryData(&logging.QueryResp{Entries: []logging.LogEntry{{Time: t1}}})
	assert.Nil(t, data.NextEnd)
	data = toHistoryData(&logging.QueryResp{})
	assert.NotNil(t, data.Logs, "an empty page is [], not null")
}
