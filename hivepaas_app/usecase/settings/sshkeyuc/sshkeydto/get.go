package sshkeydto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	maskedSecret = "****************"
)

type GetSSHKeyReq struct {
	settings.GetSettingReq
}

func NewGetSSHKeyReq() *GetSSHKeyReq {
	return &GetSSHKeyReq{}
}

func (req *GetSSHKeyReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetSSHKeyResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *SSHKeyResp   `json:"data"`
}

type SSHKeyResp struct {
	*settings.BaseSettingResp
	KeyType      base.PrivateKeyType `json:"keyType"`
	PublicKey    string              `json:"publicKey,omitempty"`
	PrivateKey   string              `json:"privateKey"`
	Passphrase   string              `json:"passphrase,omitempty"`
	SecretMasked bool                `json:"secretMasked,omitempty"`
}

func (resp *SSHKeyResp) CopyPrivateKey(field entity.EncryptedField) error {
	resp.PrivateKey = field.String()
	return nil
}

func (resp *SSHKeyResp) CopyPassphrase(field entity.EncryptedField) error {
	resp.Passphrase = field.String()
	return nil
}

func TransformSSHKey(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *SSHKeyResp, err error) {
	sshKey := setting.MustAsSSHKey()
	if err = copier.Copy(&resp, &sshKey); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.SecretMasked = sshKey.PrivateKey.IsEncrypted() || resp.Inherited
	if resp.SecretMasked {
		resp.PrivateKey = maskedSecret
		if resp.Passphrase != "" {
			resp.Passphrase = maskedSecret
		}
	}

	return resp, nil
}
