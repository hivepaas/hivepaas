package volumeserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/volume"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

func (s *service) SyncVolumes(
	ctx context.Context,
	db database.IDB,
) ([]volume.Volume, error) {
	// 1. Scan docker to get list of volumes
	volList, err := s.dockerManager.VolumeList(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	currentSettingType := base.SettingTypeClusterVolume
	currentSettingVersion := entity.CurrentClusterVolumeVersion

	// 2. Get list of existing settings from DB
	dbSettings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", currentSettingType),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	existingVols := make(map[string]*entity.Setting, len(dbSettings)) // map docker-id -> *Setting
	for _, s := range dbSettings {
		existingVols[s.RefID] = s
	}

	// 3. For each docker volume, if not exists in DB, create new setting
	var updatingSettings []*entity.Setting
	for i := range volList.Items {
		vol := &volList.Items[i]
		volID := dockerhelper.GetVolumeID(vol)
		setting := existingVols[volID]

		if setting == nil {
			setting = &entity.Setting{
				ID:      gofn.Must(ulid.NewStringULID()),
				Scope:   base.ObjectScopeGlobal,
				Type:    currentSettingType,
				Kind:    vol.Driver,
				Status:  base.SettingStatusActive,
				Name:    vol.Name,
				RefID:   volID,
				Version: currentSettingVersion,
			}

			volEntity, err := s.discoveredVolumePinning(ctx, vol)
			if err != nil {
				return nil, hperrors.Wrap(err)
			}
			// Through SetData, because the parsed struct and the stored JSON are
			// not the same thing: Data is what is written, and nothing
			// re-serializes it on the way out.
			if err := setting.SetData(volEntity); err != nil {
				return nil, hperrors.Wrap(err)
			}
			updatingSettings = append(updatingSettings, setting)
			continue
		}

		delete(existingVols, volID)

		// The pinning of a volume already recorded is left exactly as it is - see
		// discoveredVolumePinning. Only what docker is authoritative about is
		// carried over. The chosen name is not: docker's name is already recorded
		// in RefID, and copying it onto Name would replace what an operator picked
		// with a raw docker identity the first time a lazily created volume
		// materializes.
		hasChanged := false
		if setting.Kind != vol.Driver {
			setting.Kind = vol.Driver
			hasChanged = true
		}

		backfilled, err := s.backfillVolumeSpec(setting, vol)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		hasChanged = hasChanged || backfilled

		if hasChanged {
			updatingSettings = append(updatingSettings, setting)
		}
	}

	// 4. A setting left in existingVols is not a deletion: a volume created on
	// this branch is not materialized in docker until a task first mounts it,
	// and then only on whichever node runs that task - usually not the manager,
	// which is the only daemon this scan can see. "Not in the manager's list" is
	// the normal state of a fresh volume, not evidence it was removed. A volume
	// actually being deleted goes through HivePaaS's own delete path.

	// 5. Upsert the settings
	err = s.settingRepo.UpsertMulti(ctx, db, updatingSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return volList.Items, nil
}

// discoveredVolumePinning is where a volume HivePaaS has just learned about
// lives, as far as docker can say.
//
// A swarm cluster volume belongs to no node in particular, so it is left
// unpinned. Anything else was listed by the daemon this process talks to, so it
// is on that node - which is a starting guess and not a fact. A bind mount whose
// directory comes from shared storage, an NFS or Ceph mount present at the same
// path on every node, looks exactly like one on a local disk: both are the local
// driver with type=none and a device path. Docker cannot tell them apart and
// neither can this.
//
// Which is why it runs once, for a volume being recorded for the first time.
// From then on the field is the operator's answer to this same question, and
// that includes leaving it empty - an empty pinning is the claim that the path
// is the same everywhere, not a gap waiting to be filled in. A sync that
// refreshed it every pass would overwrite that answer with this guess, every few
// minutes, and the operator would have no way to make it stick.
func (s *service) discoveredVolumePinning(
	ctx context.Context,
	vol *volume.Volume,
) (*entity.ClusterVolume, error) {
	if vol.ClusterVolume != nil && vol.ClusterVolume.ID != "" {
		return &entity.ClusterVolume{}, nil
	}

	nodeID, err := s.dockerManager.NodeCurrentID(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &entity.ClusterVolume{NodeID: nodeID}, nil
}

// backfillVolumeSpec records how a volume is built, for settings written before
// that was stored.
//
// It fills a gap and never corrects one: the pin and the specification are the
// operator's answers, and a sync that rewrote them would overwrite the answer
// every pass. A cluster volume is left alone entirely - swarm owns it, and
// HivePaaS claiming to be able to rebuild it would be a claim it cannot keep.
func (s *service) backfillVolumeSpec(setting *entity.Setting, vol *volume.Volume) (bool, error) {
	if vol.ClusterVolume != nil && vol.ClusterVolume.ID != "" {
		return false, nil
	}

	current, err := setting.AsClusterVolume()
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	if current == nil {
		current = &entity.ClusterVolume{}
	}
	if current.Driver != "" {
		return false, nil
	}

	current.Managed = true
	current.Driver = vol.Driver
	current.DriverOpts = vol.Options
	current.Labels = vol.Labels

	if err := setting.SetData(current); err != nil {
		return false, hperrors.Wrap(err)
	}
	return true, nil
}
