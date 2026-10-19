package imagescanserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice/imagescanrun"
)

func (s *service) SBOM(ctx context.Context, req *imagescanservice.SBOMReq) ([]byte, error) {
	cfg, err := s.enabled(ctx, s.db)
	if err != nil {
		return nil, err
	}
	volume, err := s.loadVolume(ctx, s.db, cfg)
	if err != nil {
		return nil, err
	}
	// Some nodes unread, the image may still be on one that answered.
	images, err := s.appImages(ctx, req.App, tasklog.NewNullStore())
	var unread *unreadNodesError
	if err != nil && !errors.As(err, &unread) {
		return nil, err
	}
	if artifacts, artifactsErr := s.appArtifacts(ctx, req.App, images); artifactsErr == nil {
		images = append(images, artifacts...)
	}
	var img *runningImage
	for i := range images {
		if images[i].ID == req.ImageID {
			img = &images[i]
		}
	}
	if img == nil {
		if unread != nil {
			return nil, hperrors.Wrap(hperrors.ErrImageScanFailed).WithExtraDetail("%s", unread.Error())
		}
		return nil, hperrors.NewNotFound("Image").WithExtraDetail("The app does not run this image now.")
	}
	auth, err := s.pullAuth(ctx, req.App)
	if err != nil {
		auth = nil
	}
	config, err := dockerConfig(auth, img.RemoteRef)
	if err != nil {
		return nil, err
	}

	scanNode := volume.scanNode(img.NodeID)
	if img.Artifact != nil {
		if err = s.ensureArtifact(ctx, scanNode, *img); err != nil {
			return nil, err
		}
	}
	resp, err := s.runLocked(ctx, volume, scanNode, sbomLockTries,
		runReqOf(*img, cfg, volume, config, imagescanrun.FormatCycloneDX))
	if err != nil {
		return nil, err
	}
	out, err := imagescanservice.ParseScanOutput(resp.Output)
	if err != nil || !out.HasReport() {
		if line := lastErrorLine(resp.Log); line != "" {
			return nil, hperrors.Wrap(hperrors.ErrImageScanFailed).WithExtraDetail("%s", trivyMessage(line))
		}
		return nil, hperrors.Wrap(hperrors.ErrImageScanFailed).
			WithExtraDetail("Trivy stopped with exit code %d, and said nothing more.", resp.ExitCode)
	}
	return out.Report, nil
}
