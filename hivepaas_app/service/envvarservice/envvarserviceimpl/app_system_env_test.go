package envvarserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func envByKey(t *testing.T, envs []*envvarservice.EnvVar) map[string]*entity.EnvVar {
	t.Helper()
	byKey := map[string]*entity.EnvVar{}
	for _, env := range envs {
		byKey[env.Key] = env.EnvVar
	}
	return byKey
}

func TestKindEnvVars_Database(t *testing.T) {
	kind := &entity.AppKindSettings{
		Category: base.AppCategoryDatabase,
		Engine:   "postgres",
		Database: &entity.AppKindDatabase{
			DbName:       "app",
			Username:     "app",
			Password:     entity.NewEncryptedField("s3cret"),
			RootPassword: entity.NewEncryptedField("r00t"),
			SSLMode:      base.DatabaseSslModeDisable,
		},
	}

	envs, err := kindEnvVars(kind)
	assert.NoError(t, err)

	byKey := envByKey(t, envs)
	assert.Equal(t, "app", byKey[base.AppSystemEnvVarUser].Value)
	assert.Equal(t, "s3cret", byKey[base.AppSystemEnvVarPassword].Value)
	assert.Equal(t, "app", byKey[base.AppSystemEnvVarDatabaseName].Value)
	assert.True(t, byKey[base.AppSystemEnvVarPassword].IsShared)
	assert.False(t, byKey[base.AppSystemEnvVarRootPassword].IsShared,
		"the root password stays with the app that owns it")
}

func TestKindEnvVars_Cache(t *testing.T) {
	kind := &entity.AppKindSettings{
		Category: base.AppCategoryCache,
		Engine:   "redis",
		Cache: &entity.AppKindCache{
			Password: entity.NewEncryptedField("s3cret"),
		},
	}

	envs, err := kindEnvVars(kind)
	assert.NoError(t, err)

	byKey := envByKey(t, envs)
	assert.Equal(t, "s3cret", byKey[base.AppSystemEnvVarPassword].Value)
	assert.True(t, byKey[base.AppSystemEnvVarPassword].IsShared,
		"other apps connect to the cache with this password")

	// A cache names no user, selects no database and has no root account.
	for _, key := range []string{base.AppSystemEnvVarUser, base.AppSystemEnvVarDatabaseName,
		base.AppSystemEnvVarRootPassword, base.AppSystemEnvVarSSLMode} {
		assert.NotContains(t, byKey, key)
	}
}

func TestKindEnvVars_CacheWithoutPassword(t *testing.T) {
	kind := &entity.AppKindSettings{
		Category: base.AppCategoryCache,
		Engine:   "memcached",
		Cache:    &entity.AppKindCache{},
	}

	envs, err := kindEnvVars(kind)
	assert.NoError(t, err)

	byKey := envByKey(t, envs)
	assert.Contains(t, byKey, base.AppSystemEnvVarPassword,
		"a cache without authentication still publishes the variable, empty")
	assert.Equal(t, "", byKey[base.AppSystemEnvVarPassword].Value)
}

func TestKindEnvVars_NothingToPublish(t *testing.T) {
	tests := map[string]*entity.AppKindSettings{
		"no kind setting":            nil,
		"webapp":                     {Category: base.AppCategoryWebapp, Webapp: &entity.AppKindWebapp{}},
		"database without its block": {Category: base.AppCategoryDatabase},
		"cache without its block":    {Category: base.AppCategoryCache},
	}

	for name, kind := range tests {
		t.Run(name, func(t *testing.T) {
			envs, err := kindEnvVars(kind)
			assert.NoError(t, err)
			assert.Empty(t, envs)
		})
	}
}
