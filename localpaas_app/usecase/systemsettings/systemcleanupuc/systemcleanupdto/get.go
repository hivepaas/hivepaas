package systemcleanupdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type GetSystemCleanupReq struct {
	settings.GetSettingReq
}

func NewGetSystemCleanupReq() *GetSystemCleanupReq {
	return &GetSystemCleanupReq{}
}

func (req *GetSystemCleanupReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetSystemCleanupResp struct {
	Meta *basedto.Meta      `json:"meta"`
	Data *SystemCleanupResp `json:"data"`
}

type SystemCleanupResp struct {
	*settings.BaseSettingResp
	ScheduleInterval  timeutil.Duration                  `json:"scheduleInterval"`
	ScheduleCronExpr  string                             `json:"scheduleCronExpr"`
	ScheduleFrom      time.Time                          `json:"scheduleFrom"`
	DBObjectRetention *DBObjectRetentionResp             `json:"dbObjectRetention"`
	ClusterCleanup    *SystemClusterCleanupResp          `json:"clusterCleanup"`
	BackupCleanup     *SystemBackupCleanupResp           `json:"backupCleanup"`
	Notification      *basedto.BaseEventNotificationResp `json:"notification"`
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
	return resp, nil
}
