package useruc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

// Only a disabled user let back in takes a seat again. Pending and active both
// hold one already, so moving between them, or disabling, takes none.
func TestReenables(t *testing.T) {
	active, pending, disabled := base.UserStatusActive, base.UserStatusPending, base.UserStatusDisabled
	tests := []struct {
		before, after base.UserStatus
		want          bool
	}{
		{before: disabled, after: active, want: true},
		// An admin re-enabling a 2FA user who never set it up leaves them pending.
		{before: disabled, after: pending, want: true},
		{before: disabled, after: disabled},
		{before: active, after: active},
		{before: active, after: pending},
		{before: active, after: disabled},
		{before: pending, after: active},
		{before: pending, after: pending},
		{before: pending, after: disabled},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, reenables(tc.before, tc.after), "%s to %s", tc.before, tc.after)
	}
	// Every status is in the table, both ways.
	for _, status := range base.AllUserStatuses {
		var before, after bool
		for _, tc := range tests {
			before = before || tc.before == status
			after = after || tc.after == status
		}
		assert.True(t, before && after, "status %s is not tested", status)
	}
}

// The users' count is taken under a lock held until the transaction ends, so the
// check has to run in the transaction that writes the user, and the refusal
// has to reach the caller.
func TestCheckUserSeat(t *testing.T) {
	license := &fakeLicenseService{}
	uc := &UC{licenseService: license}
	tx := database.Tx{}

	assert.NoError(t, uc.checkUserSeat(context.Background(), tx))
	if !assert.Len(t, license.checks, 1) {
		return
	}
	assert.Equal(t, tx, license.checks[0].db, "the check must run in the creating transaction")
	assert.Equal(t, map[base.LicenseLimit]int{base.LicenseLimitUsers: 1}, license.checks[0].req.Adding)

	license.err = hperrors.ErrPlanLimitReached
	assert.ErrorIs(t, uc.checkUserSeat(context.Background(), tx), hperrors.ErrPlanLimitReached)
}

type fakeLicenseCheck struct {
	db  database.IDB
	req *licenseservice.CheckReq
}

type fakeLicenseService struct {
	licenseservice.Service
	checks []fakeLicenseCheck
	err    error
}

func (f *fakeLicenseService) Check(_ context.Context, db database.IDB, req *licenseservice.CheckReq) error {
	f.checks = append(f.checks, fakeLicenseCheck{db: db, req: req})
	return f.err
}
