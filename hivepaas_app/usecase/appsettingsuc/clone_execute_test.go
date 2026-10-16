package appsettingsuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

func inheritableEntry(t *testing.T, name, source string, parts ...string) *entity.Setting {
	t.Helper()
	mount := &entity.AppSettingMount{Source: entity.ObjectID{ID: source}}
	for _, part := range parts {
		mount.Files = append(mount.Files, &entity.AppSettingMountFile{Part: part, Path: "/etc/" + name + "/" + part})
	}
	setting := &entity.Setting{ID: "m-" + name, Type: base.SettingTypeAppSettingMount, Name: name,
		Status: base.SettingStatusActive, Inheritable: true}
	assert.NoError(t, setting.SetData(mount))
	return setting
}

// The clone would hand its requester an app that reads the entries' files: a
// private key takes the Reveal Secrets capability, asked once and recorded. Denied,
// the clone goes on without those entries.
func TestCloneRequestLeavesGatedMountsOutWhenDenied(t *testing.T) {
	for name, tc := range map[string]struct {
		err      error
		wantDrop bool
	}{
		"allowed":       {nil, false},
		"no capability": {hperrors.ErrUserNotHavePermissionOnRevealSecrets, true},
	} {
		gate := &revealGate{err: tc.err}
		uc := &UC{permissionManager: gate}
		entries := []*entity.Setting{
			inheritableEntry(t, "tls", "cert_1", "certificate", "privateKey"),
			inheritableEntry(t, "conf", "cfg_1", "content"),
		}

		drop, leftOut, err := uc.cloneMountsGate(context.Background(), nil, &entity.App{ID: "app_1"}, entries)

		assert.NoError(t, err, name)
		assert.Equal(t, tc.wantDrop, drop, name)
		if tc.wantDrop {
			assert.Equal(t, []string{"tls"}, leftOut, name)
		}
		assert.Len(t, gate.subjects, 1, "asked once, recorded: %s", name)
	}
}

func TestCloneRequestAsksNothingWithoutGatedParts(t *testing.T) {
	gate := &revealGate{}
	uc := &UC{permissionManager: gate}

	drop, _, err := uc.cloneMountsGate(context.Background(), nil, &entity.App{ID: "app_1"},
		[]*entity.Setting{inheritableEntry(t, "conf", "cfg_1", "content")})

	assert.NoError(t, err)
	assert.False(t, drop)
	assert.Empty(t, gate.subjects)
}

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

// The request is checked against what the clone will be, in its transaction, for
// an answer on screen: a function's clone that takes its code is a function, any
// other clone an app. The clone task checks again when it makes the app.
func TestCloneRequestChecksWhatTheCloneWillBe(t *testing.T) {
	for name, tc := range map[string]struct {
		category   base.AppCategory
		withSource bool
		want       base.LicenseLimit
	}{
		"a function, with its code":    {base.AppCategoryFunction, true, base.LicenseLimitFunctions},
		"a function, without its code": {base.AppCategoryFunction, false, base.LicenseLimitApps},
		"a database, with its source":  {base.AppCategoryDatabase, true, base.LicenseLimitApps},
		"an app with no kind":          {"", true, base.LicenseLimitApps},
	} {
		app := &entity.App{ID: "app_1"}
		if tc.category != "" {
			kind := &entity.Setting{ID: "kind_1", Type: base.SettingTypeAppKind, ObjectID: app.ID}
			assert.NoError(t, kind.SetData(&entity.AppKindSettings{Category: tc.category}), name)
			app.Settings = []*entity.Setting{kind}
		}
		licenses := &fakeLicenseService{}
		uc := &UC{licenseService: licenses}

		err := uc.checkClonePlanLimit(context.Background(), database.Tx{}, app,
			&entity.AppCloneSettings{CloneDeploymentSettings: tc.withSource})

		assert.NoError(t, err, name)
		if assert.Len(t, licenses.reqs, 1, name) {
			assert.IsType(t, database.Tx{}, licenses.dbs[0], name)
			assert.Equal(t, licenseservice.Adding(tc.want, 1), licenses.reqs[0], name)
		}
	}
}

func TestCloneRequestPastThePlanIsRefused(t *testing.T) {
	uc := &UC{licenseService: &fakeLicenseService{err: hperrors.Wrap(hperrors.ErrPlanLimitReached)}}

	err := uc.checkClonePlanLimit(context.Background(), database.Tx{}, &entity.App{ID: "app_1"},
		&entity.AppCloneSettings{})

	assert.ErrorIs(t, err, hperrors.ErrPlanLimitReached)
}
