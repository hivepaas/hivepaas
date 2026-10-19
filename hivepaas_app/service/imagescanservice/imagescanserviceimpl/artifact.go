package imagescanserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice/imagescanrun"
)

// artifactFile is a file of an app's artifacts its tasks bind: the build for
// the architecture of the node they run on.
type artifactFile struct {
	artifact *entity.AppArtifact
	arch     base.CPUArch
}

// appArtifacts are the files of an app's artifacts its running tasks bind, as
// what a scan reads beside its images. The nodes' architectures are the
// swarm's.
func (s *service) appArtifacts(ctx context.Context, app *entity.App, images []runningImage) (
	[]runningImage, error) {
	artifacts, err := s.artifactService.ArtifactsOf(ctx, s.db, app.ID)
	if err != nil || artifacts == nil {
		return nil, hperrors.Wrap(err)
	}
	archOf := map[string]base.CPUArch{}
	for _, img := range images {
		if _, done := archOf[img.NodeID]; done {
			continue
		}
		inspect, err := s.dockerManager.NodeInspect(ctx, img.NodeID)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if arch, known := base.CPUArchFromDocker(inspect.Node.Description.Platform.Architecture); known {
			archOf[img.NodeID] = arch
		}
	}
	return artifactTargets(artifacts, images, archOf), nil
}

// artifactTargets are the artifacts' files the images' nodes bind: one per
// build, on the first node that binds it, by its sha256 - the ID the cache
// keeps it by - and named by the artifact. A node whose architecture is not
// known binds none.
func artifactTargets(artifacts *entity.AppArtifactsSettings, images []runningImage,
	archOf map[string]base.CPUArch) []runningImage {
	if artifacts == nil {
		return nil
	}
	var targets []runningImage
	seen := map[string]bool{}
	for _, img := range images {
		arch, known := archOf[img.NodeID]
		if !known {
			continue
		}
		for _, artifact := range artifacts.Artifacts {
			if artifact == nil {
				continue
			}
			file := artifact.Files[arch]
			if file == nil || seen[file.SHA256] {
				continue
			}
			seen[file.SHA256] = true
			targets = append(targets, runningImage{
				ID:       "sha256:" + file.SHA256,
				Ref:      artifact.Name + " " + artifact.Version,
				NodeID:   img.NodeID,
				Platform: "linux/" + string(arch),
				Artifact: &artifactFile{artifact: artifact, arch: arch},
			})
		}
	}
	return targets
}

// ensureArtifact has an artifact's file on the node a scan runs on: the build
// its task binds, whatever that node's own architecture.
func (s *service) ensureArtifact(ctx context.Context, nodeID string, img runningImage) error {
	ensured, err := s.artifactService.EnsureOnNode(ctx, &artifactservice.EnsureOnNodeReq{NodeID: nodeID,
		Artifact: img.Artifact.artifact, Arch: img.Artifact.arch})
	if err != nil {
		return hperrors.Wrap(err)
	}
	if "sha256:"+ensured.SHA256 != img.ID {
		return hperrors.Wrap(hperrors.ErrImageScanFailed).WithExtraDetail(
			"Node %s has another build of %s than the one scanned.", nodeID, img.Artifact.artifact.Name)
	}
	return nil
}

// runReqOf is a scan of an image or an artifact's file, with the settings and
// the volume: an image from its node or its registry with the credentials
// given, a file from the artifacts cache. format is the scan's.
func runReqOf(img runningImage, cfg *entity.ImageScanSettings, volume *scanVolume, auth []byte,
	format string) *imagescanrun.RunReq {
	req := &imagescanrun.RunReq{
		Format:           format,
		TrivyImage:       trivyImage(),
		ImageID:          img.ID,
		ImageNode:        img.NodeID,
		RemoteRef:        img.RemoteRef,
		Platform:         img.Platform,
		DockerConfig:     auth,
		Volume:           volume.Mount,
		CacheDir:         cacheDir(),
		DBRepository:     cfg.DBRepository,
		JavaDBRepository: cfg.JavaDBRepository,
		NanoCPUs:         int64(cfg.ScanCPULimit() * 1e9), //nolint:mnd // CPUs in billionths
		MemoryBytes:      cfg.ScanMemoryLimit().Bytes(),
		Timeout:          trivyTimeout,
	}
	if img.Artifact != nil {
		req.Mode = imagescanrun.ModeRootfs
		req.RemoteRef, req.Platform, req.DockerConfig = "", "", nil
	}
	return req
}
