package imagescanserviceimpl

import (
	"context"
	"errors"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
)

// Status reads where scans run and what the last one used. What cannot be
// read - a node Docker does not name - is left out rather than failing the
// screen.
func (s *service) Status(ctx context.Context, db database.IDB) (*imagescanservice.Status, error) {
	status := &imagescanservice.Status{Image: trivyImage()}
	_, cfg, err := s.Settings(ctx, db)
	if err != nil {
		return nil, err
	}
	if cfg != nil {
		status.Enabled = cfg.Enabled
		if cfg.Volume.ID != "" {
			if status.Volume, err = s.volumeStatus(ctx, db, cfg.Volume.ID); err != nil {
				return nil, err
			}
		}
	}

	latest, err := s.imageScanRepo.LatestScan(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if latest != nil {
		status.LastScanAt, status.DBUpdatedAt, status.TrivyVersion = latest.ScannedAt, latest.DBUpdatedAt,
			latest.TrivyVersion
	}
	// The error is the last try's, which LatestScan - done scans alone - never
	// is when it failed: scans failing since the last that worked say why.
	attempt, err := s.imageScanRepo.LatestAttempt(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if attempt != nil && attempt.Status == base.ImageScanStatusFailed {
		status.LastError = attempt.Error
	}
	return status, nil
}

func (s *service) volumeStatus(ctx context.Context, db database.IDB, volumeID string) (
	*imagescanservice.VolumeStatus, error) {
	status := &imagescanservice.VolumeStatus{ID: volumeID}
	setting, err := s.settingRepo.GetByID(ctx, db, entity.NewObjectScopeGlobal(), base.SettingTypeClusterVolume,
		volumeID, false)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			status.Missing = true
			return status, nil
		}
		return nil, hperrors.Wrap(err)
	}
	vol, err := setting.AsClusterVolume()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	status.Name, status.NodeID, status.NodeLabel = setting.Name, vol.NodeID, vol.NodeLabel
	status.Shared = !vol.IsPinned()
	if vol.NodeID == "" && vol.NodeLabel != "" {
		// What cannot be read leaves the screen as it is, as a node's name does.
		if nodes, err := s.agentService.NodeIDsWithLabel(ctx, vol.NodeLabel); err == nil && len(nodes) == 0 {
			status.Problem = fmt.Sprintf("No ready node carries the label %q: no scan can run.", vol.NodeLabel)
		}
	}
	if vol.NodeID != "" {
		if node, err := s.dockerManager.NodeInspect(ctx, vol.NodeID); err == nil {
			status.NodeName = node.Node.Description.Hostname
			if node.Node.Spec.Name != "" {
				status.NodeName = node.Node.Spec.Name
			}
		}
	}
	return status, nil
}
