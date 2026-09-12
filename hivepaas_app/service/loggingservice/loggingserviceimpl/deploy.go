package loggingserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
)

func (s *service) Apply(_ context.Context, _ database.IDB) error {
	return hperrors.Wrap(loggingservice.ErrDeployFailed).WithExtraDetail("not implemented yet")
}

func (s *service) TearDown(_ context.Context) error {
	return hperrors.Wrap(loggingservice.ErrDeployFailed).WithExtraDetail("not implemented yet")
}

func (s *service) Status(_ context.Context, _ database.IDB) (*loggingservice.Status, error) {
	return nil, hperrors.Wrap(loggingservice.ErrDeployFailed).WithExtraDetail("not implemented yet")
}
