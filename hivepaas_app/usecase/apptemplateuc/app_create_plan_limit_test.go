package apptemplateuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

// fakeLicenseService records each plan check, and refuses with err; and each
// feature check, refused with featuresErr. Its license is current, or Bee.
type fakeLicenseService struct {
	licenseservice.Service
	dbs  []database.IDB
	reqs []*licenseservice.CheckReq
	err  error

	features    [][]string
	featuresErr error
	current     *license.State
}

func (f *fakeLicenseService) CheckFeatures(_ context.Context, features []string) error {
	f.features = append(f.features, features)
	return f.featuresErr
}

func (f *fakeLicenseService) Current(context.Context) (*license.State, error) {
	if f.current != nil {
		return f.current, nil
	}
	return &license.State{Plan: base.LicensePlanBee}, nil
}

func (f *fakeLicenseService) Check(_ context.Context, db database.IDB, req *licenseservice.CheckReq) error {
	f.dbs = append(f.dbs, db)
	f.reqs = append(f.reqs, req)
	return f.err
}

// A template that creates a database and a component beside the app it is asked
// for adds one app to the plan's count: the others belong to that app, and app
// lists leave them out too. It is checked on the creating transaction, before
// anything is provisioned.
func TestATemplateCountsTheAppItIsAskedFor(t *testing.T) {
	uc, fakes := newCreateTest(t)
	req := testCreateReq()
	req.Name = "shop"
	apps := planApps(req, renderedWithComponents(t, fakes))

	_, err := uc.provisionAll(context.Background(), database.Tx{}, testAuth(), req, apps, nil)

	assert.NoError(t, err)
	assert.Len(t, fakes.provision.names, 3, "a database, a component and the app")
	if assert.Len(t, fakes.licenses.reqs, 1) {
		assert.IsType(t, database.Tx{}, fakes.licenses.dbs[0])
		assert.Equal(t, licenseservice.Adding(base.LicenseLimitApps, 1), fakes.licenses.reqs[0])
	}
}

func TestATemplatePastThePlanProvisionsNothing(t *testing.T) {
	uc, fakes := newCreateTest(t)
	fakes.licenses.err = hperrors.Wrap(hperrors.ErrPlanLimitReached)
	req := testCreateReq()

	_, err := uc.provisionAll(context.Background(), database.Tx{}, testAuth(), req,
		planApps(req, fakes.templates.resp), nil)

	assert.ErrorIs(t, err, hperrors.ErrPlanLimitReached)
	assert.False(t, fakes.provision.called)
}
