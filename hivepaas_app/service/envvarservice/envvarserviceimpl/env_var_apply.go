package envvarserviceimpl

import (
	"context"
	"sort"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

const (
	applyEnvVarRetryMax = 2
)

func (s *service) ApplyEnvVarsForApps(
	ctx context.Context,
	db database.IDB,
	appEnvData []*envvarservice.AppEnvVarData,
	transaction bool,
	concurrency bool,
) map[int]error {
	if len(appEnvData) == 0 {
		return nil
	}
	maxConcurrentTasks := gofn.If(concurrency, uint(0), 1)
	errMap := gofn.ExecTaskFuncEx(ctx, maxConcurrentTasks, false,
		func(ctx context.Context, appData *envvarservice.AppEnvVarData) error {
			return s.applyEnvVarForApp(ctx, appData, transaction)
		}, appEnvData...)
	return errMap
}

func (s *service) applyEnvVarForApp(
	ctx context.Context,
	appEnvData *envvarservice.AppEnvVarData,
	transaction bool,
) (err error) {
	if transaction {
		err = s.appService.ExecuteInTx(ctx, appEnvData.App, false, func(db database.Tx) error {
			return s.applyEnvVarToService(ctx, appEnvData)
		})
	} else {
		err = s.applyEnvVarToService(ctx, appEnvData)
	}
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *service) applyEnvVarToService(
	ctx context.Context,
	envVarData *envvarservice.AppEnvVarData,
) error {
	sort.Slice(envVarData.EnvVars, func(i, j int) bool {
		return envVarData.EnvVars[i].Key < envVarData.EnvVars[j].Key
	})

	app := envVarData.App
	envVars := make([]string, 0, len(envVarData.EnvVars))
	for _, env := range envVarData.EnvVars {
		envVars = append(envVars, env.ToString("="))
	}

	err := s.dockerManager.ServiceUpdateFunc(ctx, app.ServiceID, nil,
		func(_ int, service *swarm.Service) (bool, error) {
			if service.Spec.TaskTemplate.ContainerSpec == nil {
				service.Spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
			}
			currEnvVars := service.Spec.TaskTemplate.ContainerSpec.Env
			if gofn.ContentEqual(currEnvVars, envVars) { // No change
				return false, nil
			}
			service.Spec.TaskTemplate.ContainerSpec.Env = envVars
			return true, nil
		}, applyEnvVarRetryMax, 0)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
