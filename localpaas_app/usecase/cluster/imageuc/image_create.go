package imageuc

import (
	"context"

	"github.com/docker/docker/api/types/image"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/imageuc/imagedto"
	"github.com/localpaas/localpaas/services/docker"
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

	options := func(opts *image.CreateOptions) {
		opts.RegistryAuth = data.AuthHeader
	}
	_, err = uc.dockerManager.ImageCreate(ctx, req.Name, options)
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
		regAuth, err := uc.settingRepo.GetByID(ctx, db, base.NewSettingScopeGlobal(), base.SettingTypeRegistryAuth,
			req.RegistryAuth.ID, true)
		if err != nil {
			return apperrors.Wrap(err)
		}

		data.RegistryAuth = regAuth.MustAsRegistryAuth().MustDecrypt()
		data.AuthHeader, err = docker.GenerateAuthHeader(data.RegistryAuth.Username,
			data.RegistryAuth.Password.MustGetPlain())
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}
