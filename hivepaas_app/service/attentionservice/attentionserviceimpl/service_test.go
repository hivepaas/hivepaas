package attentionserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

// The env's name is read with the apps: an item names it, and links the app's
// screens with it. Read without it, an item had no env, and no link.
func TestTheAppsAreReadWithTheirEnvsName(t *testing.T) {
	var apps []*entity.App
	query := bunex.ApplySelect(bun.NewDB(nil, pgdialect.New()).NewSelect().Model(&apps), appsRead()...).String()

	assert.Contains(t, query, `"project_env"."name"`)
}
