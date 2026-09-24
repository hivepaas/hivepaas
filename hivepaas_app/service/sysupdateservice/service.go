package sysupdateservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	SysUpdate(ctx context.Context, db database.IDB, req *SysUpdateReq) (*SysUpdateResp, error)

	// PlanUpdate says what an update to target would do, component by component,
	// without doing any of it: what each service runs now, what it would run,
	// and whether the update would move it at all. It asks what the update asks,
	// so what it says is what the update would find.
	PlanUpdate(ctx context.Context, db database.IDB, target *base.ReleaseInfo) (*UpdatePlan, error)
}
