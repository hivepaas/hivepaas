package appserviceimpl

import (
	"context"
	"encoding/json"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	labelHivePaaSAppPrevServiceMode = "hivepaas.app.prevServiceMode"
)

func (s *service) SetAppStatus(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	status base.AppStatus,
	recursive bool,
) error {
	// Update status of all child apps
	if app.ParentID == "" && recursive {
		childApps, _, err := s.appRepo.List(ctx, db, "", nil,
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			bunex.SelectWhere("app.parent_id = ?", app.ID),
		)
		if err != nil {
			return apperrors.New(err)
		}
		for _, childApp := range childApps {
			if err := s.SetAppStatus(ctx, db, childApp, status, recursive); err != nil {
				return apperrors.New(err)
			}
		}
	}

	if app.Status == status {
		return nil
	}
	app.Status = status
	app.UpdatedAt = timeutil.NowUTC()
	app.UpdateVer++

	if app.Status == base.AppStatusDisabled {
		if err := s.stopApp(ctx, app, nil); err != nil {
			return apperrors.New(err)
		}
	}
	if app.Status == base.AppStatusActive {
		if err := s.startApp(ctx, app, nil); err != nil {
			return apperrors.New(err)
		}
	}

	err := s.appRepo.Update(ctx, db, app, bunex.UpdateColumns("status", "updated_at", "update_ver"))
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (s *service) SetAppRunning(ctx context.Context, app *entity.App, running bool) error {
	if app.ServiceID == "" {
		return nil
	}
	inspect, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
	if err != nil {
		return apperrors.New(err)
	}
	service := &inspect.Service

	if service.Spec.Mode.Replicated == nil {
		return apperrors.New(apperrors.ErrServiceModeReplicatedRequired)
	}

	if running {
		return s.startApp(ctx, app, service)
	} else {
		return s.stopApp(ctx, app, service)
	}
}

func (s *service) stopApp(ctx context.Context, app *entity.App, service *swarm.Service) error {
	if app.ServiceID == "" {
		return nil
	}

	if service == nil {
		inspect, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
		if err != nil {
			return apperrors.New(err)
		}
		service = &inspect.Service
	}

	if service.Spec.Mode.Replicated != nil &&
		(service.Spec.Mode.Replicated.Replicas == nil || *service.Spec.Mode.Replicated.Replicas == 0) {
		return nil
	}

	prevSvcMode, err := json.Marshal(service.Spec.Mode)
	if err != nil {
		return apperrors.New(err)
	}
	service.Spec.Labels[labelHivePaaSAppPrevServiceMode] = string(prevSvcMode)

	// Scale down to 0
	service.Spec.Mode = swarm.ServiceMode{
		Replicated: &swarm.ReplicatedService{
			Replicas: new(uint64(0)),
		},
	}

	err = gofn.ExecRetry(func() error {
		_, err := s.dockerManager.ServiceUpdate(ctx, app.ServiceID, &service.Version, &service.Spec)
		return apperrors.New(err)
	}, 2, 3*time.Second) //nolint:mnd
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}

func (s *service) startApp(ctx context.Context, app *entity.App, service *swarm.Service) error {
	if app.ServiceID == "" {
		return nil
	}

	if service == nil {
		inspect, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
		if err != nil {
			return apperrors.New(err)
		}
		service = &inspect.Service
	}

	prevSvcModeStr := service.Spec.Labels[labelHivePaaSAppPrevServiceMode]
	if prevSvcModeStr != "" {
		mode := swarm.ServiceMode{}
		err := json.Unmarshal(reflectutil.UnsafeStrToBytes(prevSvcModeStr), &mode)
		if err != nil {
			return apperrors.New(err)
		}
		service.Spec.Mode = mode
		delete(service.Spec.Labels, labelHivePaaSAppPrevServiceMode)
	} else {
		service.Spec.Mode = swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: new(uint64(1)),
			},
		}
	}

	err := gofn.ExecRetry(func() error {
		_, err := s.dockerManager.ServiceUpdate(ctx, app.ServiceID, &service.Version, &service.Spec)
		return apperrors.New(err)
	}, 2, 3*time.Second) //nolint:mnd
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}
