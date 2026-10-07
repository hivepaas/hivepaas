package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// An error's text can carry what a process printed, NUL bytes and all. bun 1.3
// refuses to send one, so the error would not be recorded at all - and the
// recording's own failure is ignored, by design.
func TestSysErrorInsertMultiReplacesNULBytes(t *testing.T) {
	repo := &sysErrorRepo{}
	db := bun.NewDB(nil, pgdialect.New())
	sysErrors := []*entity.SysError{{
		ID:         "E1",
		RequestID:  "R\x001",
		Code:       "ERR\x00",
		Detail:     "exec failed: sleep\x00infinity",
		Cause:      "exit 1\x00",
		DebugLog:   "out\x00put",
		StackTrace: "main.go:1\x00",
	}}

	sql := repo.insertMultiQuery(db, sysErrors).String()

	assert.NotContains(t, sql, "?!(", "what bun writes in place of a value it refuses")
	assert.NotContains(t, sql, "\x00")
	assert.Contains(t, sql, "'exec failed: sleep�infinity'")
	assert.Contains(t, sql, "'out�put'")
}
