package registryauthdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetRegistryAuthReq struct {
	settings.GetSettingReq
}

func NewGetRegistryAuthReq() *GetRegistryAuthReq {
	return &GetRegistryAuthReq{}
}

func (req *GetRegistryAuthReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetRegistryAuthResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *RegistryAuthResp `json:"data"`
}

type RegistryAuthResp struct {
	*settings.BaseSettingResp
	Kind         base.RegistryAuthKind `json:"kind"`
	Address      string                `json:"address"`
	Username     string                `json:"username"`
	Password     string                `json:"password"`
	Readonly     bool                  `json:"readonly"`
	SecretMasked bool                  `json:"secretMasked,omitempty"`
	// AWSECR is an aws-ecr credential's keys. Named apart from the entity's ECR
	// so the copier leaves it to TransformRegistryAuth.
	AWSECR *RegistryAuthECRResp `json:"ecr,omitempty"`
}

// RegistryAuthECRResp is an aws-ecr credential's AWS side, and when the token
// kept for it expires - never the token.
type RegistryAuthECRResp struct {
	Region string `json:"region"`
	// KeyAuth is the key auth holding the AWS keys: its id and name, or its id
	// alone and missing when it no longer exists. Never its keys, which are the
	// key auth's to reveal.
	KeyAuth        *settings.BaseSettingResp `json:"keyAuth"`
	RoleARN        string                    `json:"roleArn,omitempty"`
	TokenExpiresAt time.Time                 `json:"tokenExpiresAt,omitzero"`
}

// keyAuthOf is the linked key auth as a reference.
func keyAuthOf(id string, refObjects *entity.RefObjects) (*settings.BaseSettingResp, error) {
	if id == "" {
		return nil, nil
	}
	var keyAuth *entity.Setting
	if refObjects != nil {
		keyAuth = refObjects.RefSettings[id]
	}
	if keyAuth == nil {
		return settings.NewMissingSetting(id, base.SettingTypeKeyAuth), nil
	}
	resp, err := settings.TransformSettingBase(keyAuth)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}

func (resp *RegistryAuthResp) CopyPassword(field entity.EncryptedField) error {
	resp.Password = field.String()
	return nil
}

func TransformRegistryAuth(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (resp *RegistryAuthResp, err error) {
	config := setting.MustAsRegistryAuth()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if config.ECR != nil {
		resp.AWSECR = &RegistryAuthECRResp{
			Region:         config.ECR.Region,
			RoleARN:        config.ECR.RoleARN,
			TokenExpiresAt: config.TokenExpiresAt,
		}
		resp.AWSECR.KeyAuth, err = keyAuthOf(config.ECR.KeyAuth.ID, refObjects)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	resp.SecretMasked = config.Password.IsEncrypted() || resp.Inherited
	if resp.SecretMasked && config.Kind != base.RegistryAuthKindAWSECR {
		resp.Password = basedto.MaskedSecret
	}

	return resp, nil
}
