package clustercleanupserviceimpl

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/distribution/reference"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

// unusedImagesRecordGrace is how long a node's record of its unused images
// outlives the retention: the record of a node no run reaches any more - one
// that left the swarm - goes by itself.
const unusedImagesRecordGrace = 7 * timeutil.Day

func (s *service) KeptImages(ctx context.Context) ([]string, error) {
	resp, err := s.dockerManager.ServiceList(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	images := systemappservice.CurrentRelease().Images()
	for i := range resp.Items {
		svc := &resp.Items[i]
		if spec := svc.Spec.TaskTemplate.ContainerSpec; spec != nil {
			images = append(images, spec.Image)
		}
		if prev := svc.PreviousSpec; prev != nil && prev.TaskTemplate.ContainerSpec != nil {
			images = append(images, prev.TaskTemplate.ContainerSpec.Image)
		}
	}
	images = slices.DeleteFunc(images, func(image string) bool { return image == "" })
	slices.Sort(images)
	return slices.Compact(images), nil
}

// cleanupImages removes the images of this node that nothing has used for the
// retention.
//
// Docker cannot tell when an image was last used: its prune counts the
// retention from when an image was built, or with the containerd image store
// from when it came to the node, so an image pulled long ago and used until
// yesterday would go today. Each run here notes instead when it first found an
// image unused, and an image goes once that is the retention ago. One used
// again starts over.
//
// An image is used when:
//   - a container of this node, running or stopped, runs it;
//   - a service names it, as it runs or as it ran before its last update. A
//     service scaled to zero runs nothing, but runs its image again when scaled
//     up, and a rollback runs the previous one; and an image built on a
//     single node is never pushed, so it could not be pulled back;
//   - this release of HivePaaS names it.
//
// The last two come from the manager, as KeepImages. A manager of an older
// version sends none: the node then prunes as that version did.
func (s *service) cleanupImages(ctx context.Context, data *clusterCleanupData) error {
	if data.CleanupImages == base.CleanupFlagFalse || !data.CleanupSettings.PruneImages {
		return nil
	}
	var err error
	if len(data.KeepImages) == 0 {
		err = s.pruneDanglingImages(ctx, data)
	} else {
		err = s.removeUnusedImages(ctx, data)
	}
	if err != nil {
		data.Output.ImagesPruneError = err.Error()
	}
	return hperrors.Wrap(err)
}

// pruneDanglingImages is Docker's prune of the images no tag names: what an
// older version did, the only safe thing without knowing what to keep.
func (s *service) pruneDanglingImages(ctx context.Context, data *clusterCleanupData) error {
	resp, err := s.dockerManager.ImagePrune(ctx, true, data.CleanupSettings.GeneralRetention.ToDuration())
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.Output.ImagesDeleted = len(resp.Report.ImagesDeleted)
	data.Output.SpaceReclaimed += resp.Report.SpaceReclaimed
	return nil
}

func (s *service) removeUnusedImages(ctx context.Context, data *clusterCleanupData) error {
	nodeID, err := s.dockerManager.NodeCurrentID(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	unusedSince, err := s.unusedImagesRepo.Get(ctx, nodeID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	images, err := s.dockerManager.ImageList(ctx, func(opts *client.ImageListOptions) {
		opts.SharedSize = true
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	// After the images: a container created in between, of an image listed, is
	// seen.
	containers, err := s.dockerManager.ContainerList(ctx, func(opts *client.ContainerListOptions) {
		opts.All = true
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	inUse := make(map[string]bool, len(containers.Items))
	for i := range containers.Items {
		inUse[containers.Items[i].ImageID] = true
	}

	now := time.Now().UTC()
	retention := data.CleanupSettings.GeneralRetention.ToDuration()
	plan := planImages(images.Items, inUse, newImageKeeper(data.KeepImages), unusedSince, now, retention)

	var errs error
	for i := range plan.remove {
		img := &plan.remove[i]
		removed, e := s.removeImage(ctx, img)
		switch {
		case e != nil:
			// Tried again next run.
			errs = errors.Join(errs, e)
			plan.unusedSince[img.ID] = sinceOr(unusedSince, img.ID, now)
		case removed:
			data.Output.ImagesDeleted++
			data.Output.SpaceReclaimed += uniqueSize(img)
		}
		// Neither: used after all, its count starts over.
	}
	if e := s.unusedImagesRepo.Set(ctx, nodeID, plan.unusedSince, retention+unusedImagesRecordGrace); e != nil {
		errs = errors.Join(errs, hperrors.Wrap(e))
	}
	return errs
}

// removeImage removes an image by every name it has - its tags, then the
// digests it was pulled by - or by its ID when it has none: Docker refuses to
// remove by ID an image several repositories name. It answers false when the
// image turns out to be used: a container was created from it since the images
// were listed, or another image is built on it.
func (s *service) removeImage(ctx context.Context, img *image.Summary) (bool, error) {
	refs := imageNames(img)
	if len(refs) == 0 {
		refs = []string{img.ID}
	}
	removed := false
	for _, ref := range refs {
		resp, err := s.dockerManager.ImageRemove(ctx, ref, func(opts *client.ImageRemoveOptions) {
			opts.PruneChildren = true
		})
		switch {
		case errors.Is(err, hperrors.ErrInfraConflict):
			return false, nil
		case errors.Is(err, hperrors.ErrInfraNotFound):
			// Gone with a name removed before.
			continue
		case err != nil:
			return false, hperrors.Wrap(err)
		}
		for _, item := range resp.Items {
			if item.Deleted != "" {
				removed = true
			}
		}
	}
	return removed, nil
}

// imagePlan is what a run does with a node's images: those it removes, and
// since when each image it leaves unused is known to be.
type imagePlan struct {
	remove      []image.Summary
	unusedSince map[string]time.Time
}

// planImages picks the images to remove: those no container uses and the
// keeper does not keep, found unused at least the retention ago. Every other
// unused image is recorded as unused since it was first found so - now, when
// it was not before.
func planImages(
	images []image.Summary,
	inUse map[string]bool,
	keeper *imageKeeper,
	unusedSince map[string]time.Time,
	now time.Time,
	retention time.Duration,
) *imagePlan {
	plan := &imagePlan{unusedSince: map[string]time.Time{}}
	for i := range images {
		img := &images[i]
		if inUse[img.ID] || keeper.keeps(img) {
			continue
		}
		since := sinceOr(unusedSince, img.ID, now)
		if now.Sub(since) >= retention {
			plan.remove = append(plan.remove, *img)
			continue
		}
		plan.unusedSince[img.ID] = since
	}
	return plan
}

// sinceOr is when an image was recorded unused, or now when it was not; or now
// when the record is later than now, as a clock set back makes it.
func sinceOr(unusedSince map[string]time.Time, imageID string, now time.Time) time.Time {
	since, found := unusedSince[imageID]
	if !found || since.After(now) {
		return now
	}
	return since
}

// imageKeeper tells the images to keep from the references it is given: by
// name and tag, or by digest.
type imageKeeper struct {
	// tagged are the references' normalized names with their tags, as
	// docker.io/library/traefik:v3.7.13.
	tagged map[string]bool
	// digests are the references' digests.
	digests map[string]bool
}

func newImageKeeper(refs []string) *imageKeeper {
	keeper := &imageKeeper{tagged: map[string]bool{}, digests: map[string]bool{}}
	for _, ref := range refs {
		named, err := reference.ParseNormalizedNamed(ref)
		if err != nil {
			continue
		}
		if canonical, ok := named.(reference.Canonical); ok {
			keeper.digests[canonical.Digest().String()] = true
			if _, tagged := named.(reference.Tagged); !tagged {
				continue
			}
		}
		named = reference.TagNameOnly(named)
		if tagged, ok := named.(reference.Tagged); ok {
			keeper.tagged[named.Name()+":"+tagged.Tag()] = true
		}
	}
	return keeper
}

// keeps tells whether an image is one of the references: one of its tags is
// theirs, or the digest it was pulled by. With the containerd image store, an
// image's ID is that digest too.
func (keeper *imageKeeper) keeps(img *image.Summary) bool {
	if keeper.digests[img.ID] {
		return true
	}
	for _, tag := range img.RepoTags {
		named, err := reference.ParseNormalizedNamed(tag)
		if err != nil {
			continue
		}
		if tagged, ok := named.(reference.Tagged); ok && keeper.tagged[named.Name()+":"+tagged.Tag()] {
			return true
		}
	}
	for _, ref := range img.RepoDigests {
		named, err := reference.ParseNormalizedNamed(ref)
		if err != nil {
			continue
		}
		if canonical, ok := named.(reference.Canonical); ok && keeper.digests[canonical.Digest().String()] {
			return true
		}
	}
	return false
}

// imageNames are an image's tags, then the digests it was pulled by; not the
// <none> Docker lists for an image without.
func imageNames(img *image.Summary) []string {
	var names []string
	for _, name := range slices.Concat(img.RepoTags, img.RepoDigests) {
		if _, err := reference.ParseNormalizedNamed(name); err == nil {
			names = append(names, name)
		}
	}
	return names
}

// uniqueSize is the space an image takes that no other image shares, which its
// removal frees at least.
func uniqueSize(img *image.Summary) uint64 {
	size := img.Size
	if img.SharedSize > 0 {
		size -= img.SharedSize
	}
	return uint64(max(size, 0))
}
