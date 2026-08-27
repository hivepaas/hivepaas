package healthcheckserviceimpl

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
			return hperrors.Wrap(err)
		}
		defer conn.Close()

		healthClient := grpc_health_v1.NewHealthClient(conn)
		resp, err := healthClient.Check(reqCtx, &grpc_health_v1.HealthCheckRequest{Service: healthchk.Service})
		if err != nil {
			return hperrors.Wrap(err)
		}

		data.Output.GRPC.ReturnStatus = base.HealthcheckGRPCStatus(resp.Status)
		if healthchk.ReturnStatus != base.HealthcheckGRPCStatus(resp.Status) {
			return hperrors.Wrap(hperrors.ErrActionFailed)
		}

	default:
		return hperrors.NewUnsupported(fmt.Sprintf("gRPC health version '%v'", healthchk.Version))
	}

	return nil
}
