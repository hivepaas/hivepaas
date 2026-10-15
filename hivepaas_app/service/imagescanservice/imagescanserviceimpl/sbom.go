package imagescanserviceimpl

import (
	"context"

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
	images, err := s.appImages(ctx, req.App, tasklog.NewNullStore())
	if err != nil {
		return nil, err
	}
	var img *runningImage
	for i := range images {
		if images[i].ID == req.ImageID {
			img = &images[i]
		}
	}
	if img == nil {
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

	resp, err := s.runLocked(ctx, volume, volume.scanNode(img.NodeID), &imagescanrun.RunReq{
		Format:       imagescanrun.FormatCycloneDX,
		TrivyImage:   trivyImage(),
		ImageID:      img.ID,
		ImageNode:    img.NodeID,
		RemoteRef:    img.RemoteRef,
		Platform:     img.Platform,
		DockerConfig: config,
		Volume:       volume.Mount,
		CacheDir:     cacheDir(),
		NanoCPUs:     int64(cfg.ScanCPULimit() * 1e9), //nolint:mnd // CPUs in billionths
		MemoryBytes:  cfg.ScanMemoryLimit().Bytes(),
		Timeout:      trivyTimeout,
	})
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
