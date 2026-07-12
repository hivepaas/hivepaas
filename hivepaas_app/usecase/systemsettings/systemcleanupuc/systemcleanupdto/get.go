package systemcleanupdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetSystemCleanupReq struct {
	settings.GetUniqueSettingReq
}

func NewGetSystemCleanupReq() *GetSystemCleanupReq {
	return &GetSystemCleanupReq{}
}

func (req *GetSystemCleanupReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetSystemCleanupResp struct {
	Meta *basedto.Meta      `json:"meta"`
	Data *SystemCleanupResp `json:"data"`
}

type SystemCleanupResp struct {
	*settings.BaseSettingResp
	Schedule          *ScheduleResp                      `json:"schedule"`
	DBObjectRetention *DBObjectRetentionResp             `json:"dbObjectRetention"`
	ClusterCleanup    *SystemClusterCleanupResp          `json:"clusterCleanup"`
	BackupCleanup     *SystemBackupCleanupResp           `json:"backupCleanup"`
	CacheCleanup      *SystemCacheCleanupResp            `json:"cacheCleanup"`
	FileCleanup       *SystemFileCleanupResp             `json:"fileCleanup"`
	Notification      *basedto.BaseEventNotificationResp `json:"notification"`

	// Calculated fields
	NextRuns []time.Time `json:"nextRuns"`
}

type ScheduleResp struct {
	CronExpr    string            `json:"cronExpr,omitempty"` // cronExpr and interval are mutually exclusive
	Interval    timeutil.Duration `json:"interval,omitempty"`
	InitialTime time.Time         `json:"initialTime"`
}

type DBObjectRetentionResp struct {
	Enabled        bool              `json:"enabled"`
	Tasks          timeutil.Duration `json:"tasks"`
	SysErrors      timeutil.Duration `json:"sysErrors"`
	Deployments    timeutil.Duration `json:"deployments"`
	DeletedObjects timeutil.Duration `json:"deletedObjects"`
}

type SystemClusterCleanupResp struct {
	Enabled         bool `json:"enabled"`
	PruneImages     bool `json:"pruneImages"`
	PruneVolumes    bool `json:"pruneVolumes"`
	PruneNetworks   bool `json:"pruneNetworks"`
	PruneContainers bool `json:"pruneContainers"`
}

type SystemBackupCleanupResp struct {
	Enabled              bool              `json:"enabled"`
	CloudBackupRetention timeutil.Duration `json:"cloudBackupRetention"`
	LocalBackupRetention timeutil.Duration `json:"localBackupRetention"`
}

type SystemCacheCleanupResp struct {
	Enabled            bool              `json:"enabled"`
	RepoCacheRetention timeutil.Duration `json:"repoCacheRetention"`
}

type SystemFileCleanupResp struct {
	Enabled bool `json:"enabled"`
}

func TransformSystemCleanup(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (resp *SystemCleanupResp, err error) {
	config := setting.MustAsSystemCleanup()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.Notification = basedto.TransformBaseEventNotification(config.Notification, refObjects)

	// Add next runs
	resp.NextRuns, _ = config.Schedule.CalcNextRuns(time.Now(), 5) //nolint

	return resp, nil
}
