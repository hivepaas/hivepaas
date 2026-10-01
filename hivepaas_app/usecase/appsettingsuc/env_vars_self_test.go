package appsettingsuc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice/envself"
)

// Which credentials are set is read from the kind settings; none of them is
// carried further.
func TestSelfFactsAreWhichCredentialsAreSet(t *testing.T) {
	facts := &envself.Facts{Set: map[string]bool{}}
	selfFacts(&entity.AppKindSettings{Category: base.AppCategoryDatabase, Engine: "mysql",
		Database: &entity.AppKindDatabase{DbName: "app", Username: "app",
			Password: entity.NewEncryptedField("s3cret")}}, facts)
	assert.Equal(t, base.AppCategoryDatabase, facts.Category)
	assert.Equal(t, "mysql", facts.Engine)
	assert.True(t, facts.Set[base.AppSystemEnvVarPassword])
	assert.False(t, facts.Set[base.AppSystemEnvVarRootPassword])

	facts = &envself.Facts{Set: map[string]bool{}}
	selfFacts(&entity.AppKindSettings{Category: base.AppCategoryCache, Engine: "redis",
		Cache: &entity.AppKindCache{}}, facts)
	assert.False(t, facts.Set[base.AppSystemEnvVarPassword])
}
