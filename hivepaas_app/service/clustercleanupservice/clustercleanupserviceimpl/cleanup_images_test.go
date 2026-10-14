package clustercleanupserviceimpl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustercleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	traefikDigest = "sha256:24841fe2de7304c149343d877d2923b4c8800a38ba015dea9174c23b20e344a0"
	otherDigest   = "sha256:0404ca69b522f8629d7d4e9034a7afe0300b713354e8bf12ec9657581cf59400"
)

func TestImageKeeperKeepsTheImagesServicesName(t *testing.T) {
	keeper := newImageKeeper([]string{
		"traefik:v3.7.13@" + traefikDigest,
		"e2e_shop-web:dev-36d81dd",
		"ghcr.io/hivepaas/placeholder@" + otherDigest,
		"busybox",
		"not a reference",
	})

	cases := []struct {
		name string
		img  image.Summary
		kept bool
	}{
		{"pulled by the digest the service names, classic store",
			image.Summary{ID: "sha256:aaa", RepoDigests: []string{"traefik@" + traefikDigest}}, true},
		{"the digest is the ID, containerd store",
			image.Summary{ID: traefikDigest}, true},
		{"the tag the service names, the name normalized",
			image.Summary{ID: "sha256:bbb", RepoTags: []string{"docker.io/library/traefik:v3.7.13"}}, true},
		{"built on the node, never pushed",
			image.Summary{ID: "sha256:ccc", RepoTags: []string{"e2e_shop-web:dev-36d81dd"}}, true},
		{"named by digest only",
			image.Summary{ID: "sha256:ddd", RepoDigests: []string{"ghcr.io/hivepaas/placeholder@" + otherDigest}}, true},
		{"a reference without a tag is latest",
			image.Summary{ID: "sha256:eee", RepoTags: []string{"busybox:latest"}}, true},
		{"another tag of the same repository",
			image.Summary{ID: "sha256:fff", RepoTags: []string{"e2e_shop-web:dev-0000000"}}, false},
		{"another version",
			image.Summary{ID: "sha256:ggg", RepoTags: []string{"traefik:v3.7.12"}}, false},
		{"untagged",
			image.Summary{ID: "sha256:hhh", RepoTags: []string{"<none>:<none>"}, RepoDigests: []string{"<none>@<none>"}},
			false},
	}
	for _, c := range cases {
		assert.Equal(t, c.kept, keeper.keeps(&c.img), c.name)
	}
}

func TestPlanImagesCountsFromWhenAnImageWasFirstFoundUnused(t *testing.T) {
	now := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	retention := 3 * timeutil.Day
	images := []image.Summary{
		{ID: "sha256:running"},
		{ID: "sha256:kept", RepoTags: []string{"traefik:v3.7.13"}},
		{ID: "sha256:new"},
		{ID: "sha256:young"},
		{ID: "sha256:old"},
		{ID: "sha256:future"},
	}
	unusedSince := map[string]time.Time{
		// In use again: its count is dropped.
		"sha256:running": now.Add(-10 * timeutil.Day),
		"sha256:young":   now.Add(-2 * timeutil.Day),
		"sha256:old":     now.Add(-3 * timeutil.Day),
		// Set by a clock ahead of this one.
		"sha256:future": now.Add(timeutil.Day),
		// Gone from the node.
		"sha256:gone": now.Add(-timeutil.Day),
	}
	plan := planImages(images, map[string]bool{"sha256:running": true}, newImageKeeper([]string{"traefik:v3.7.13"}),
		unusedSince, now, retention)

	assert.Len(t, plan.remove, 1)
	assert.Equal(t, "sha256:old", plan.remove[0].ID)
	assert.Equal(t, map[string]time.Time{
		"sha256:new":    now,
		"sha256:young":  now.Add(-2 * timeutil.Day),
		"sha256:future": now,
	}, plan.unusedSince)
}

func TestPlanImagesWithoutRetentionRemovesAnUnusedImageAtOnce(t *testing.T) {
	now := time.Now()
	plan := planImages([]image.Summary{{ID: "sha256:new"}}, nil, newImageKeeper(nil), nil, now, 0)
	assert.Len(t, plan.remove, 1)
	assert.Empty(t, plan.unusedSince)
}

