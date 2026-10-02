package registryauthserviceimpl

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
)

var now = time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)

// TestMain gives the tests a data key: a credential's secrets and token are
// encrypted with it.
func TestMain(m *testing.M) {
	key, err := datakey.Generate()
	if err != nil {
		panic(err)
	}
	datakey.SetActive(key)
	os.Exit(m.Run())
}

// fakeECR answers tokens tok-1, tok-2... living 12 hours, or an error.
type fakeECR struct {
	calls int
	err   error
}

func (f *fakeECR) Token(_ context.Context, auth *entity.RegistryAuthECR, secret string) (*ecrToken, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	if auth.AccessKeyID != "AKIAEXAMPLE000000000" || secret != "s3cret" {
		return nil, errors.New("UnrecognizedClientException: the security token included in the request is invalid")
	}
	return &ecrToken{password: "tok-" + string(rune('0'+f.calls)), expiresAt: now.Add(12 * time.Hour)}, nil
}

// settingRow is the credential's row: what GetByID reads, what Update writes.
type settingRow struct {
	repository.SettingRepo
	row     *entity.Setting
	locked  bool
	updates []string
}

func (r *settingRow) GetByID(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ base.SettingType,
	_ string, _ bool, opts ...bunex.SelectQueryOption) (*entity.Setting, error) {
	r.locked = len(opts) > 0
	copied := *r.row
	return &copied, nil
}

func (r *settingRow) Update(_ context.Context, _ database.IDB, setting *entity.Setting,
	_ ...bunex.UpdateQueryOption) error {
	r.updates = append(r.updates, setting.Data)
	r.row = setting
	return nil
}

func ecrSetting(t *testing.T, token string, expiresAt time.Time) *entity.Setting {
	t.Helper()
	auth := &entity.RegistryAuth{
		Kind: base.RegistryAuthKindAWSECR, Username: "AWS", Address: "123456789012.dkr.ecr.eu-west-1.amazonaws.com",
		ECR: &entity.RegistryAuthECR{Region: "eu-west-1", AccessKeyID: "AKIAEXAMPLE000000000",
			SecretAccessKey: entity.NewEncryptedField("s3cret")},
		TokenExpiresAt: expiresAt,
	}
	if token != "" {
		auth.Token = entity.NewEncryptedField(token)
	}
	setting := &entity.Setting{ID: "ra1", Type: base.SettingTypeRegistryAuth, UpdateVer: 4}
	setting.MustSetData(auth)
	return setting
}

func newTestService(row *settingRow, ecr *fakeECR) *service {
	return &service{settingRepo: row, ecr: ecr, now: func() time.Time { return now },
		inTx:     func(ctx context.Context, fn func(tx database.Tx) error) error { return fn(database.Tx{}) },
		interval: func(context.Context) time.Duration { return entity.RegistryAuthRenewalIntervalDefault }}
}

// A token kept while it lives the renewal's interval and an hour more is used
// as it is: no AWS call, nothing written.
func TestAFreshTokenIsUsedAsKept(t *testing.T) {
	setting := ecrSetting(t, "kept", now.Add(8*time.Hour))
	row, ecr := &settingRow{row: setting}, &fakeECR{}
	auth, err := newTestService(row, ecr).AuthConfig(context.Background(), setting)
	assert.NoError(t, err)
	assert.Equal(t, "AWS", auth.Username)
	assert.Equal(t, "kept", auth.Password)
	assert.Equal(t, "123456789012.dkr.ecr.eu-west-1.amazonaws.com", auth.ServerAddress)
	assert.Zero(t, ecr.calls)
	assert.Empty(t, row.updates)
}

