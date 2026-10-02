package registryauthdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	addressMaxLen  = 200
	usernameMaxLen = 100
	// passwordMaxLen fits a Google Artifact Registry service account's JSON
	// key, user _json_key_base64: about 3 KB in base64.
	passwordMaxLen = 8 * 1024

	awsRoleARNMaxLen = 2048
)

type CreateRegistryAuthReq struct {
	settings.CreateSettingReq
	*RegistryAuthBaseReq
}

type RegistryAuthBaseReq struct {
	Name string `json:"name"`
	// Kind is how it signs in: empty for a username and a password, aws-ecr for
	// AWS keys that get Amazon ECR tokens.
	Kind     base.RegistryAuthKind `json:"kind"`
	Address  string                `json:"address"`
	Username string                `json:"username"`
	Password string                `json:"password"`
	Readonly bool                  `json:"readonly"`
	// ECR is the AWS side of an aws-ecr credential; the account and region are
	// read from its address.
	ECR *RegistryAuthECRReq `json:"ecr"`
}

type RegistryAuthECRReq struct {
	// KeyAuth is the key auth holding the AWS access key id and secret key.
	KeyAuth basedto.ObjectIDReq `json:"keyAuth"`
	// RoleARN, when set, is assumed with the keys first.
	RoleARN string `json:"roleArn"`
}

func (req *RegistryAuthBaseReq) ToEntity() *entity.RegistryAuth {
	if req.Kind == base.RegistryAuthKindAWSECR && req.ECR != nil {
		_, region, _ := entity.ParseECRAddress(req.Address)
		return &entity.RegistryAuth{
			Kind:     req.Kind,
			Username: "AWS",
			Address:  req.Address,
			Readonly: req.Readonly,
			ECR: &entity.RegistryAuthECR{
				Region:  region,
				KeyAuth: entity.ObjectID{ID: req.ECR.KeyAuth.ID},
				RoleARN: req.ECR.RoleARN,
			},
		}
	}
	return &entity.RegistryAuth{
		Username: req.Username,
		Password: entity.NewEncryptedField(req.Password),
		Address:  req.Address,
		Readonly: req.Readonly,
	}
}

// SecretFields lists the request's secret values in one place, so the paths that
// care about them do not each restate the list.
func (req *RegistryAuthBaseReq) SecretFields() []basedto.SecretField {
	return []basedto.SecretField{
		{Path: "password", Value: &req.Password},
	}
}

// KeepMaskedSecrets restores the stored values for the secrets the request only
// carries as the masked placeholder the GET response substitutes for them.
func (req *RegistryAuthBaseReq) KeepMaskedSecrets(regAuth, current *entity.RegistryAuth) {
	if current == nil {
		return
	}
	if basedto.IsMaskedSecret(req.Password) {
		regAuth.Password = current.Password
	}
}

// KeepToken carries the ECR token kept in the credential over to its new data
// while it signs in as before; new keys, or a new registry, start without one.
func KeepToken(regAuth, current *entity.RegistryAuth) {
	if current != nil && regAuth.SameECRKeys(current) {
		regAuth.Token = current.Token
		regAuth.TokenExpiresAt = current.TokenExpiresAt
		regAuth.TokenKeyVer = current.TokenKeyVer
	}
}

func (req *RegistryAuthBaseReq) modifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.Address = strings.ToLower(strings.TrimSpace(req.Address))
	req.Username = strings.TrimSpace(req.Username)
	req.Kind = base.RegistryAuthKind(strings.TrimSpace(string(req.Kind)))
	if req.ECR != nil {
		req.ECR.KeyAuth.ID = strings.TrimSpace(req.ECR.KeyAuth.ID)
		req.ECR.RoleARN = strings.TrimSpace(req.ECR.RoleARN)
	}
	return nil
}

func (req *RegistryAuthBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStr(&req.Address, true, 1, addressMaxLen, field+"address")...)
	res = append(res, basedto.ValidateStrIn(&req.Kind, false, base.AllRegistryAuthKinds, field+"kind")...)
	if req.Kind == base.RegistryAuthKindAWSECR {
		return append(res, req.validateECR(field)...)
	}
	res = append(res, basedto.ValidateStr(&req.Username, true, 1, usernameMaxLen, field+"username")...)
	res = append(res, basedto.ValidateStr(&req.Password, true, 1, passwordMaxLen, field+"password")...)
	res = append(res, basedto.ValidatePlainSecret(&req.Password, field+"password")...)
	res = append(res, basedto.ValidateCond(req.ECR == nil, field+"ecr")...)
	return res
}

// validateECR is an aws-ecr credential: an ECR registry's address, the key auth
// holding its AWS keys, and no username or password of its own.
func (req *RegistryAuthBaseReq) validateECR(field string) (res []vld.Validator) {
	_, _, isECR := entity.ParseECRAddress(req.Address)
	res = append(res, basedto.ValidateCond(isECR, field+"address")...)
	res = append(res, basedto.ValidateCond(req.Username == "" || req.Username == "AWS", field+"username")...)
	res = append(res, basedto.ValidateCond(req.Password == "", field+"password")...)
	res = append(res, basedto.ValidateCond(req.ECR != nil, field+"ecr")...)
	if req.ECR == nil {
		return res
	}
	ecr := req.ECR
	res = append(res, basedto.ValidateObjectIDReq(&ecr.KeyAuth, true, field+"ecr.keyAuth")...)
	res = append(res, basedto.ValidateStr(&ecr.RoleARN, false, 0, awsRoleARNMaxLen, field+"ecr.roleArn")...)
	return res
}

func NewCreateRegistryAuthReq() *CreateRegistryAuthReq {
	return &CreateRegistryAuthReq{}
}

func (req *CreateRegistryAuthReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateRegistryAuthReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CreateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	// Creation has no stored value to fall back on, so the placeholder is not a
	// meaningful input here the way it is on update.
	validators = append(validators, basedto.ValidateNoMaskedSecrets(req.SecretFields())...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateRegistryAuthResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
