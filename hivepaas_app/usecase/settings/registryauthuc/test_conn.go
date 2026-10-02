package registryauthuc

import (
	"context"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/registryauthuc/registryauthdto"
)

func (uc *UC) TestRegistryAuthConn(
	ctx context.Context,
	auth *basedto.Auth,
	req *registryauthdto.TestRegistryAuthConnReq,
) (*registryauthdto.TestRegistryAuthConnResp, error) {
	regAuth := req.ToEntity()
	// The keys are a key auth's: one the caller cannot read is not tried for
	// them, nor told whether it signs in.
	if regAuth.ECR != nil {
		keyAuth, err := uc.keyAuthOf(ctx, uc.DB, regAuth)
		if err != nil {
			return nil, err
		}
		if err = uc.CheckAccessOnSetting(ctx, uc.DB, auth, keyAuth, base.ActionTypeRead); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	// An ECR credential signs in with a token got for its keys now, kept
	// nowhere: the credential being tested may not be saved.
	login, err := uc.registryAuthService.TryAuth(ctx, regAuth)
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
