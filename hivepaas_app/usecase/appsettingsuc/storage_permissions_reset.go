package appsettingsuc

import (
	"context"
	"strconv"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

// ResetAppStoragePermissions resets what is in the directory of one of an app's
// mounts. It is asked for, never done on a deployment: a deployment opens a
// directory up only while it is empty, and this is how an app is given data
// another user wrote.
//
// The files change as soon as the helper has run; the record is written after,
// since a directory walked through cannot be walked back.
func (uc *UC) ResetAppStoragePermissions(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.ResetAppStoragePermissionsReq,
) (*appsettingsdto.ResetAppStoragePermissionsResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	mnt, err := uc.findAppMount(ctx, app, req.Key)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resetReq := &volumeservice.ResetAppStoragePermissionsReq{App: app, Mount: *mnt}
	owner := "everyone"
	if req.Owner != nil {
		resetReq.Owner = &volumeservice.StorageOwner{UID: req.Owner.UID, GID: req.Owner.GID}
		owner = fmtOwner(req.Owner)
	}
	resp, err := uc.volumeService.ResetAppStoragePermissions(ctx, uc.db, resetReq)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		return uc.recordAppUpdate(ctx, db, auth, app, base.AuditLogSourceAPIUpdate, "storage", auditdetail.New().
			Set("action", "resetPermissions").
			Set("target", mnt.Target).
			Set("owner", owner))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appsettingsdto.ResetAppStoragePermissionsResp{
		Data: &appsettingsdto.ResetAppStoragePermissionsDataResp{Path: resp.Path},
	}, nil
}

// findAppMount is the app's mount the storage settings list under key.
func (uc *UC) findAppMount(ctx context.Context, app *entity.App, key string) (*mount.Mount, error) {
	if app.ServiceID != "" {
		service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, false)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if service != nil && service.Spec.TaskTemplate.ContainerSpec != nil {
			mounts := service.Spec.TaskTemplate.ContainerSpec.Mounts
			for i := range mounts {
				if uc.calcMountKey(&mounts[i]) == key {
					return &mounts[i], nil
				}
			}
		}
	}
	return nil, hperrors.NewNotFound("mount")
}

func fmtOwner(owner *appsettingsdto.StorageOwnerReq) string {
	return strconv.Itoa(owner.UID) + ":" + strconv.Itoa(owner.GID)
}
