package sshkeyuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/sshutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sshkeyuc/sshkeydto"
)

func (uc *UC) GenerateSSHKey(
	ctx context.Context,
	auth *basedto.Auth,
	req *sshkeydto.GenerateSSHKeyReq,
) (_ *sshkeydto.GenerateSSHKeyResp, err error) {
	privKey, pubKey, err := sshutil.GenerateKey(req.KeyType, req.Passphrase)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &sshkeydto.GenerateSSHKeyResp{
		Data: &sshkeydto.GenerateSSHKeyDataResp{
			KeyType:    req.KeyType,
			PrivateKey: privKey,
			PublicKey:  pubKey,
		},
	}, nil
}
