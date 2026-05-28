package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	CurrentSystemCleanupVersion = 1
)

var _ = registerSettingParser(base.SettingTypeSystemCleanup, &systemCleanupParser{})

type systemCleanupParser struct {
}

func (s *systemCleanupParser) New() SettingData {
	return &SystemCleanup{}
}

type SystemCleanup struct {
	Schedule          SchedJobSchedule       `json:"schedule"`
	DBObjectRetention DBObjectRetention      `json:"dbObjectRetention"`
	ClusterCleanup    SystemClusterCleanup   `json:"clusterCleanup"`
	BackupCleanup     SystemBackupCleanup    `json:"backupCleanup"`
	Notification      *BaseEventNotification `json:"notification,omitempty"`
}

type DBObjectRetention struct {
	Enabled        bool              `json:"enabled"`
	Tasks          timeutil.Duration `json:"tasks"`
	SysErrors      timeutil.Duration `json:"sysErrors"`
	Deployments    timeutil.Duration `json:"deployments"`
	DeletedObjects timeutil.Duration `json:"deletedObjects"`
}

type SystemClusterCleanup struct {
	Enabled              bool              `json:"enabled"`
	OnlyObjectsOlderThan timeutil.Duration `json:"onlyObjectsOlderThan"`
	PruneImages          bool              `json:"pruneImages"`
	PruneVolumes         bool              `json:"pruneVolumes"`
	PruneNetworks        bool              `json:"pruneNetworks"`
	PruneContainers      bool              `json:"pruneContainers"`
}

type SystemBackupCleanup struct {
	Enabled              bool              `json:"enabled"`
	CloudBackupRetention timeutil.Duration `json:"cloudBackupRetention,omitempty"`
	LocalBackupRetention timeutil.Duration `json:"localBackupRetention,omitempty"`
}

func (s *SystemCleanup) GetType() base.SettingType {
	return base.SettingTypeSystemCleanup
}

func (s *SystemCleanup) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.Notification != nil {
		refIDs.AddRefIDs(s.Notification.GetRefObjectIDs())
	}
	return refIDs
}

func (s *SystemCleanup) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSystemCleanupVersion {
		return false, nil
	}
	if setting.Version > CurrentSystemCleanupVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSystemCleanupVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsSystemCleanup() (*SystemCleanup, error) {
	return parseSettingAs[*SystemCleanup](s)
}

func (s *Setting) MustAsSystemCleanup() *SystemCleanup {
	return gofn.Must(s.AsSystemCleanup())
}
