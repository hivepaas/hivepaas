package registryauthuc

import (
	"context"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/registryauthuc/registryauthdto"
)

func (uc *UC) TestRegistryAuthConn(
	ctx context.Context,
	auth *basedto.Auth,
	req *registryauthdto.TestRegistryAuthConnReq,
) (*registryauthdto.TestRegistryAuthConnResp, error) {
	// An ECR credential signs in with a token got for its keys now, kept
	// nowhere: the credential being tested may not be saved.
	login, err := uc.registryAuthService.TryAuth(ctx, req.ToEntity())
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	_, err = uc.dockerManager.RegistryLogin(ctx, func(opts *client.RegistryLoginOptions) {
		opts.Username = login.Username
		opts.Password = login.Password
		opts.ServerAddress = login.ServerAddress
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &registryauthdto.TestRegistryAuthConnResp{}, nil
}
