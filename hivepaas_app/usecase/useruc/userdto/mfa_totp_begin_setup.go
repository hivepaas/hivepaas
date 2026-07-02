package userdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type BeginMFATotpSetupReq struct {
	CurrentPasscode string `json:"passcode"`
}

func NewBeginMFATotpSetupReq() *BeginMFATotpSetupReq {
	return &BeginMFATotpSetupReq{}
}

func (req *BeginMFATotpSetupReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStr(&req.CurrentPasscode, false,
		minPasscodeLen, maxPasscodeLen, "passcode")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type BeginMFATotpSetupResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *MFATotpSetupDataResp `json:"data"`
}

type MFATotpSetupDataResp struct {
	Secret    string             `json:"secret"`
	TotpToken string             `json:"totpToken"`
	QRCode    *MFATotpQRCodeResp `json:"qrCode"`
}

type MFATotpQRCodeResp struct {
	DataBase64 string `json:"dataBase64"`
	ImageType  string `json:"imageType"`
	ImageSize  int    `json:"imageSize"`
}
