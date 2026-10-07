package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// A tag can come from outside: a kopia snapshot's tags, written by whatever
// wrote the snapshot, are recorded in the transaction that syncs the snapshots.
// One NUL would fail it all.
func TestTagUpsertMultiReplacesNULBytes(t *testing.T) {
	repo := &tagRepo{}
	db := bun.NewDB(nil, pgdialect.New())
	tags := []*entity.Tag{{ObjectID: "S1", Tag: "app=web"}, {ObjectID: "S1", Tag: "by\x00kopia"}}

	sql := repo.upsertMultiQuery(db, tags, nil, nil).String()

	assert.NotContains(t, sql, "?!(", "what bun writes in place of a value it refuses")
	assert.Contains(t, sql, "'by�kopia'")
}
