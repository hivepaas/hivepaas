package nodeuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (uc *UC) DeleteNode(
	ctx context.Context,
	auth *basedto.Auth,
	req *nodedto.DeleteNodeReq,
) (*nodedto.DeleteNodeResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{
		AfterLoading: func(
			ctx context.Context,
			db database.Tx,
			data *settings.DeleteSettingData,
		) error {
			if data.Setting.ObjectID == req.Scope.MainObjectID() {
				var options []docker.NodeRemoveOption
				if req.Force {
					options = append(options, docker.NodeRemoveForce(true))
				}
				_, err := uc.dockerManager.NodeRemove(ctx, dockerhelper.ParseID(data.Setting.ID), options...)
				if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
					return apperrors.Wrap(err)
				}
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &nodedto.DeleteNodeResp{}, nil
}
