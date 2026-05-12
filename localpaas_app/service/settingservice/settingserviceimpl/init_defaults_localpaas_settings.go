package settingserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
)

const (
	localPaaSSettingName        = "LocalPaaS settings"
	localPaaSWorkerReplicas     = 0
	localPaaSRunWorkerInMainApp = true
)

func (s *service) initDefaultLocalPaaSSettings(
	ctx context.Context,
	db database.IDB,
	timeNow time.Time,
) (err error) {
	localPaaSSetting := &entity.Setting{
		ID:              gofn.Must(ulid.NewStringULID()),
		Scope:           base.SettingScopeGlobal,
		Type:            base.SettingTypeLocalPaaSSettings,
		Status:          base.SettingStatusActive,
		Name:            localPaaSSettingName,
		AvailInProjects: true,
		Default:         true,
		Version:         entity.CurrentLocalPaaSSettingsVersion,
		CreatedAt:       timeNow,
		UpdatedAt:       timeNow,
	}
	cfg := config.Current
	localPaaS := &entity.LocalPaaSSettings{
		WorkerSettings: &entity.LocalPaaSWorkerSettings{
			Replicas:           localPaaSWorkerReplicas,
			Concurrency:        cfg.Tasks.Queue.Concurrency,
			RunWorkerInMainApp: localPaaSRunWorkerInMainApp,
		},
		TaskSettings: &entity.LocalPaaSTaskSettings{
			TaskCheckInterval:  timeutil.Duration(cfg.Tasks.Queue.TaskCheckInterval),
			TaskCreateInterval: timeutil.Duration(cfg.Tasks.Queue.TaskCreateInterval),
		},
		HealthcheckSettings: &entity.LocalPaaSHealthcheckSettings{
			BaseInterval: timeutil.Duration(cfg.Tasks.Healthcheck.BaseInterval),
		},
	}

	localPaaSSetting.MustSetData(localPaaS)

	err = s.settingRepo.Insert(ctx, db, localPaaSSetting)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
