package backuprepodto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type CleanupBackupRepoReq struct {
	settings.GetSettingReq
}

func NewCleanupBackupRepoReq() *CleanupBackupRepoReq {
	return &CleanupBackupRepoReq{}
}

func (req *CleanupBackupRepoReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CleanupBackupRepoResp struct {
	Meta *basedto.Meta              `json:"meta"`
	Data *CleanupBackupRepoDataResp `json:"data"`
}

type CleanupBackupRepoDataResp struct {
	// SnapshotsInRepo is what the repository holds once the retention policy has been applied.
	SnapshotsInRepo int `json:"snapshotsInRepo"`

	// RecordsRemoved / RecordsAdded are the stored snapshot records reconciled against the
	// repository. In the normal case RecordsRemoved is what the retention policy expired, but the
	// engine reports no IDs for that, so this counts everything the repository no longer has -
	// including snapshots deleted outside the app. RecordsAdded counts ones that appeared the
	// same way.
	RecordsRemoved int `json:"recordsRemoved"`
	RecordsAdded   int `json:"recordsAdded"`

	// RemovedSnapshots lists what was dropped, so the result is auditable rather than just a count.
	RemovedSnapshots []*CleanupRemovedSnapshotResp `json:"removedSnapshots"`
}

type CleanupRemovedSnapshotResp struct {
	ID        string    `json:"id"`
	ShortID   string    `json:"shortId"`
	Time      time.Time `json:"time"`
	Hostname  string    `json:"hostname,omitempty"`
	Paths     []string  `json:"paths,omitempty"`
	SizeBytes int64     `json:"sizeBytes,omitempty"`
}

func TransformCleanupBackupRepo(
	snapshotsInRepo int,
	removed []*entity.BackupSnapshot,
	added int,
) *CleanupBackupRepoDataResp {
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

	return &CleanupBackupRepoDataResp{
		SnapshotsInRepo:  snapshotsInRepo,
		RecordsRemoved:   len(removed),
		RecordsAdded:     added,
		RemovedSnapshots: removedResp,
	}
}
