package totp

import (
	"bytes"
	"image/png"
	"time"

	"github.com/pquerna/otp/totp"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tracerr"
)

func GenerateSecretAndQRCode(imageSize int) (string, bytes.Buffer, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "HivePaaS",
		AccountName: "HivePaaS",
	})
	if err != nil {
		return "", bytes.Buffer{}, tracerr.Wrap(err)
	}
	qrCode, err := key.Image(imageSize, imageSize)
	if err != nil {
		return "", bytes.Buffer{}, tracerr.Wrap(err)
	}
	buf := bytes.Buffer{}
	err = png.Encode(&buf, qrCode)
	if err != nil {
		return "", bytes.Buffer{}, tracerr.Wrap(err)
	}
	return key.Secret(), buf, nil
}

func GeneratePasscode(secret string) (string, error) {
	code, err := totp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		return "", tracerr.Wrap(err)
	}
	return code, nil
}

func VerifyPasscode(passcode, secret string) bool {
	return totp.Validate(passcode, secret)
}
