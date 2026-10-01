package sessionuc

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/jwtsession"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

func keySessionUser() *basedto.User {
	return &basedto.User{User: &entity.User{ID: "usr_1"}, AuthClaims: &jwtsession.AuthClaims{
		UserID: "usr_1", IsRefresh: true, IsAPIKey: true, APIKeyID: "key_1",
		AccessAction: &base.AccessActions{Read: true},
	}}
}

func keySetting(t *testing.T, owner string, key *entity.APIKey) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: "key_1", Type: base.SettingTypeAPIKey, ObjectID: owner,
		Status: base.SettingStatusActive, ExpireAt: time.Now().Add(time.Hour)}
	assert.NoError(t, setting.SetData(key))
	return setting
}

// The renewal of a session of a key stays a session of that key, as the key is
// now. It once came back a session of the person, with all they may do: a
// read-only key was a refresh away from everything.
func TestARenewedKeySessionKeepsTheKeysLimits(t *testing.T) {
	read := &base.AccessActions{Read: true}
	setting := keySetting(t, "usr_1", &entity.APIKey{KeyID: "k", AccessAction: read,
		Capabilities: []base.ResourceCapability{base.ResourceCapSecretReveal}})

	req := &sessiondto.BaseCreateSessionReq{}
	assert.NoError(t, renewAPIKeySession(req, keySessionUser(), setting))
	assert.True(t, req.IsAPIKey)
	assert.Equal(t, "key_1", req.APIKeyID)
	assert.Equal(t, read, req.AccessAction)
	assert.Equal(t, []base.ResourceCapability{base.ResourceCapSecretReveal}, req.Capabilities)
}

func TestAKeySessionIsNotRenewedWithoutAValidKey(t *testing.T) {
	active := keySetting(t, "usr_1", &entity.APIKey{KeyID: "k"})
	revoked := keySetting(t, "usr_1", &entity.APIKey{KeyID: "k"})
	revoked.Status = base.SettingStatusDisabled
	expired := keySetting(t, "usr_1", &entity.APIKey{KeyID: "k"})
	expired.ExpireAt = time.Now().Add(-time.Minute)
	anothers := keySetting(t, "usr_2", &entity.APIKey{KeyID: "k"})

	for name, setting := range map[string]*entity.Setting{
		"no key: signed in before its id was carried, or deleted": nil,
		"revoked":   revoked,
		"expired":   expired,
		"another's": anothers,
	} {
		err := renewAPIKeySession(&sessiondto.BaseCreateSessionReq{}, keySessionUser(), setting)
		assert.True(t, errors.Is(err, hperrors.ErrAPIKeyInvalid), "%s: %v", name, err)
	}
	assert.NoError(t, renewAPIKeySession(&sessiondto.BaseCreateSessionReq{}, keySessionUser(), active))
}
