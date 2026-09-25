package settingmountserviceimpl

import (
	"context"
	"errors"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

const removalRetryMax = 2

// Sweep removes the app's mounted objects that its service no longer
// references. Docker refuses to remove one a service still names; that one is
// retried, then left to the next sweep - a deployment is not failed for it.
func (s *service) Sweep(ctx context.Context, app *entity.App) error {
	have, err := s.listMounted(ctx, settingmountservice.LabelAppID+"="+app.ID, settingmountservice.LabelEntry)
	if err != nil {
		return hperrors.Wrap(err)
	}
	referenced := map[string]bool{}
	if app.ServiceID != "" {
		inspect, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return hperrors.Wrap(err)
		}
		if err == nil && inspect.Service.Spec.TaskTemplate.ContainerSpec != nil {
			for _, ref := range inspect.Service.Spec.TaskTemplate.ContainerSpec.Secrets {
				referenced[ref.SecretID] = true
			}
			for _, ref := range inspect.Service.Spec.TaskTemplate.ContainerSpec.Configs {
				referenced[ref.ConfigID] = true
			}
		}
	}
	return s.remove(ctx, have, referenced)
}

// RemoveApp removes every mounted object of an app, once its service is gone.
func (s *service) RemoveApp(ctx context.Context, appID string) error {
	have, err := s.listMounted(ctx, settingmountservice.LabelAppID+"="+appID, settingmountservice.LabelEntry)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return s.remove(ctx, have, nil)
}

func (s *service) remove(ctx context.Context, have *mounted, keep map[string]bool) (errs error) {
	for id := range have.secrets {
		if keep[id] {
			continue
		}
		errs = errors.Join(errs, s.retryRemoval(ctx, func() error {
			_, err := s.dockerManager.SecretRemove(ctx, id)
			return err //nolint:wrapcheck
		}))
	}
	for id := range have.configs {
		if keep[id] {
			continue
		}
		errs = errors.Join(errs, s.retryRemoval(ctx, func() error {
			_, err := s.dockerManager.ConfigRemove(ctx, id)
			return err //nolint:wrapcheck
		}))
	}
	return errs
}

func (s *service) retryRemoval(ctx context.Context, remove func() error) error {
	err := gofn.ExecRetryCtx(ctx, func() error {
		if err := remove(); err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return err
		}
		return nil
	}, removalRetryMax, s.removalRetryDelay)
	return hperrors.Wrap(err)
}
