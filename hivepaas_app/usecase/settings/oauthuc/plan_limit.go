package oauthuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

// checkPaidSignIn refuses a change that adds a paid sign-in provider when the
// plan in force allows none to be added. before is the setting as stored, nil
// when it is being created, and after is the setting as it will be stored.
func (uc *UC) checkPaidSignIn(ctx context.Context, db database.Tx, before, after *entity.Setting) error {
	if !addsPaidSignIn(before, after) {
		return nil
	}
	err := uc.licenseService.Check(ctx, db, licenseservice.Adding(base.LicenseLimitSignInProviders, 1))
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// addsPaidSignIn reports whether a sign-in provider going from before (nil when
// it is being created) to after adds one of a paid kind: creating one, changing
// one to a paid kind from a kind that is not, or enabling one.
//
// Anything else done to a paid one that exists adds nothing. Those keep signing
// people in on every plan, since turning them off would lock out the users who
// have no password, so their settings and secrets stay editable too.
func addsPaidSignIn(before, after *entity.Setting) bool {
	if !license.IsPaidSignIn(base.OAuthKind(after.Kind)) {
		return false
	}
	if before == nil || !license.IsPaidSignIn(base.OAuthKind(before.Kind)) {
		return true
	}
	return after.Status == base.SettingStatusActive && before.Status != base.SettingStatusActive
}
