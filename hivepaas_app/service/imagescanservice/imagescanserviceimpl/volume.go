package imagescanserviceimpl

import (
	"context"
	"errors"
	"path"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice/imagescanrun"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

// cacheSubdir is the directory inside the volume Trivy's cache is in: where a
// system app of env scanning would keep its data, so the volume can hold
// other things beside it.
const cacheSubdir = base.HivepaasProjectKey + "/scanning/trivy"

func (s *service) Validate(ctx context.Context, db database.IDB, cfg *entity.ImageScanSettings) error {
	if cfg == nil || !cfg.Enabled {
		return nil
	}
	if cfg.Volume.ID == "" {
		return hperrors.Wrap(hperrors.ErrImageScanSettingsInvalid).
			WithExtraDetail("Choose the volume Trivy keeps its DBs and its layer cache on.")
	}
	setting, err := s.settingRepo.GetByID(ctx, db, entity.NewObjectScopeGlobal(), base.SettingTypeClusterVolume,
		cfg.Volume.ID, true)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if setting == nil {
		return hperrors.Wrap(hperrors.ErrImageScanSettingsInvalid).
			WithExtraDetail("The volume chosen for image scanning no longer exists.")
	}
	vol, err := setting.AsClusterVolume()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if isMemory(vol) {
		return hperrors.Wrap(hperrors.ErrImageScanSettingsInvalid).WithExtraDetail(
			"The volume %q is memory, emptied when each scan's container stops: every scan would download "+
				"Trivy's DBs again, about 3 GB. Choose one on a disk.", setting.Name)
	}
	if volumeservice.DriverOptsReachDockerSocket(vol.DriverOpts) {
		return hperrors.Wrap(hperrors.ErrImageScanSettingsInvalid).
			WithExtraDetail("The volume %q reaches the Docker socket of the node.", setting.Name)
	}
	return nil
}

// isMemory says a volume is a Tmpfs one.
func isMemory(vol *entity.ClusterVolume) bool {
	return vol.DriverOpts["type"] == "tmpfs"
}

// scanVolume is the volume Trivy's data is on, mounted as scans mount it.
type scanVolume struct {
	*volumeservice.VolumeMount
	// labelNodes are the ready nodes carrying the label of a volume pinned by
	// one, sorted.
	labelNodes []string
}

// loadVolume reads the configuration's volume as scans mount it.
func (s *service) loadVolume(ctx context.Context, db database.IDB, cfg *entity.ImageScanSettings) (
	*scanVolume, error) {
	if err := s.Validate(ctx, db, cfg); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrImageScanNotReady).WithExtraDetail("%s",
			hperrors.GetErrorDetail(err, ""))
	}
	mnt, err := s.volumeService.BuildVolumeMount(ctx, db, cfg.Volume.ID, imagescanrun.DataDir)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	volume := &scanVolume{VolumeMount: mnt}
	if label := mnt.Volume.NodeLabel; mnt.Volume.NodeID == "" && label != "" {
		nodes, err := s.agentService.NodeIDsWithLabel(ctx, label)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if len(nodes) == 0 {
			return nil, hperrors.Wrap(hperrors.ErrImageScanNotReady).
				WithExtraDetail("No ready node carries the label %q of the volume scans run by.", label)
		}
		slices.Sort(nodes)
		volume.labelNodes = nodes
	}
	return volume, nil
}

// scanNode is the node an image running on imageNode is scanned on: the
// volume's, or for a shared volume, imageNode. Of the nodes carrying a
// volume's label, the one running the image when one does, which can read it
// from its own Docker.
func (v *scanVolume) scanNode(imageNode string) string {
	switch {
	case v.Volume.NodeID != "":
		return v.Volume.NodeID
	case len(v.labelNodes) > 0:
		if slices.Contains(v.labelNodes, imageNode) {
			return imageNode
		}
		return v.labelNodes[0]
	}
	return imageNode
}

// lockKey is the lock scans on a node take: one at a time per copy of the
// volume - one per node for a pinned one, one for the cluster for a shared one.
// Trivy's cache is a file one process opens at a time; a second fails.
func (v *scanVolume) lockKey(scanNode string) string {
	key := "imagescan:volume:lock:" + v.Setting.ID
	if v.Volume.IsPinned() {
		key += ":" + scanNode
	}
	return key
}

// cacheDir is Trivy's cache directory, in the scan's container.
func cacheDir() string {
	return path.Join(imagescanrun.DataDir, cacheSubdir)
}
