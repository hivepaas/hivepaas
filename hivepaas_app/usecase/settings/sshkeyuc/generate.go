package sshkeyuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/sshutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sshkeyuc/sshkeydto"
)

func (uc *UC) GenerateSSHKey(
	ctx context.Context,
	auth *basedto.Auth,
	req *sshkeydto.GenerateSSHKeyReq,
) (_ *sshkeydto.GenerateSSHKeyResp, err error) {
	privKey, pubKey, err := sshutil.GenerateKey(req.KeyType, req.Passphrase)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &sshkeydto.GenerateSSHKeyResp{
		Data: &sshkeydto.GenerateSSHKeyDataResp{
			KeyType:    req.KeyType,
			PrivateKey: privKey,
			PublicKey:  pubKey,
		},
	}, nil
}