// One that would expire before the next renewal and its hour of margin is got
// again, under the row's lock, and kept - its version left as it was.
func TestATokenTooOldToHandOverIsGotAgainAndKept(t *testing.T) {
	setting := ecrSetting(t, "old", now.Add(6*time.Hour+59*time.Minute))
	row, ecr := &settingRow{row: setting}, &fakeECR{}
	svc := newTestService(row, ecr)
	auth, err := svc.AuthConfig(context.Background(), setting)
	assert.NoError(t, err)
	assert.Equal(t, "tok-1", auth.Password)
	assert.Equal(t, 1, ecr.calls)
	assert.True(t, row.locked, "the row is read FOR UPDATE")
	if assert.Len(t, row.updates, 1) {
		kept := row.row.MustAsRegistryAuth()
		token, _ := kept.Token.GetPlain()
		assert.Equal(t, "tok-1", token)
		assert.Equal(t, now.Add(12*time.Hour), kept.TokenExpiresAt)
		assert.Equal(t, 4, row.row.UpdateVer, "the version is the person's, not the token's")
		assert.NotContains(t, row.updates[0], "tok-1", "kept encrypted")
	}
}

// Another process that got one while this one waited for the lock: its token
// is used, AWS is not asked again.
func TestATokenGotByAnotherWhileWaitingIsUsed(t *testing.T) {
	stale := ecrSetting(t, "", time.Time{})
	row, ecr := &settingRow{row: ecrSetting(t, "theirs", now.Add(12*time.Hour))}, &fakeECR{}
	auth, err := newTestService(row, ecr).AuthConfig(context.Background(), stale)
	assert.NoError(t, err)
	assert.Equal(t, "theirs", auth.Password)
	assert.Zero(t, ecr.calls)
	assert.Empty(t, row.updates)
}

// Keys AWS refuses are the caller's error, naming the registry.
func TestKeysAWSRefusesAreWorded(t *testing.T) {
	setting := ecrSetting(t, "", time.Time{})
	auth := setting.MustAsRegistryAuth()
	auth.ECR.SecretAccessKey = entity.NewEncryptedField("wrong")
	setting.MustSetData(auth)
	row := &settingRow{row: setting}
	_, err := newTestService(row, &fakeECR{}).AuthConfig(context.Background(), setting)
	info, _ := hperrors.ParseError(err, translation.LangEn)
	assert.Contains(t, info.Detail, "123456789012.dkr.ecr.eu-west-1.amazonaws.com were refused")
	assert.Contains(t, info.Detail, "security token included in the request is invalid")
	assert.Empty(t, row.updates)
}

// A credential being tested gets a token, kept nowhere.
func TestTryingACredentialKeepsNothing(t *testing.T) {
	row, ecr := &settingRow{}, &fakeECR{}
	unsaved := ecrSetting(t, "", time.Time{}).MustAsRegistryAuth()
	auth, err := newTestService(row, ecr).TryAuth(context.Background(), unsaved)
	assert.NoError(t, err)
	assert.Equal(t, "tok-1", auth.Password)
	assert.Empty(t, row.updates)
}

// A username and a password are answered as they are.
func TestABasicCredentialIsAnsweredAsStored(t *testing.T) {
	setting := &entity.Setting{ID: "ra2", Type: base.SettingTypeRegistryAuth}
	setting.MustSetData(&entity.RegistryAuth{Username: "bot", Password: entity.NewEncryptedField("pw"),
		Address: "ghcr.io"})
	ecr := &fakeECR{}
	header, err := newTestService(&settingRow{}, ecr).AuthHeader(context.Background(), setting)
	assert.NoError(t, err)
	assert.NotEmpty(t, header)
	assert.Zero(t, ecr.calls)
}

func TestAnECRTokenIsDecoded(t *testing.T) {
	password, err := decodeECRToken(base64.StdEncoding.EncodeToString([]byte("AWS:eyJwYXlsb2FkIjoi")))
	assert.NoError(t, err)
	assert.Equal(t, "eyJwYXlsb2FkIjoi", password)
	for _, bad := range []string{"not base64!", base64.StdEncoding.EncodeToString([]byte("bob:pw")),
		base64.StdEncoding.EncodeToString([]byte("AWS:"))} {
		_, err = decodeECRToken(bad)
		assert.ErrorIs(t, err, errMalformedToken, bad)
	}
}
