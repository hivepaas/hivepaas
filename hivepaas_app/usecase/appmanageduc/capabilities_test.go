package appmanageduc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// lackingUC is a managed app whose engine has neither replicas nor a verifier
// app.
func lackingUC() *UC {
	return &UC{licenseService: &fakeFeatures{}, appService: &fakeApps{app: managedApp()},
		managedService: &fakeManaged{engine: &fakeEngine{lacking: true}}}
}

// What an engine cannot do is refused before anything is read or made for it,
// and the person is told which engine lacks what.
func TestAnEngineWithoutReplicasRefusesThem(t *testing.T) {
	ctx := context.Background()
	volume := &basedto.ObjectIDReq{ID: "vol-2"}

	_, err := lackingUC().planReplica(ctx, &appmanageddto.AddManagedReplicaReq{ProjectID: "p1",
		ProjectEnvID: "p1:prod", AppID: "app-1", DataVolume: volume})
	assert.ErrorIs(t, err, hperrors.ErrAppManagedUnsupportedByEngine, "adding one")
	info, _ := hperrors.ParseError(err, translation.LangEn)
	assert.Equal(t, "PostgreSQL databases do not have replicas yet", info.Detail)

	err = lackingUC().removeReplica(ctx, database.Tx{}, nil, &appmanageddto.RemoveManagedReplicaReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", AppID: "app-1", MemberID: "r1"})
	assert.ErrorIs(t, err, hperrors.ErrAppManagedUnsupportedByEngine, "removing one")

	_, err = lackingUC().planPromotion(ctx, &appmanageddto.PromoteManagedMemberReq{ProjectID: "p1",
		ProjectEnvID: "p1:prod", AppID: "app-1", Member: &basedto.ObjectIDReq{ID: "r1"}})
	assert.ErrorIs(t, err, hperrors.ErrAppManagedUnsupportedByEngine, "promoting one")

	_, err = lackingUC().updateManagedSettings(ctx, database.Tx{}, nil, &appmanageddto.UpdateManagedSettingsReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", AppID: "app-1", UpdateVer: 3,
		BackupFrom: &basedto.ObjectIDReq{ID: "r1"}}, nil)
	assert.ErrorIs(t, err, hperrors.ErrAppManagedUnsupportedByEngine, "taking the base backups on one")
}

// Taking the base backups on the database itself is all such an engine does:
// asking for it is no refusal.
func TestAnEngineWithoutReplicasTakesItsBackupsHome(t *testing.T) {
	loaded, err := lackingUC().loadManaged(context.Background(), nil, "p1", "app-1")
	assert.NoError(t, err)

	assert.NoError(t, loaded.checkBackupFrom(nil))
	assert.NoError(t, loaded.checkBackupFrom(&basedto.ObjectIDReq{ID: ""}))
	assert.ErrorIs(t, loaded.checkBackupFrom(&basedto.ObjectIDReq{ID: "r1"}),
		hperrors.ErrAppManagedUnsupportedByEngine)
}

func TestAnEngineWithoutAVerifierAppRefusesOne(t *testing.T) {
	_, err := lackingUC().planVerifier(context.Background(), restoreTestReq(entity.ManagedRestoreTestVerifier))
	assert.ErrorIs(t, err, hperrors.ErrAppManagedUnsupportedByEngine)
	info, _ := hperrors.ParseError(err, translation.LangEn)
	assert.Equal(t, "PostgreSQL databases do not have verifier apps yet", info.Detail)

	// The restore test in the container is every engine's.
	plan, err := lackingUC().planVerifier(context.Background(), restoreTestReq(entity.ManagedRestoreTestContainer))
	assert.NoError(t, err)
	assert.Nil(t, plan)
}
