package imageuc

import (
	"context"

	"github.com/moby/moby/api/types/registry"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/imageuc/imagedto"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (uc *UC) CreateImage(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagedto.CreateImageReq,
) (*imagedto.CreateImageResp, error) {
	data := &createImageData{}
	err := uc.loadImageData(ctx, uc.db, req, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	_, err = uc.dockerManager.ImagePull(ctx, req.Name, func(opts *client.ImagePullOptions) {
		opts.RegistryAuth = data.AuthHeader
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &imagedto.CreateImageResp{
		Data: &basedto.ObjectIDResp{},
	}, nil
}

type createImageData struct {
	RegistryAuth *entity.RegistryAuth
	AuthHeader   string
}

func (uc *UC) loadImageData(
	ctx context.Context,
	db database.IDB,
	req *imagedto.CreateImageReq,
	data *createImageData,
) error {
	if req.RegistryAuth.ID != "" {
		regAuth, err := uc.settingRepo.GetByID(ctx, db, base.NewObjectScopeGlobal(), base.SettingTypeRegistryAuth,
			req.RegistryAuth.ID, true)
		if err != nil {
			return apperrors.Wrap(err)
		}

		data.RegistryAuth = regAuth.MustAsRegistryAuth()
		password, err := data.RegistryAuth.Password.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
		data.AuthHeader, err = docker.GenerateAuthHeader(&registry.AuthConfig{
			Username: data.RegistryAuth.Username,
			Password: password,
		})
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}
