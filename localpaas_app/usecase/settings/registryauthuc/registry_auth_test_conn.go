package registryauthuc

import (
	"context"

	"github.com/docker/docker/api/types/registry"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/registryauthuc/registryauthdto"
)

func (uc *UC) TestRegistryAuthConn(
	ctx context.Context,
	auth *basedto.Auth,
	req *registryauthdto.TestRegistryAuthConnReq,
) (*registryauthdto.TestRegistryAuthConnResp, error) {
	_, err := uc.dockerManager.RegistryLogin(ctx, &registry.AuthConfig{
		Username:      req.Username,
		Password:      req.Password,
		ServerAddress: req.Address,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &registryauthdto.TestRegistryAuthConnResp{}, nil
}
