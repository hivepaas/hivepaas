package appdeploymentserviceimpl

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

func (s *service) CreateDeploymentAndTask(
	app *entity.App,
	deploymentSettings *entity.AppDeploymentSettings,
) (*entity.Deployment, *entity.Task, error) {
	timeNow := timeutil.NowUTC()
	deployment := &entity.Deployment{
		ID:        gofn.Must(ulid.NewStringULID()),
		AppID:     app.ID,
		Settings:  deploymentSettings,
		Status:    base.DeploymentStatusNotStarted,
		Version:   entity.CurrentDeploymentVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}

	deploymentTask := &entity.Task{
		ID:       gofn.Must(ulid.NewStringULID()),
		Scope:    base.ObjectScopeApp,
		ObjectID: app.ID,
		TargetID: deployment.ID,
		Type:     base.TaskTypeAppDeploy,
		Status:   base.TaskStatusNotStarted,
		Config: entity.TaskConfig{
			Priority: base.TaskPriorityDefault,
			Timeout:  timeutil.Duration(base.DeploymentTimeoutDefault),
		},
		Version:   entity.CurrentTaskVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	err := deploymentTask.SetArgs(&entity.TaskAppDeployArgs{
		Deployment: entity.ObjectID{ID: deployment.ID},
	})
	if err != nil {
		return nil, nil, apperrors.New(err)
	}

	return deployment, deploymentTask, nil
}
