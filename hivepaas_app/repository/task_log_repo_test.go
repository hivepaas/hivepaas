package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// bun used to drop a NUL on its way to the database. Since 1.3 it writes an error
// in place of the value, which fails the statement - and the task's transaction
// with it - so a process that printed one would lose the task's ending too.
func TestInsertMultiReplacesNULBytesInTheData(t *testing.T) {
	repo := &taskLogRepo{}
	db := bun.NewDB(nil, pgdialect.New())
	logs := []*entity.TaskLog{
		{TaskID: "T1", Data: "Step 1/3 : FROM alpine"},
		{TaskID: "T1", Data: "sleep\x00infinity\x00"},
	}

	sql := repo.insertMultiQuery(db, logs).String()

	assert.NotContains(t, sql, "?!(", "what bun writes in place of a value it refuses")
	assert.Contains(t, sql, "'Step 1/3 : FROM alpine'")
	assert.Contains(t, sql, "'sleep\uFFFDinfinity\uFFFD'")
}