// imageNode is a node's Docker: its images, its containers, and what removing
// an image by a name answers.
type imageNode struct {
	docker.Manager
	images     []image.Summary
	containers []container.Summary
	services   []swarm.Service
	removeErr  map[string]error
	removed    []string
	pruned     bool
}

func (n *imageNode) NodeCurrentID(context.Context) (string, error) { return "node1", nil }

func (n *imageNode) ImageList(context.Context, ...docker.ImageListOption) (*client.ImageListResult, error) {
	return &client.ImageListResult{Items: n.images}, nil
}

func (n *imageNode) ContainerList(context.Context, ...docker.ContainerListOption) (*client.ContainerListResult,
	error) {
	return &client.ContainerListResult{Items: n.containers}, nil
}

func (n *imageNode) ServiceList(context.Context, ...docker.ServiceListOption) (*client.ServiceListResult, error) {
	return &client.ServiceListResult{Items: n.services}, nil
}

func (n *imageNode) ImageRemove(_ context.Context, ref string, _ ...docker.ImageRemoveOption) (
	*client.ImageRemoveResult, error) {
	if err := n.removeErr[ref]; err != nil {
		return nil, hperrors.NewInfra(err)
	}
	n.removed = append(n.removed, ref)
	return &client.ImageRemoveResult{Items: []image.DeleteResponse{{Untagged: ref}, {Deleted: "sha256:" + ref}}}, nil
}

func (n *imageNode) ImagePrune(_ context.Context, danglingOnly bool, _ time.Duration, _ ...docker.ImagePruneOption) (
	*client.ImagePruneResult, error) {
	n.pruned = danglingOnly
	return &client.ImagePruneResult{}, nil
}

// unusedImages is the record, in memory.
type unusedImages struct {
	since map[string]time.Time
	ttl   time.Duration
}

func (r *unusedImages) Get(context.Context, string) (map[string]time.Time, error) {
	if r.since == nil {
		return map[string]time.Time{}, nil
	}
	return r.since, nil
}

func (r *unusedImages) Set(_ context.Context, _ string, since map[string]time.Time, ttl time.Duration) error {
	r.since, r.ttl = since, ttl
	return nil
}

func imageCleanupData(keepImages ...string) *clusterCleanupData {
	return &clusterCleanupData{
		ClusterCleanupReq: &clustercleanupservice.ClusterCleanupReq{
			CleanupSettings: &entity.SystemClusterCleanup{
				PruneImages:      true,
				GeneralRetention: timeutil.Duration(3 * timeutil.Day),
			},
			CleanupImages: base.CleanupFlagTrue,
			KeepImages:    keepImages,
		},
		Output: &entity.ClusterNodeCleanupOutput{},
	}
}

// An image pulled ten days ago, used until yesterday, is not used today: it
// goes three days from now, not today.
func TestAnImageUsedUntilYesterdayIsKeptForTheRetention(t *testing.T) {
	tenDaysAgo := time.Now().Add(-10 * timeutil.Day).Unix()
	node := &imageNode{images: []image.Summary{{ID: "sha256:app", RepoTags: []string{"app:v1"}, Created: tenDaysAgo}}}
	record := &unusedImages{}
	svc := &service{dockerManager: node, unusedImagesRepo: record}

	data := imageCleanupData("traefik:v3.7.13")
	assert.NoError(t, svc.cleanupImages(context.Background(), data))
	assert.Empty(t, node.removed)
	assert.WithinDuration(t, time.Now(), record.since["sha256:app"], time.Minute)
	assert.Equal(t, 10*timeutil.Day, record.ttl, "the retention, and a week")

	// Three days later.
	record.since["sha256:app"] = time.Now().Add(-3 * timeutil.Day)
	data = imageCleanupData("traefik:v3.7.13")
	assert.NoError(t, svc.cleanupImages(context.Background(), data))
	assert.Equal(t, []string{"app:v1"}, node.removed)
	assert.Equal(t, 1, data.Output.ImagesDeleted)
	assert.Empty(t, record.since)
}

