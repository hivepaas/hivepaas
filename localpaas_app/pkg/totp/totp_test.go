package totp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTOTPFlow(t *testing.T) {
	// 1. Generate secret and QR code
	secret, qrBuf, err := GenerateSecretAndQRCode(200)
	assert.NoError(t, err)
	assert.NotEmpty(t, secret)
	assert.Greater(t, qrBuf.Len(), 0)

	// 2. Generate a passcode using the secret
	passcode, err := GeneratePasscode(secret)
	assert.NoError(t, err)
	assert.Len(t, passcode, 6) // Standard TOTP is 6 digits

	// 3. Verify the passcode
	isValid := VerifyPasscode(passcode, secret)
	assert.True(t, isValid)

	// 4. Verify with wrong passcode
	assert.False(t, VerifyPasscode("000000", secret))

	// 5. Verify with wrong secret
	// Note: VerifyPasscode uses totp.Validate which might return false for invalid secrets instead of erroring
	assert.False(t, VerifyPasscode(passcode, "JBSWY3DPEHPK3PXP")) // Different valid base32 secret
}

func TestGeneratePasscode_Error(t *testing.T) {
	// Invalid base32 secret (e.g. contains '1', '8', '9', '0' which are not in Base32 alphabet)
	_, err := GeneratePasscode("INVALID123")
	assert.Error(t, err)
}
