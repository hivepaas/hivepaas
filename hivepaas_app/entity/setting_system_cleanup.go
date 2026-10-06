package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	// CurrentSystemCleanupVersion 2 added SystemAppsSync.
	CurrentSystemCleanupVersion = 2
)

var _ = registerSettingParser(base.SettingTypeSystemCleanup, &systemCleanupParser{})

type systemCleanupParser struct {
}

func (s *systemCleanupParser) New() SettingData {
	return &SystemCleanup{}
}

type SystemCleanup struct {
	Schedule          SchedJobSchedule     `json:"schedule"`
	DBObjectRetention DBObjectRetention    `json:"dbObjectRetention"`
	ClusterCleanup    SystemClusterCleanup `json:"clusterCleanup"`
	CacheCleanup      SystemCacheCleanup   `json:"cacheCleanup"`
	FileCleanup       SystemFileCleanup    `json:"fileCleanup"`
	// SystemAppsSync is on when absent: a setting saved before it existed
	// syncs too. See SystemAppsSyncEnabled.
	SystemAppsSync *SystemAppsSync        `json:"systemAppsSync,omitempty"`
	Notification   *BaseEventNotification `json:"notification,omitempty"`
}

// SystemAppsSync brings the apps HivePaaS runs for itself - the registry, the
// logging stack - to their settings: one switched off is removed, its data
// kept; one missing is created, or deployed again when its service is gone.
// What it cannot mend - one scaled to zero, one whose tasks fail - and the
// nodes running OBI are reported.
type SystemAppsSync struct {
	Enabled bool `json:"enabled"`
}

// SystemAppsSyncEnabled says whether the system apps are synced: unless
// switched off.
func (s *SystemCleanup) SystemAppsSyncEnabled() bool {
	return s.SystemAppsSync == nil || s.SystemAppsSync.Enabled
}

type DBObjectRetention struct {
	Enabled        bool              `json:"enabled"`
	Tasks          timeutil.Duration `json:"tasks"`
	SysErrors      timeutil.Duration `json:"sysErrors"`
	AuditLogs      timeutil.Duration `json:"auditLogs"`
	Deployments    timeutil.Duration `json:"deployments"`
	DeletedObjects timeutil.Duration `json:"deletedObjects"`
}

type SystemClusterCleanup struct {
	Enabled             bool              `json:"enabled"`
	GeneralRetention    timeutil.Duration `json:"generalRetention"`
	BuildCacheRetention timeutil.Duration `json:"buildCacheRetention"`
	PruneImages         bool              `json:"pruneImages"`
	PruneVolumes        bool              `json:"pruneVolumes"`
	PruneNetworks       bool              `json:"pruneNetworks"`
	PruneContainers     bool              `json:"pruneContainers"`
	PruneBuildCache     bool              `json:"pruneBuildCache"`
}

type SystemCacheCleanup struct {
	Enabled            bool              `json:"enabled"`
	RepoCacheRetention timeutil.Duration `json:"repoCacheRetention,omitempty"`
}

type SystemFileCleanup struct {
	Enabled bool `json:"enabled"`
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

func (s *SystemCleanup) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsSystemCleanup() (*SystemCleanup, error) {
	return parseSettingAs[*SystemCleanup](s)
}

func (s *Setting) MustAsSystemCleanup() *SystemCleanup {
	return gofn.Must(s.AsSystemCleanup())
}
