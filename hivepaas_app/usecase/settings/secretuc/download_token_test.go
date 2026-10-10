package secretuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

// revealGate answers every reveal as told, and keeps what it was asked.
type revealGate struct {
	permission.Manager
	refuse error
	asked  []*permission.RevealSubject
}

func (g *revealGate) AuthorizeSecretReveal(
	_ context.Context, _ database.IDB, _ *basedto.Auth, subject *permission.RevealSubject,
) error {
	g.asked = append(g.asked, subject)
	return g.refuse
}

var appScope = &entity.ObjectScope{ScopeType: base.ObjectScopeApp, AppID: "app_1"}

func ownSecret() *entity.Setting {
	return &entity.Setting{ID: "s1", Name: "TOKEN", ObjectID: "app_1", CurrentObjectID: "app_1"}
}

// A secret downloaded is a secret revealed: it asks the reveal gate, which
// records the asking, about this secret of this app.
func TestDownloadingASecretAsksTheRevealGate(t *testing.T) {
	gate := &revealGate{}
	uc := &UC{BaseUC: &settings.BaseUC{PermissionManager: gate}}

	err := uc.authorizeDownload(context.Background(), &basedto.Auth{}, appScope, ownSecret())

	assert.NoError(t, err)
	if assert.Len(t, gate.asked, 1) {
		assert.Equal(t, "s1", gate.asked[0].ResID)
		assert.Equal(t, "app_1", gate.asked[0].ObjectID)
	}
}

// Refused a reveal - the operator's switch off, or no capability - the secret
// is not downloaded either.
func TestASecretThatMayNotBeRevealedIsNotDownloaded(t *testing.T) {
	uc := &UC{BaseUC: &settings.BaseUC{PermissionManager: &revealGate{refuse: hperrors.ErrRevealSecretsDisabled}}}

	err := uc.authorizeDownload(context.Background(), &basedto.Auth{}, appScope, ownSecret())

	assert.ErrorIs(t, err, hperrors.ErrRevealSecretsDisabled)
}

// An inherited secret belongs to the scope that made it: it is not revealed
// from below, by a download no more than by a reveal.
func TestAnInheritedSecretIsNotDownloadedFromBelow(t *testing.T) {
	gate := &revealGate{}
	uc := &UC{BaseUC: &settings.BaseUC{PermissionManager: gate}}
	inherited := &entity.Setting{ID: "s2", Name: "DB_PASSWORD", ObjectID: "project_1", CurrentObjectID: "app_1"}

	err := uc.authorizeDownload(context.Background(), &basedto.Auth{}, appScope, inherited)

	assert.ErrorIs(t, err, hperrors.ErrUserNotHavePermissionOnRevealSecrets)
}
