package appsettingsdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// A request carries the block of the category it is for and no others, so every
// one of these is nil on some ordinary save. Reading a field off the nil one
// took the whole request down with it.
func TestKindSettingsToEntityWithOnlyItsOwnBlock(t *testing.T) {
	req := &AppKindSettingsReq{
		Category: base.AppCategoryCache,
		Engine:   "redis",
		Cache: &AppKindCacheReq{
			MaxMemory:       128 * 1024 * 1024,
			EvictionRule:    "noeviction",
			PersistenceMode: "rdb",
		},
	}

	kind := req.ToEntity()

	assert.NotNil(t, kind.Cache)
	assert.Equal(t, "noeviction", kind.Cache.EvictionRule)
	assert.Nil(t, kind.Database, "a cache app has no database block, and that is not a crash")
	assert.Nil(t, kind.Webapp)
	assert.Nil(t, kind.Storage)
}

func TestKindSettingsToEntityForADatabase(t *testing.T) {
	req := &AppKindSettingsReq{
		Category: base.AppCategoryDatabase,
		Engine:   "postgres",
		Database: &AppKindDatabaseReq{DbName: "app", Username: "app"},
	}

	kind := req.ToEntity()

	assert.NotNil(t, kind.Database)
	assert.Equal(t, "app", kind.Database.DbName)
	assert.Nil(t, kind.Cache)
}
