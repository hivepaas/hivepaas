package appmanageduc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// fakeFeatures records each feature check, and refuses with err.
type fakeFeatures struct {
	licenseservice.Service
	asked [][]string
	err   error
}

func (f *fakeFeatures) CheckFeatures(_ context.Context, features []string) error {
	f.asked = append(f.asked, features)
	return f.err
}

func managedWith(test entity.ManagedRestoreTest, placement string) *loadedManaged {
	managed := setting("managed-1", base.SettingTypeAppManaged, &entity.AppManagedSettings{Engine: "postgres",
		Postgres: &entity.ManagedPostgres{Storage: entity.ManagedStorage{Placement: placement}, RestoreTest: test}})
	app := &entity.App{ID: "app-1", Settings: []*entity.Setting{managed}}
	return &loadedManaged{app: app, setting: managed, managed: managed.MustAsAppManagedSettings()}
}

var inContainer = entity.ManagedRestoreTest{Mode: entity.ManagedRestoreTestContainer}

var inVerifier = entity.ManagedRestoreTest{Mode: entity.ManagedRestoreTestVerifier,
	VerifierApp: entity.ObjectID{ID: "ver-1"}}

func restoreTestReq(mode string) *appmanageddto.UpdateManagedSettingsReq {
	req := &appmanageddto.UpdateManagedSettingsReq{ProjectID: "p1", ProjectEnvID: "p1:prod", AppID: "app-1",
		UpdateVer: 3, RestoreTest: &appmanageddto.ManagedRestoreTestReq{Mode: mode}}
	if mode == entity.ManagedRestoreTestVerifier {
		req.RestoreTest.Volume = &basedto.ObjectIDReq{ID: "vol-2"}
	}
	return req
}

// A verifier is a paid plan's: the license is asked before anything is read
// or downloaded for it (the UC has nothing else to read with here).
func TestAVerifierNeedsTheFeature(t *testing.T) {
	license := &fakeFeatures{err: hperrors.Wrap(hperrors.ErrPlanFeatureRequired)}
	uc := &UC{licenseService: license}

	_, err := uc.planVerifier(context.Background(), restoreTestReq(entity.ManagedRestoreTestVerifier))

	assert.ErrorIs(t, err, hperrors.ErrPlanFeatureRequired)
	assert.Equal(t, [][]string{{string(base.LicenseFeatureManagedDatabases)}}, license.asked)
}

func TestAVerifierIsRefusedWhereItCannotWork(t *testing.T) {
	for name, tc := range map[string]struct {
		loaded *loadedManaged
		want   error
	}{
		"an archive on a volume": {managedWith(inContainer, entity.ManagedPlacementShared),
			hperrors.ErrAppManagedVerifierNeedsObjectStorage},
		"a verifier already": {managedWith(inVerifier, entity.ManagedPlacementOffSite),
			hperrors.ErrAppManagedVerifierExists},
	} {
		uc := &UC{licenseService: &fakeFeatures{}, appService: &fakeApps{app: tc.loaded.app},
			managedService: &fakeManaged{engine: &fakeEngine{}}}

		_, err := uc.planVerifier(context.Background(), restoreTestReq(entity.ManagedRestoreTestVerifier))

		assert.ErrorIs(t, err, tc.want, name)
	}
}

// Moving back asks no license: a lapsed plan can always leave its verifier.
func TestMovingBackIntoTheContainerDeletesTheVerifier(t *testing.T) {
	license := &fakeFeatures{err: hperrors.Wrap(hperrors.ErrPlanFeatureRequired)}
	uc := &UC{licenseService: license}
	req := restoreTestReq(entity.ManagedRestoreTestContainer)

	plan, err := uc.planVerifier(context.Background(), req)
	assert.NoError(t, err)
	assert.Nil(t, plan)
	assert.Empty(t, license.asked)

	move, deleting, err := uc.restoreTestMove(context.Background(), database.Tx{}, nil,
		managedWith(inVerifier, entity.ManagedPlacementOffSite), req, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "ver-1", deleting)
	if assert.NotNil(t, move) {
		assert.Nil(t, move.Verifier, "back into the container")
	}
}

func TestAMoveToWhereTheTestRunsIsNone(t *testing.T) {
	uc := &UC{}

	move, deleting, err := uc.restoreTestMove(context.Background(), database.Tx{}, nil,
		managedWith(inContainer, entity.ManagedPlacementOffSite), restoreTestReq(entity.ManagedRestoreTestContainer),
		nil, nil)

	assert.NoError(t, err)
	assert.Nil(t, move)
	assert.Empty(t, deleting)
}
