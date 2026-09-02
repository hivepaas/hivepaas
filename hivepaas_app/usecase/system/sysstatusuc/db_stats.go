package sysstatusuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/sysstatusuc/sysstatusdto"
)

// GetDBStats reports the connection pool counters of this process. They are per-process, so with
// several instances behind a load balancer each answers for itself.
func (uc *UC) GetDBStats(
	_ context.Context,
	_ *basedto.Auth,
) (*sysstatusdto.GetDBStatsResp, error) {
	return &sysstatusdto.GetDBStatsResp{
		Data: sysstatusdto.TransformDBStats(uc.db.Stats()),
	}, nil
}
