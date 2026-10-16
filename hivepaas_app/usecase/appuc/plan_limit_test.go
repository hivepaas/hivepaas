package appuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

// fakeLicenseService records each plan check, and refuses with err.
type fakeLicenseService struct {
	licenseservice.Service
	dbs  []database.IDB
	reqs []*licenseservice.CheckReq
	err  error
}

func (f *fakeLicenseService) Check(_ context.Context, db database.IDB, req *licenseservice.CheckReq) error {
	f.dbs = append(f.dbs, db)
	f.reqs = append(f.reqs, req)
	return f.err
}

// Creating an app or a function asks for one more of its kind, on the creating
// transaction: the plan-limit lock the check takes lasts until the creation
// commits.
func TestCreatingChecksOneMoreOfItsKindInItsTransaction(t *testing.T) {
	for _, limit := range []base.LicenseLimit{base.LicenseLimitApps, base.LicenseLimitFunctions} {
		licenses := &fakeLicenseService{}
		uc := &UC{licenseService: licenses}

		err := uc.checkPlanLimit(context.Background(), database.Tx{}, limit)

		assert.NoError(t, err, limit)
		if assert.Len(t, licenses.reqs, 1, limit) {
			assert.IsType(t, database.Tx{}, licenses.dbs[0], limit)
			assert.Equal(t, licenseservice.Adding(limit, 1), licenses.reqs[0], limit)
		}
	}
}

func TestCreatingPastThePlanIsRefused(t *testing.T) {
	uc := &UC{licenseService: &fakeLicenseService{err: hperrors.Wrap(hperrors.ErrPlanLimitReached)}}

	err := uc.checkPlanLimit(context.Background(), database.Tx{}, base.LicenseLimitApps)

	assert.ErrorIs(t, err, hperrors.ErrPlanLimitReached)
}
