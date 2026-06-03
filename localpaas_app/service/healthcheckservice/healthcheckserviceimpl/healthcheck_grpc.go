package healthcheckserviceimpl

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
)

func (s *service) doHealthcheckGRPC(
	ctx context.Context,
	data *healthcheckData,
) (err error) {
	healthchk := data.Healthcheck.GRPC
	if data.Output.GRPC == nil {
		data.Output.GRPC = &entity.TaskHealthcheckOutputGRPC{}
	}

	reqCtx := ctx
	if data.Healthcheck.Timeout > 0 {
		ctx, cancel := context.WithTimeout(ctx, data.Healthcheck.Timeout.ToDuration())
		defer cancel()
		reqCtx = ctx
	}

	switch healthchk.Version {
	case base.HealthcheckGRPCV1:
		conn, err := grpc.NewClient(
			healthchk.Addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		defer conn.Close()

		healthClient := grpc_health_v1.NewHealthClient(conn)
		resp, err := healthClient.Check(reqCtx, &grpc_health_v1.HealthCheckRequest{Service: healthchk.Service})
		if err != nil {
			return apperrors.Wrap(err)
		}

		data.Output.GRPC.ReturnStatus = base.HealthcheckGRPCStatus(resp.Status)
		if healthchk.ReturnStatus != base.HealthcheckGRPCStatus(resp.Status) {
			return apperrors.Wrap(apperrors.ErrActionFailed)
		}

	default:
		return apperrors.NewUnsupported(apperrors.Fmt("gRPC health version '%v'", healthchk.Version))
	}

	return nil
}
