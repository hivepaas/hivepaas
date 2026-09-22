package entity_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func appIn(projectKey, envKey, appKey string) *entity.App {
	return &entity.App{
		Key:        appKey,
		GlobalKey:  projectKey + "/" + envKey + "/" + appKey,
		Project:    &entity.Project{Key: projectKey},
		ProjectEnv: &entity.ProjectEnv{Key: envKey},
	}
}

func TestImageRepoNameJoinsProjectAndApp(t *testing.T) {
	name, err := appIn("shop", "dev", "api").ImageRepoName()

	assert.NoError(t, err)
	assert.Equal(t, "shop-api", name)
}

// The environment is in the tag, so one app is one repository however many
// environments it runs in.
func TestImageRepoNameIsTheSameInEveryEnvironment(t *testing.T) {
	dev, err := appIn("shop", "dev", "api").ImageRepoName()
	assert.NoError(t, err)
	prod, err := appIn("shop", "prod", "api").ImageRepoName()
	assert.NoError(t, err)

	assert.Equal(t, dev, prod)
}

// The same app key in two projects was one repository before this change, which
// mixed two projects' images and let a busy one evict a quiet one's.
func TestImageRepoNameSeparatesProjects(t *testing.T) {
	first, _ := appIn("shop", "dev", "api").ImageRepoName()
	second, _ := appIn("blog", "dev", "api").ImageRepoName()

	assert.NotEqual(t, first, second)
}

func TestImageRepoNameNeedsItsProject(t *testing.T) {
	app := &entity.App{Key: "api", ProjectEnv: &entity.ProjectEnv{Key: "dev"}}

	_, err := app.ImageRepoName()
	assert.Error(t, err)
}

// A project and an app may each be named with 100 characters, and the address
// and the account are still to go in front of the result.
func TestImageRepoNameTruncatesAndStaysUnique(t *testing.T) {
	long := strings.Repeat("a", 80)
	first, err := appIn(long, "dev", "one-"+strings.Repeat("b", 80)).ImageRepoName()
	assert.NoError(t, err)
	second, err := appIn(long, "dev", "one-"+strings.Repeat("c", 80)).ImageRepoName()
	assert.NoError(t, err)

	assert.LessOrEqual(t, len(first), 110)
	assert.NotEqual(t, first, second, "two cut names must not collide")
}

func TestImageRepoNameIsValidOCI(t *testing.T) {
	name, err := appIn("shop__one", "dev", "api--two").ImageRepoName()

	assert.NoError(t, err)
	assert.Equal(t, "shop_one-api-two", name)
	assert.NotContains(t, name, "__")
	assert.NotContains(t, name, "--")
	assert.False(t, strings.HasPrefix(name, "-"))
	assert.False(t, strings.HasSuffix(name, "-"))
}

func TestImageTagCarriesTheEnvironment(t *testing.T) {
	tag, err := appIn("shop", "dev", "api").ImageTag("9f3c1de0ab")

	assert.NoError(t, err)
	assert.Equal(t, "dev-9f3c1de", tag)
}

// Production is prefixed like everything else: nothing marks an environment as
// production, so a bare tag would be a convention only some installations get.
func TestImageTagPrefixesProductionToo(t *testing.T) {
	tag, err := appIn("shop", "prod", "api").ImageTag("9f3c1de0ab")

	assert.NoError(t, err)
	assert.Equal(t, "prod-9f3c1de", tag)
}

func TestImageTagPrefixIsCutToTwentyCharacters(t *testing.T) {
	prefix, err := appIn("shop", strings.Repeat("e", 40), "api").ImageTagPrefix()

	assert.NoError(t, err)
	assert.Len(t, prefix, 20)
}

func TestImageTagRefusesAShortCommit(t *testing.T) {
	_, err := appIn("shop", "dev", "api").ImageTag("9f3")
	assert.Error(t, err)
}
