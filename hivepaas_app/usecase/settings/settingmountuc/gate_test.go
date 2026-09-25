package settingmountuc

import (
	"context"
	"encoding/json"
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

type fakePermissions struct {
	permission.Manager
	err      error
	subjects []*permission.RevealSubject
}

func (f *fakePermissions) AuthorizeSecretReveal(
	_ context.Context, _ database.IDB, _ *basedto.Auth, subject *permission.RevealSubject,
) error {
	f.subjects = append(f.subjects, subject)
	return f.err
}

func entrySetting(t *testing.T, status base.SettingStatus, source string, parts ...string) *entity.Setting {
	t.Helper()
	mount := &entity.AppSettingMount{Source: entity.ObjectID{ID: source}}
	for _, part := range parts {
		mount.Files = append(mount.Files, &entity.AppSettingMountFile{Part: part, Path: "/etc/" + part})
	}
	setting := &entity.Setting{ID: "m1", Name: "cert", Type: base.SettingTypeAppSettingMount, Status: status}
	assert.NoError(t, setting.SetData(mount))
	return setting
}

// A disabled entry hands nothing out, so enabling one with a private key is
// growth, and asks.
func TestStatusGrantsCountOnlyWhenActive(t *testing.T) {
	disabled := entrySetting(t, base.SettingStatusDisabled, "cert_1", "certificate", "privateKey")
	enabled := entrySetting(t, base.SettingStatusActive, "cert_1", "certificate", "privateKey")

	assert.Empty(t, activeGrants(disabled))
	assert.Len(t, activeGrants(enabled), 1)
}

func TestTheGateIsAskedOnlyForWhatIsAdded(t *testing.T) {
	perms := &fakePermissions{}
	uc := &UC{BaseUC: &settings.BaseUC{PermissionManager: perms}}
	scope := &entity.ObjectScope{ScopeType: base.ObjectScopeApp, AppID: "app_1"}
	before := entrySetting(t, base.SettingStatusActive, "cert_1", "privateKey")

	assert.NoError(t, uc.authorizeGrants(context.Background(), nil, scope, before, before,
		base.AuditLogSourceAPIUpdate))
	assert.Empty(t, perms.subjects, "unchanged grants ask nothing")

	after := entrySetting(t, base.SettingStatusActive, "cert_2", "privateKey")
	assert.NoError(t, uc.authorizeGrants(context.Background(), nil, scope, before, after,
		base.AuditLogSourceAPIUpdate))
	if assert.Len(t, perms.subjects, 1) {
		subject := perms.subjects[0]
		assert.Equal(t, "app_1", subject.ObjectID)
		assert.Equal(t, base.ResourceTypeSettingMount, subject.ResType)
		assert.Equal(t, "cert", subject.ResName)
		var detail map[string]any
		assert.NoError(t, json.Unmarshal([]byte(subject.Detail), &detail))
		assert.Equal(t, []any{map[string]any{"source": "cert_2", "part": "privateKey"}}, detail["grants"])
	}

	perms.err = hperrors.ErrUserNotHavePermissionOnRevealSecrets
	err := uc.authorizeGrants(context.Background(), nil, scope, nil, after, base.AuditLogSourceAPICreate)
	assert.ErrorIs(t, err, hperrors.ErrUserNotHavePermissionOnRevealSecrets)
}
