package systemcleanupdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type UpdateSystemCleanupReq struct {
	settings.UpdateSettingReq
	*SystemCleanupBaseReq
}

type SystemCleanupBaseReq struct {
	Status            base.SettingStatus                `json:"status"`
	Schedule          ScheduleReq                       `json:"schedule"`
	DBObjectRetention DBObjectRetentionReq              `json:"dbObjectRetention"`
	ClusterCleanup    SystemClusterCleanupReq           `json:"clusterCleanup"`
	BackupCleanup     SystemBackupCleanupReq            `json:"backupCleanup"`
	Notification      *basedto.BaseEventNotificationReq `json:"notification"`
}

func (req *SystemCleanupBaseReq) ToEntity() *entity.SystemCleanup {
	return &entity.SystemCleanup{
		Schedule:          req.Schedule.ToEntity(),
		DBObjectRetention: req.DBObjectRetention.ToEntity(),
		ClusterCleanup:    req.ClusterCleanup.ToEntity(),
		BackupCleanup:     req.BackupCleanup.ToEntity(),
		Notification:      req.Notification.ToEntity(),
	}
}

func (req *SystemCleanupBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	sched := req.Schedule.ToEntity()
	res = append(res, vld.Must((&sched).IsValid() == nil).OnError(
		vld.SetField(field+"schedule.Interval|schedule.CronExpr", nil),
		vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED_ONLY"),
	))
	res = append(res, basedto.ValidateTime(&req.Schedule.InitialTime, true,
		time.Now().Add(-timeutil.Dur365Days), time.Time{}, field+"schedule.initialTime")...)
	res = append(res, req.DBObjectRetention.validate(field+"dbObjectRetention")...)
	res = append(res, req.ClusterCleanup.validate(field+"clusterCleanup")...)
	res = append(res, req.BackupCleanup.validate(field+"backupCleanup")...)
	res = append(res, req.Notification.Validate(field+"notification")...)
	return res
}

type ScheduleReq struct {
	CronExpr    string            `json:"cronExpr"` // cronExpr and interval are mutually exclusive
	Interval    timeutil.Duration `json:"interval"`
	InitialTime time.Time         `json:"initialTime"`
}

func (req *ScheduleReq) ToEntity() entity.SchedJobSchedule {
	return entity.SchedJobSchedule{
		CronExpr:    req.CronExpr,
		Interval:    req.Interval,
		InitialTime: req.InitialTime,
	}
}

type DBObjectRetentionReq struct {
	Enabled        bool              `json:"enabled"`
	Tasks          timeutil.Duration `json:"tasks"`
	SysErrors      timeutil.Duration `json:"sysErrors"`
	Deployments    timeutil.Duration `json:"deployments"`
	DeletedObjects timeutil.Duration `json:"deletedObjects"`
}

func (req *DBObjectRetentionReq) ToEntity() entity.DBObjectRetention {
	return entity.DBObjectRetention{
		Enabled:        req.Enabled,
		Tasks:          req.Tasks,
		SysErrors:      req.SysErrors,
		Deployments:    req.Deployments,
		DeletedObjects: req.DeletedObjects,
	}
}

func (req *DBObjectRetentionReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	oneDay := timeutil.Duration(time.Hour * 24) //nolint:mnd
	durValid := req.Tasks >= oneDay && req.SysErrors >= oneDay &&
		req.Deployments >= oneDay && req.DeletedObjects >= oneDay
	res = append(res, vld.Must(durValid).OnError(
		vld.SetField(field+"duration values", nil),
		vld.SetCustomKey("ERR_VLD_VALUE_MUST_GREATER_THAN"),
		vld.SetParam("Min", oneDay.String()),
	))
	return res
}

type SystemClusterCleanupReq struct {
	Enabled         bool `json:"enabled"`
	PruneImages     bool `json:"pruneImages"`
	PruneVolumes    bool `json:"pruneVolumes"`
	PruneNetworks   bool `json:"pruneNetworks"`
	PruneContainers bool `json:"pruneContainers"`
}

func (req *SystemClusterCleanupReq) ToEntity() entity.SystemClusterCleanup {
	return entity.SystemClusterCleanup{
		Enabled:         req.Enabled,
		PruneImages:     req.PruneImages,
		PruneVolumes:    req.PruneVolumes,
		PruneNetworks:   req.PruneNetworks,
		PruneContainers: req.PruneContainers,
	}
}

func (req *SystemClusterCleanupReq) validate(_ string) []vld.Validator {
	return nil
}

type SystemBackupCleanupReq struct {
	Enabled              bool              `json:"enabled"`
	CloudBackupRetention timeutil.Duration `json:"cloudBackupRetention"`
	LocalBackupRetention timeutil.Duration `json:"localBackupRetention"`
}

func (req *SystemBackupCleanupReq) ToEntity() entity.SystemBackupCleanup {
	return entity.SystemBackupCleanup{
		Enabled:              req.Enabled,
		CloudBackupRetention: req.CloudBackupRetention,
		LocalBackupRetention: req.LocalBackupRetention,
	}
}

func (req *SystemBackupCleanupReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	durValid := req.CloudBackupRetention >= 0 && req.LocalBackupRetention >= 0
	res = append(res, vld.Must(durValid).OnError(
		vld.SetField(field+"duration values", nil),
		vld.SetCustomKey("ERR_VLD_VALUE_MUST_GREATER_THAN"),
		vld.SetParam("Min", 0),
	))
	return res
}

func NewUpdateSystemCleanupReq() *UpdateSystemCleanupReq {
	return &UpdateSystemCleanupReq{}
}

func (req *UpdateSystemCleanupReq) ModifyRequest() error {
	if req.Schedule.InitialTime.IsZero() {
		req.Schedule.InitialTime = timeutil.NowUTC()
	}
	req.Schedule.InitialTime = req.Schedule.InitialTime.Truncate(time.Minute)
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateSystemCleanupReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSystemCleanupResp struct {
	Meta *basedto.Meta `json:"meta"`
}
