package useruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

// checkUserSeat refuses one more user past the plan's limit. A user active or
// pending holds a seat, so it is called when a person is invited for the first
// time and when a disabled user is let back in.
//
// The users' plan-limit lock it takes is held until tx ends, so tx must be the
// transaction that writes the user: a second invite waits for this one to
// commit, and counts it.
func (uc *UC) checkUserSeat(ctx context.Context, tx database.Tx) error {
	err := uc.licenseService.Check(ctx, tx, licenseservice.Adding(base.LicenseLimitUsers, 1))
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// reenables reports whether an update takes a user from disabled to a status that
// holds a seat again. Whatever the update does to the status of a user who is
// not disabled, the seat is the one already held.
func reenables(before, after base.UserStatus) bool {
	return before == base.UserStatusDisabled && after != base.UserStatusDisabled
}
