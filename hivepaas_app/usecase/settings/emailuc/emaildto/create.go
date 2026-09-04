package emaildto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	urlMaxLen      = 512
	portMax        = 65535
	usernameMaxLen = 100
	passwordMaxLen = 100
)

type CreateEmailReq struct {
	settings.CreateSettingReq
	*EmailBaseReq
}

type EmailBaseReq struct {
	Name string         `json:"name"`
	Kind base.EmailKind `json:"kind"`
	SMTP *EmailSMTP     `json:"smtp"`
	HTTP *EmailHTTP     `json:"http"`
}

func (req *EmailBaseReq) ToEntity() *entity.Email {
	email := &entity.Email{}
	switch req.Kind {
	case base.EmailKindSMTP:
		email.SMTP = req.SMTP.ToEntity()
	case base.EmailKindHTTP:
		email.HTTP = req.HTTP.ToEntity()
	}
	return email
}

type EmailSMTP struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password"`
	SSL         bool   `json:"ssl"`
}

func (req *EmailSMTP) ToEntity() *entity.EmailSMTP {
	return &entity.EmailSMTP{
		Host:        req.Host,
		Port:        req.Port,
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Password:    entity.NewEncryptedField(req.Password),
		SSL:         req.SSL,
	}
}

func (req *EmailSMTP) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Host, true, 1, urlMaxLen, field+"host")...)
	res = append(res, basedto.ValidateNumber(&req.Port, true, 1, portMax, field+"port")...)
	res = append(res, basedto.ValidateStr(&req.Username, true, 1, usernameMaxLen, field+"username")...)
	res = append(res, basedto.ValidateStr(&req.DisplayName, false, 1, usernameMaxLen, field+"displayName")...)
	res = append(res, basedto.ValidateStr(&req.Password, true, 1, passwordMaxLen, field+"password")...)
	res = append(res, basedto.ValidatePlainSecret(&req.Password, field+"password")...)
	return res
}

type EmailHTTP struct {
	Endpoint     string                        `json:"endpoint"`
	Method       base.HTTPMethod               `json:"method"`
	ContentType  string                        `json:"contentType"`
	Headers      map[string]string             `json:"headers"`
	FieldMapping *entity.EmailHTTPFieldMapping `json:"fieldMapping"` // NOTE: use entity.EmailHTTPFieldMapping directly
	Username     string                        `json:"username"`
	DisplayName  string                        `json:"displayName"`
	Password     string                        `json:"password"`
}

func (req *EmailHTTP) ToEntity() *entity.EmailHTTP {
	return &entity.EmailHTTP{
		Endpoint:     req.Endpoint,
		Method:       req.Method,
		ContentType:  req.ContentType,
		Headers:      req.Headers,
		FieldMapping: req.FieldMapping,
		Username:     req.Username,
		DisplayName:  req.DisplayName,
		Password:     entity.NewEncryptedField(req.Password),
	}
}

func (req *EmailHTTP) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Endpoint, true, 1, urlMaxLen, field+"endpoint")...)
	res = append(res, basedto.ValidateStrIn(&req.Method, true, base.AllHTTPMethods, field+"method")...)
	res = append(res, basedto.ValidatePlainSecret(&req.Password, field+"password")...)
	// TODO: add the remaining validation
	return res
}

func (req *EmailBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	switch req.Kind {
	case base.EmailKindSMTP:
		res = append(res, basedto.ValidateCond(req.SMTP != nil, field+"smtp")...)
		res = append(res, req.SMTP.validate(field+"smtp")...)
	case base.EmailKindHTTP:
		res = append(res, basedto.ValidateCond(req.HTTP != nil, field+"http")...)
		res = append(res, req.HTTP.validate(field+"http")...)
	}
	return res
}

func NewCreateEmailReq() *CreateEmailReq {
	return &CreateEmailReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateEmailReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CreateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateEmailResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
