package getstarteduc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/getstarteduc/getstarteddto"
)

// RequestDashboardCert asks for the dashboard's certificate now, past the wait
// a failed attempt leaves. One already being obtained is refused rather than
// asked for twice.
func (uc *UC) RequestDashboardCert(
	ctx context.Context,
	auth *basedto.Auth,
	_ *getstarteddto.RequestDashboardCertReq,
) (*getstarteddto.RequestDashboardCertResp, error) {
	if err := requireAdmin(auth); err != nil {
		return nil, hperrors.Wrap(err)
	}

	var tasks []*entity.Task
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		item, err := uc.getStartedService.DashboardCert(ctx, db)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if err = refuseWhileObtaining(item); err != nil {
			return hperrors.Wrap(err)
		}
		certRequest, err := uc.getStartedService.RequestDashboardCert(ctx, db, true)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if err = refuseNotAsked(certRequest); err != nil {
			return hperrors.Wrap(err)
		}
		tasks = certRequest.Tasks
		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// A task can be picked up only once its row exists, which is once the
	// transaction has committed. A failure to schedule it only delays it: the
	// queue's own scan finds it.
	_ = uc.taskQueue.ScheduleTask(ctx, tasks...)

	item, err := uc.getStartedService.DashboardCert(ctx, uc.db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &getstarteddto.RequestDashboardCertResp{
		Meta: &basedto.Meta{},
		Data: getstarteddto.TransformDashboardCert(item),
	}, nil
}

func refuseWhileObtaining(item *getstartedservice.Item) error {
	if item.Status == getstartedservice.ItemStatusObtaining {
		return hperrors.NewConflict("The dashboard's certificate").
			WithExtraDetail("it is being obtained already")
	}
	return nil
}

// refuseNotAsked turns a request that asked for nothing into an answer that
// says why, rather than a button that seems to do nothing.
func refuseNotAsked(certRequest *getstartedservice.CertRequest) error {
	if certRequest.NotAsked != "" {
		return hperrors.NewConflict("The dashboard's certificate").WithExtraDetail("%s", certRequest.NotAsked)
	}
	return nil
}

func requireAdmin(auth *basedto.Auth) error {
	if auth == nil || auth.User == nil || !auth.User.IsAdmin() {
		return hperrors.Wrap(hperrors.ErrForbidden)
	}
	return nil
}
