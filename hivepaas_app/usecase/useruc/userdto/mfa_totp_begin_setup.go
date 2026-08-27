package userdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type BeginMFATotpSetupReq struct {
	CurrentPasscode string `json:"passcode"`
}

func NewBeginMFATotpSetupReq() *BeginMFATotpSetupReq {
	return &BeginMFATotpSetupReq{}
}

func (req *BeginMFATotpSetupReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.CurrentPasscode, false,
		minPasscodeLen, maxPasscodeLen, "passcode")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
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
