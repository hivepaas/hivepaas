package appsettingsuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
// private key takes the Reveal Secrets gate, asked once and recorded. Denied,
// the clone goes on without those entries.
func TestCloneRequestLeavesGatedMountsOutWhenDenied(t *testing.T) {
	for name, tc := range map[string]struct {
		err      error
		wantDrop bool
	}{
		"allowed":              {nil, false},
		"no capability":        {hperrors.ErrUserNotHavePermissionOnRevealSecrets, true},
		"secrets not returned": {hperrors.ErrRevealSecretsDisabled, true},
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
