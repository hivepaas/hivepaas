package clusterserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	itemRemovalRetryDelay = 3 * time.Second
	itemRemovalRetryIncr  = 3 * time.Second
)

func (s *service) ServiceInspect(
	ctx context.Context,
	serviceID string,
	caching bool,
) (*swarm.Service, error) {
	if serviceID == "" {
		return nil, nil
	}

	// TODO: handle caching flag

	resp, err := s.dockerManager.ServiceInspect(ctx, serviceID)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &resp.Service, nil
}

func (s *service) ServiceUpdate(
	ctx context.Context,
	serviceID string,
	version *swarm.Version,
	service *swarm.ServiceSpec,
	options ...docker.ServiceUpdateOption,
) (*client.ServiceUpdateResult, error) {
	if serviceID == "" {
		return nil, nil
	}
	resp, err := s.dockerManager.ServiceUpdate(ctx, serviceID, version, service, options...)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return resp, nil
}

func (s *service) ServiceRemove(
	ctx context.Context,
	serviceID string,
	retryMax int,
	retryDelay time.Duration,
) (err error) {
	if serviceID == "" {
		return nil
	}
	fn := func() error {
		_, err := s.dockerManager.ServiceRemove(ctx, serviceID)
		if err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				return nil
			}
			return apperrors.New(err)
		}
		return nil
	}
	if retryMax > 0 {
		if retryDelay <= 0 {
			retryDelay = itemRemovalRetryDelay
		}
		err = gofn.ExecRetryCtx(ctx, fn, retryMax, retryDelay, gofn.ExecRetryDelayIncr(itemRemovalRetryIncr))
	} else {
		err = fn()
	}
	if err != nil {
		// TODO: create a cleanup task
		return apperrors.New(err)
	}
	return nil
}

func (s *service) ServicesRemove(
	ctx context.Context,
	serviceIDs []string,
	retryMax int,
	retryDelay time.Duration,
) (err error) {
	if len(serviceIDs) == 0 {
		return nil
	}
	if len(serviceIDs) == 1 {
		return s.ServiceRemove(ctx, serviceIDs[0], retryMax, retryDelay)
	}
	errMap := gofn.ExecTaskFuncEx(ctx, 10, false, //nolint:mnd
		func(ctx context.Context, itemID string) error {
			return s.ServiceRemove(ctx, itemID, retryMax, retryDelay)
		}, serviceIDs...)
	for _, e := range errMap {
		err = errors.Join(err, e)
	}
	return err
}
