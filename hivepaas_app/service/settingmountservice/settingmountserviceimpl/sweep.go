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
//
// Without a service to read, nothing is removed: a service being recreated is
// between removal and creation, and its new spec needs what it references.
// Deleting an app is RemoveApp's.
func (s *service) Sweep(ctx context.Context, app *entity.App) error {
	if app.ServiceID == "" {
		return nil
	}
	inspect, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil
		}
		return hperrors.Wrap(err)
	}
	referenced := map[string]bool{}
	if contSpec := inspect.Service.Spec.TaskTemplate.ContainerSpec; contSpec != nil {
		for _, ref := range contSpec.Secrets {
			referenced[ref.SecretID] = true
		}
		for _, ref := range contSpec.Configs {
			referenced[ref.ConfigID] = true
		}
	}
	have, err := s.listMounted(ctx, settingmountservice.LabelAppID+"="+app.ID, settingmountservice.LabelEntry)
	if err != nil {
		return hperrors.Wrap(err)
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
