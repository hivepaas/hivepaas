package appdeploymentserviceimpl

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
)

func (s *service) CreateDeploymentAndTask(
	app *entity.App,
	deploymentSettings *entity.AppDeploymentSettings,
	args appdeploymentservice.DeploymentArgs,
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
		RunAt:     timeNow,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	err := deploymentTask.SetArgs(&entity.TaskAppDeployArgs{
		Deployment: entity.ObjectID{ID: deployment.ID},
		NoCache:    args.NoCache,
		ImageTags:  args.ImageTags,
	})
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}

	return deployment, deploymentTask, nil
}