func TestRemovingAnImageGoesThroughEachOfItsNames(t *testing.T) {
	old := time.Now().Add(-4 * timeutil.Day)
	node := &imageNode{
		images: []image.Summary{
			{ID: "sha256:two-tags", RepoTags: []string{"a:1", "b:1"}, RepoDigests: []string{"a@" + otherDigest},
				Size: 300, SharedSize: 100},
			{ID: "sha256:untagged", RepoTags: []string{"<none>:<none>"}, Size: 50, SharedSize: -1},
			{ID: "sha256:busy", RepoTags: []string{"busy:1"}},
			{ID: "sha256:broken", RepoTags: []string{"broken:1"}},
		},
		removeErr: map[string]error{
			"a@" + otherDigest: errdefs.ErrNotFound, // gone with its last tag
			"busy:1":           errdefs.ErrConflict, // a container was created from it since
			"broken:1":         errors.New("disk on fire"),
		},
	}
	record := &unusedImages{since: map[string]time.Time{
		"sha256:two-tags": old, "sha256:untagged": old, "sha256:busy": old, "sha256:broken": old,
	}}
	svc := &service{dockerManager: node, unusedImagesRepo: record}

	data := imageCleanupData("traefik:v3.7.13")
	err := svc.cleanupImages(context.Background(), data)
	assert.ErrorContains(t, err, "disk on fire")
	assert.Contains(t, data.Output.ImagesPruneError, "disk on fire")

	assert.Equal(t, []string{"a:1", "b:1", "sha256:untagged"}, node.removed)
	assert.Equal(t, 2, data.Output.ImagesDeleted)
	assert.Equal(t, uint64(200+50), data.Output.SpaceReclaimed)
	// The one that failed is tried again next run; the one in use starts over.
	assert.Equal(t, map[string]time.Time{"sha256:broken": old}, record.since)
}

func TestContainersStoppedOrNotKeepTheirImages(t *testing.T) {
	old := time.Now().Add(-4 * timeutil.Day)
	node := &imageNode{
		images: []image.Summary{{ID: "sha256:running"}, {ID: "sha256:stopped"}},
		containers: []container.Summary{
			{ImageID: "sha256:running", State: container.StateRunning},
			{ImageID: "sha256:stopped", State: container.StateExited},
		},
	}
	record := &unusedImages{since: map[string]time.Time{"sha256:running": old, "sha256:stopped": old}}
	svc := &service{dockerManager: node, unusedImagesRepo: record}

	assert.NoError(t, svc.cleanupImages(context.Background(), imageCleanupData("traefik:v3.7.13")))
	assert.Empty(t, node.removed)
	assert.Empty(t, record.since)
}

// A manager of an older version says nothing of what to keep: the node prunes
// as that version did, the images no tag names.
func TestWithoutImagesToKeepOnlyUntaggedImagesArePruned(t *testing.T) {
	node := &imageNode{images: []image.Summary{{ID: "sha256:app", RepoTags: []string{"app:v1"}}}}
	svc := &service{dockerManager: node, unusedImagesRepo: &unusedImages{}}

	assert.NoError(t, svc.cleanupImages(context.Background(), imageCleanupData()))
	assert.True(t, node.pruned)
	assert.Empty(t, node.removed)
}

func TestKeptImagesAreTheServicesNowAndBeforeAndTheRelease(t *testing.T) {
	spec := func(image string) swarm.ServiceSpec {
		return swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Image: image}}}
	}
	previous := spec("shop:v1")
	node := &imageNode{services: []swarm.Service{
		{Spec: spec("shop:v2"), PreviousSpec: &previous},
		{Spec: spec("idle:v1")},
		{Spec: swarm.ServiceSpec{}},
	}}

	kept, err := (&service{dockerManager: node}).KeptImages(context.Background())
	assert.NoError(t, err)
	assert.Contains(t, kept, "shop:v2")
	assert.Contains(t, kept, "shop:v1")
	assert.Contains(t, kept, "idle:v1")
	for _, img := range systemappservice.CurrentRelease().Images() {
		assert.Contains(t, kept, img)
	}
	assert.NotContains(t, kept, "")
}
