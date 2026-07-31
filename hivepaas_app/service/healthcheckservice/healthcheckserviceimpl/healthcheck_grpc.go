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
	periodicJob := data.PeriodicSetting.MustAsPeriodicJob()
	healthchk := data.Healthcheck.GRPC
	if data.Output.GRPC == nil {
		data.Output.GRPC = &entity.TaskPeriodicHealthcheckOutputGRPC{}
	}

	reqCtx := ctx
	if periodicJob.Timeout > 0 {
		ctx, cancel := context.WithTimeout(ctx, periodicJob.Timeout.ToDuration())
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
		return apperrors.NewUnsupported(fmt.Sprintf("gRPC health version '%v'", healthchk.Version))
	}

	return nil
}
