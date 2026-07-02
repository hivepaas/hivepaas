package healthcheckserviceimpl

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
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
			return apperrors.New(err)
		}
		defer conn.Close()

		healthClient := grpc_health_v1.NewHealthClient(conn)
		resp, err := healthClient.Check(reqCtx, &grpc_health_v1.HealthCheckRequest{Service: healthchk.Service})
		if err != nil {
			return apperrors.New(err)
		}

		data.Output.GRPC.ReturnStatus = base.HealthcheckGRPCStatus(resp.Status)
		if healthchk.ReturnStatus != base.HealthcheckGRPCStatus(resp.Status) {
			return apperrors.New(apperrors.ErrActionFailed)
		}

	default:
		return apperrors.NewUnsupported(fmt.Sprintf("gRPC health version '%v'", healthchk.Version))
	}

	return nil
}
