package apptemplateserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/services/registry"
)

// fakeTagLister stands in for the registry client, and counts calls so the cache
// can be shown to work.
type fakeTagLister struct {
	tags  []string
	calls int
	err   error
}

func (f *fakeTagLister) ListTags(_ context.Context, _ registry.Reference, _ int) (*registry.ListTagsResult, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return &registry.ListTagsResult{Tags: f.tags}, nil
}

func newImageTagsTest(t *testing.T, tags ...string) (*service, *fakeTagLister) {
	t.Helper()
	svc, _ := newServiceTest(t, config.EnvDev, testRepoDir)
	lister := &fakeTagLister{tags: tags}
	svc.registryClient = lister
	svc.tagCache = newImageTagsCache()
	return svc, lister
}

func TestImageTagsOffersNewerBuildsOfTheSameImage(t *testing.T) {
	// The test repository's demo template pins demo:2.1.0 for version "2".
	svc, lister := newImageTagsTest(t, "2.1.0", "2.2.0", "1.9.3", "latest", "2")

	resp, err := svc.ImageTags(context.Background(), &apptemplateservice.ImageTagsReq{Name: "demo"})

	assert.NoError(t, err)
	assert.Equal(t, "registry-1.docker.io/library/demo", resp.Repository)
	assert.Equal(t, "2.1.0", resp.CurrentTag)
	assert.Equal(t, []*apptemplateservice.ImageTag{
		{Tag: "2.2.0", Class: templatemodel.ImageOverrideSameLine, Newer: true},
		{Tag: "1.9.3", Class: templatemodel.ImageOverrideOtherMajor, Newer: false},
	}, resp.Tags)
	assert.Equal(t, 1, lister.calls)
}

func TestImageTagsReadsTheRegistryOncePerRepository(t *testing.T) {
	svc, lister := newImageTagsTest(t, "2.2.0")

	for range 3 {
		_, err := svc.ImageTags(context.Background(), &apptemplateservice.ImageTagsReq{Name: "demo"})
		assert.NoError(t, err)
	}

	assert.Equal(t, 1, lister.calls, "a repository is read once per cache window, not once per click")
}

func TestImageTagsFollowsTheChosenVersion(t *testing.T) {
	svc, _ := newImageTagsTest(t, "1.9.4")

	resp, err := svc.ImageTags(context.Background(),
		&apptemplateservice.ImageTagsReq{Name: "demo", Version: "1"})

	assert.NoError(t, err)
	assert.Equal(t, "1.9.3", resp.CurrentTag, "the deprecated version is still a version to compare against")
	assert.Equal(t, "1.9.4", resp.Tags[0].Tag)
}

func TestImageTagsRefusesAVersionTheTemplateDoesNotHave(t *testing.T) {
	svc, _ := newImageTagsTest(t, "2.2.0")

	_, err := svc.ImageTags(context.Background(),
		&apptemplateservice.ImageTagsReq{Name: "demo", Version: "99"})

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateVersionNotFound)
}

func TestImageTagsPassesTheRegistryErrorThrough(t *testing.T) {
	svc, lister := newImageTagsTest(t)
	lister.err = hperrors.Wrap(hperrors.ErrRegistryRateLimited)

	_, err := svc.ImageTags(context.Background(), &apptemplateservice.ImageTagsReq{Name: "demo"})

	assert.ErrorIs(t, err, hperrors.ErrRegistryRateLimited)
	assert.Equal(t, 1, lister.calls, "a failed read is not cached")
}
