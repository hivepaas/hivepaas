package backuprepodto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type SyncBackupRepoReq struct {
	settings.GetSettingReq
}

func NewSyncBackupRepoReq() *SyncBackupRepoReq {
	return &SyncBackupRepoReq{}
}

func (req *SyncBackupRepoReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type SyncBackupRepoResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *SyncBackupRepoDataResp `json:"data"`
}

type SyncBackupRepoDataResp struct {
	// OptionsChanged says whether the repository turned out to be configured differently from what
	// the setting held. When false the fields below only restate what was already stored.
	OptionsChanged bool `json:"optionsChanged"`

	// Compression, PackSize and Retention are what the repository is really configured with, which
	// is what every later backup will use regardless of what the setting said before this ran.
	Compression string                     `json:"compression,omitempty"`
	PackSize    unit.DataSize              `json:"packSize,omitempty"`
	Retention   *BackupRetentionPolicyResp `json:"retention,omitempty"`

	// SnapshotsInRepo is what the repository holds. Nothing was expired to get this number: a sync
	// only reads, so this is the repository exactly as it stands.
	SnapshotsInRepo int `json:"snapshotsInRepo"`

	// RecordsRemoved / RecordsAdded are the stored records reconciled against the repository, so
	// they are the drift that happened outside the app - snapshots deleted or taken by something
	// else holding the repository password.
	RecordsRemoved int `json:"recordsRemoved"`
	RecordsAdded   int `json:"recordsAdded"`

	// RemovedSnapshots lists what was dropped, so the result is auditable rather than just a count.
	RemovedSnapshots []*CleanupRemovedSnapshotResp `json:"removedSnapshots"`
}

func TransformSyncBackupRepo(
	repo *entity.BackupRepo,
	optionsChanged bool,
	snapshotsInRepo int,
	removed []*entity.BackupSnapshot,
	added int,
) *SyncBackupRepoDataResp {
	removedResp := make([]*CleanupRemovedSnapshotResp, 0, len(removed))
	for _, snapshot := range removed {
		removedResp = append(removedResp, &CleanupRemovedSnapshotResp{
			ID:        snapshot.ID,
			ShortID:   snapshot.ShortID,
			Time:      snapshot.Time,
			Hostname:  snapshot.Hostname,
			Paths:     snapshot.Paths,
			SizeBytes: snapshot.SizeBytes,
		})
	}

	resp := &SyncBackupRepoDataResp{
		OptionsChanged:   optionsChanged,
		Compression:      repo.Compression,
		PackSize:         repo.PackSize,
		SnapshotsInRepo:  snapshotsInRepo,
		RecordsRemoved:   len(removed),
		RecordsAdded:     added,
		RemovedSnapshots: removedResp,
	}
	if repo.Retention != nil {
		resp.Retention = &BackupRetentionPolicyResp{
			KeepLast:    repo.Retention.KeepLast,
			KeepDaily:   repo.Retention.KeepDaily,
			KeepWeekly:  repo.Retention.KeepWeekly,
			KeepMonthly: repo.Retention.KeepMonthly,
		}
	}
	return resp
}
