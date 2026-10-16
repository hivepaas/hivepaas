package hperrors

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
)

// A plan or license refusal is never 403: the dashboard reads 403 as a missing
// permission, and offers no upgrade for one.
func TestLicenseErrorStatuses(t *testing.T) {
	for err, status := range map[error]int{
		ErrPlanLimitReached:         http.StatusPreconditionFailed,
		ErrLicenseInvalid:           http.StatusPreconditionFailed,
		ErrLicenseOtherInstallation: http.StatusPreconditionFailed,
		ErrLicenseExpired:           http.StatusPreconditionFailed,
		ErrLicenseSigningKeyInvalid: http.StatusInternalServerError,
		ErrLicensePortalUnavailable: http.StatusServiceUnavailable,
		ErrLicensePortalRefused:     http.StatusPreconditionFailed,
	} {
		assert.Equal(t, status, Wrap(err).StatusCode(), err)
	}
}

func TestLicenseOtherInstallationNamesThisInstallation(t *testing.T) {
	msg, err := Wrap(ErrLicenseOtherInstallation).WithParam("InstallationID", "x2kq7").Message(translation.LangEn)
	assert.NoError(t, err)
	assert.Equal(t, "This license key is for another installation. This installation's ID is x2kq7.", msg)

	msg, err = Wrap(ErrLicenseExpired).Message(translation.LangEn)
	assert.NoError(t, err)
	assert.Equal(t, "This license key has expired", msg)
}
