package useruc

import (
	"context"
	"encoding/base64"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/totp"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
)

const (
	qrCodeImageSize = 300
	qrCodeImageType = "image/png"
)

func (uc *UC) BeginMFATotpSetup(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.BeginMFATotpSetupReq,
) (*userdto.BeginMFATotpSetupResp, error) {
	user, err := uc.userRepo.GetByID(ctx, uc.db, auth.User.ID)
	if err != nil {
		return nil, apperrors.New(err)
	}

	if user.SecurityOption == base.UserSecurityEnforceSSO {
		return nil, apperrors.New(apperrors.ErrActionNotAllowed).
			WithMsgLog("user authentication method is enforce-sso")
	}

	// Verify current passcode if 2FA is enabled on user
	if user.TotpSecret != "" && !totp.VerifyPasscode(req.CurrentPasscode, user.TotpSecret) {
		return nil, apperrors.New(apperrors.ErrPasscodeMismatched)
	}

	secret, qrCode, err := totp.GenerateSecretAndQRCode(qrCodeImageSize)
	if err != nil {
		return nil, apperrors.New(err)
	}

	totpToken, err := uc.userService.GenerateMFATotpSetupToken(user.ID, secret)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &userdto.BeginMFATotpSetupResp{
		Data: &userdto.MFATotpSetupDataResp{
			Secret:    secret,
			TotpToken: totpToken,
			QRCode: &userdto.MFATotpQRCodeResp{
				DataBase64: base64.StdEncoding.EncodeToString(qrCode.Bytes()),
				ImageType:  qrCodeImageType,
				ImageSize:  qrCodeImageSize,
			},
		},
	}, nil
}
