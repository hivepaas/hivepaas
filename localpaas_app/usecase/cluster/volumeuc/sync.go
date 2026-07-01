package volumeuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/localpaas/localpaas/services/docker"
)

func (uc *UC) SyncVolume(
	ctx context.Context,
	auth *basedto.Auth,
	_ *volumedto.SyncVolumeReq,
) (*volumedto.SyncVolumeResp, error) {
	// 1. Scan docker to get list of volumes
	res, err := uc.dockerManager.VolumeList(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}

	// 2. Get list of existing settings from DB
	dbSettings, _, err := uc.SettingRepo.List(ctx, uc.DB, nil, nil,
		bunex.SelectWhere("setting.type = ?", currentSettingType),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}

	existingVolIDs := make(map[string]bool, len(dbSettings))
	for _, s := range dbSettings {
		volData, err := s.AsClusterVolume()
		if err != nil {
			return nil, apperrors.New(err)
		}
		existingVolIDs[volData.VolumeID] = true
	}

	// 3. For each docker volume, if not exists in DB, create new setting
	var newSettings []*entity.Setting
	for i := range res.Items {
		vol := &res.Items[i]
		volID := vol.Name
		if vol.ClusterVolume != nil {
			volID = vol.ClusterVolume.ID
		}

		if existingVolIDs[volID] {
			continue
		}

		// Insert setting
		volEntity := &entity.ClusterVolume{
			VolumeID: volID,
			Name:     vol.Name,
			Driver:   docker.VolumeDriver(vol.Driver),
		}

		setting := &entity.Setting{
			ID:      gofn.Must(ulid.NewStringULID()),
			Scope:   base.ObjectScopeGlobal,
			Type:    currentSettingType,
			Kind:    vol.Driver,
			Status:  base.SettingStatusActive,
			Name:    vol.Name,
			Version: currentSettingVersion,
		}
		if err := setting.SetData(volEntity); err != nil {
			return nil, apperrors.New(err)
		}
		newSettings = append(newSettings, setting)
	}

	if err := uc.SettingRepo.InsertMulti(ctx, uc.DB, newSettings); err != nil {
		return nil, apperrors.New(err)
	}

	return &volumedto.SyncVolumeResp{}, nil
}
