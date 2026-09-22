package imagebuildserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func buildApp() *entity.App {
	return &entity.App{
		ID: "app-1", Key: "api", GlobalKey: "shop/dev/api",
		Project:    &entity.Project{Key: "shop"},
		ProjectEnv: &entity.ProjectEnv{Key: "dev"},
	}
}

func registryAuth() *entity.RegistryAuth {
	return &entity.RegistryAuth{Address: "registry.example.com", Username: "hivepaas"}
}

func TestReferencesWithARegistry(t *testing.T) {
	refs, err := buildImageReferences(buildApp(), "9f3c1de0ab", nil, registryAuth())

	assert.NoError(t, err)
	assert.Equal(t, []string{
		"registry.example.com/hivepaas/shop-api:dev-9f3c1de",
		"shop-api:dev-9f3c1de",
	}, refs)
}

// The service spec runs ImageTags[0], so the commit reference has to stay first
// however many tags the deployment added.
func TestReferencesPutTheCommitFirst(t *testing.T) {
	refs, err := buildImageReferences(buildApp(), "9f3c1de0ab", []string{"v1.4.0", "stable"},
		registryAuth())

	assert.NoError(t, err)
	assert.Equal(t, "registry.example.com/hivepaas/shop-api:dev-9f3c1de", refs[0])
	assert.Contains(t, refs, "registry.example.com/hivepaas/shop-api:dev-v1.4.0")
	assert.Contains(t, refs, "registry.example.com/hivepaas/shop-api:dev-stable")
}

// Two environments share one repository, so a custom tag of one must not
// overwrite the other's.
func TestReferencesPrefixCustomTagsWithTheEnvironment(t *testing.T) {
	app := buildApp()
	app.ProjectEnv = &entity.ProjectEnv{Key: "prod"}

	refs, err := buildImageReferences(app, "9f3c1de0ab", []string{"v1.4.0"}, registryAuth())

	assert.NoError(t, err)
	assert.Contains(t, refs, "registry.example.com/hivepaas/shop-api:prod-v1.4.0")
	assert.NotContains(t, refs, "registry.example.com/hivepaas/shop-api:v1.4.0")
}

// Without a registry nothing can be pushed, so only the local reference is built
// - which is what a single-node installation runs.
func TestReferencesWithoutARegistry(t *testing.T) {
	refs, err := buildImageReferences(buildApp(), "9f3c1de0ab", []string{"v1.4.0"}, nil)

	assert.NoError(t, err)
	assert.Equal(t, []string{"shop-api:dev-9f3c1de", "shop-api:dev-v1.4.0"}, refs)
}

func TestReferencesNeedAnApp(t *testing.T) {
	_, err := buildImageReferences(&entity.App{Key: "api"}, "9f3c1de0ab", nil, registryAuth())
	assert.Error(t, err)
}
