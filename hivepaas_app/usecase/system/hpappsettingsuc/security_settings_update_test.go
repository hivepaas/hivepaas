package hpappsettingsuc

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

const testAppSecret = "the-real-app-secret"

// spyAuditService keeps what was recorded, so a test can assert on the record
// itself rather than only on the answer the caller got.
type spyAuditService struct {
	entries []*auditservice.Entry
	err     error
}

func (s *spyAuditService) Record(_ context.Context, _ database.IDB, entry *auditservice.Entry) error {
	s.entries = append(s.entries, entry)
	return s.err
}

// spyHpAppService embeds the interface so only the one method under test needs a
// body; anything else this code path touches would panic loudly rather than pass.
type spyHpAppService struct {
	hpappservice.Service
	reloads int
	err     error
}

func (s *spyHpAppService) ReloadHpAppConfig(_ context.Context) error {
	s.reloads++
	return s.err
}

// fakeAppSecretAttemptRepo stands in for redis. A missing entry answers
// ErrNotFound, the way the real one does.
type fakeAppSecretAttemptRepo struct {
	attempt *cacheentity.AppSecretAttempt
	getErr  error
	dels    int
}

func (f *fakeAppSecretAttemptRepo) Get(
	_ context.Context, _ string,
) (*cacheentity.AppSecretAttempt, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if f.attempt == nil {
		return nil, hperrors.ErrNotFound
	}
	return f.attempt, nil
}

func (f *fakeAppSecretAttemptRepo) Set(
	_ context.Context, _ string, attempt *cacheentity.AppSecretAttempt, _ time.Duration,
) error {
	f.attempt = attempt
	return nil
}

func (f *fakeAppSecretAttemptRepo) Del(_ context.Context, _ string) error {
	f.dels++
	f.attempt = nil
	return nil
}

func newSecurityUCTest(t *testing.T) (*UC, *spyAuditService, *spyHpAppService) {
	t.Helper()
	cfg := &config.Config{
		Env:     config.EnvDev,
		AppPath: t.TempDir(),
		Secret:  testAppSecret,
	}
	cfg.Users.Demo.UserID = "demo-user"
	config.SetCurrent(cfg)

	audit := &spyAuditService{}
	hpApp := &spyHpAppService{}
	uc := &UC{
		auditService:              audit,
		hpAppService:              hpApp,
		cacheAppSecretAttemptRepo: &fakeAppSecretAttemptRepo{},
	}
	return uc, audit, hpApp
}

// attemptsOf reaches the fake back out of the UC, so a test can look at what the
// backoff actually recorded.
func attemptsOf(uc *UC) *fakeAppSecretAttemptRepo {
	return uc.cacheAppSecretAttemptRepo.(*fakeAppSecretAttemptRepo) //nolint:forcetypeassert
}

// updateReq asks to switch secret retrieval on - the change every test below
// cares about, since it is the one that has to be hard to make.
func updateReq(appSecret string) *hpappsettingsdto.UpdateSecuritySettingsReq {
	return &hpappsettingsdto.UpdateSecuritySettingsReq{
		AppSecret:           appSecret,
		ReturnSecretsViaAPI: true,
	}
}

func adminAuth() *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{
		User: &entity.User{ID: "admin-user", Username: "admin"},
	}}
}

// savedFlag reads the flag back out of the managed file itself rather than out of
// the loader, so the assertion is about what the next process will actually find.
func savedFlag(t *testing.T) bool {
	t.Helper()
	path := config.ManagedSettingsPath(config.Current().AppPath)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	var file struct {
		Security struct {
			ReturnSecretsViaAPI *bool `toml:"return_secrets_via_api"`
		} `toml:"security"`
	}
	_, err := toml.DecodeFile(path, &file)
	assert.NoError(t, err)

	return file.Security.ReturnSecretsViaAPI != nil && *file.Security.ReturnSecretsViaAPI
}

func TestUpdateSecuritySettingsRejectsAWrongAppSecret(t *testing.T) {
	uc, audit, hpApp := newSecurityUCTest(t)

	_, err := uc.UpdateSecuritySettings(context.Background(), adminAuth(),
		&hpappsettingsdto.UpdateSecuritySettingsReq{
			AppSecret:           "not-the-app-secret",
			ReturnSecretsViaAPI: true,
		})

	assert.ErrorIs(t, err, hperrors.ErrAppSecretMismatched)
	assert.False(t, config.Current().Security.ReturnSecretsViaAPI, "a refused change must not be applied")
	assert.Equal(t, 0, hpApp.reloads, "a refused change must not be broadcast")

	// The refusal is the interesting half: a wrong app secret against an
	// admin-only endpoint is somebody working from a session they should not have.
	assert.Len(t, audit.entries, 1)
	assert.Equal(t, base.AuditLogResultDenied, audit.entries[0].Result)
	assert.Equal(t, base.AuditLogTypeSecuritySettingsUpdate, audit.entries[0].Type)
	assert.NotContains(t, audit.entries[0].Detail, "app-secret", "the record must not carry the secret")
}

func TestUpdateSecuritySettingsAppliesTheChange(t *testing.T) {
	uc, audit, hpApp := newSecurityUCTest(t)

	_, err := uc.UpdateSecuritySettings(context.Background(), adminAuth(),
		&hpappsettingsdto.UpdateSecuritySettingsReq{
			AppSecret:           testAppSecret,
			ReturnSecretsViaAPI: true,
		})

	assert.NoError(t, err)
	assert.True(t, config.Current().Security.ReturnSecretsViaAPI)
	assert.True(t, savedFlag(t), "the change must survive this process")
	assert.Equal(t, 1, hpApp.reloads, "the other replicas have to be told")

	assert.Len(t, audit.entries, 1)
	assert.Equal(t, base.AuditLogResultAllowed, audit.entries[0].Result)
	assert.Contains(t, audit.entries[0].Detail, `"returnSecretsViaApi":false`)
	assert.Contains(t, audit.entries[0].Detail, `"returnSecretsViaApi":true`)
}

// An exemption is a hole in the flag, so the record has to show it. A record that
// said only "the flag stayed off" while a secret type was added to the list would
// be worse than none: true, and silent about the part that mattered.
func TestUpdateSecuritySettingsRecordsTheExemptions(t *testing.T) {
	uc, audit, _ := newSecurityUCTest(t)

	_, err := uc.UpdateSecuritySettings(context.Background(), adminAuth(),
		&hpappsettingsdto.UpdateSecuritySettingsReq{
			AppSecret:               testAppSecret,
			AlwaysReturnSecretTypes: []string{string(base.SecretTypeSwarmJoinToken)},
		})

	assert.NoError(t, err)
	assert.False(t, config.Current().Security.ReturnSecretsViaAPI,
		"an exemption must not turn the flag on")
	assert.Equal(t, []string{string(base.SecretTypeSwarmJoinToken)},
		config.Current().Security.AlwaysReturnSecretTypes)

	assert.Len(t, audit.entries, 1)
	assert.Contains(t, audit.entries[0].Detail, `"alwaysReturnSecretTypes":["swarm-join-token"]`)
}

// Turning the flag back off has to reach the file, or the next start would read
// the old "on" and quietly resume handing secrets out.
func TestUpdateSecuritySettingsTurnsTheFlagOff(t *testing.T) {
	uc, _, hpApp := newSecurityUCTest(t)
	assert.NoError(t, config.SaveSecuritySettings(&config.Security{ReturnSecretsViaAPI: true}))

	_, err := uc.UpdateSecuritySettings(context.Background(), adminAuth(),
		&hpappsettingsdto.UpdateSecuritySettingsReq{
			AppSecret:           testAppSecret,
			ReturnSecretsViaAPI: false,
		})

	assert.NoError(t, err)
	assert.False(t, config.Current().Security.ReturnSecretsViaAPI)
	assert.False(t, savedFlag(t))
	assert.Equal(t, 1, hpApp.reloads)
}

func TestUpdateSecuritySettingsSkipsAnUnchangedRequest(t *testing.T) {
	uc, audit, hpApp := newSecurityUCTest(t)

	_, err := uc.UpdateSecuritySettings(context.Background(), adminAuth(),
		&hpappsettingsdto.UpdateSecuritySettingsReq{
			AppSecret:           testAppSecret,
			ReturnSecretsViaAPI: false, // already the current value
		})

	assert.NoError(t, err)
	assert.Equal(t, 0, hpApp.reloads, "nothing changed, so nothing to restart for")

	// Still recorded. Otherwise resubmitting the current settings would be a free
	// way to test app secrets, leaving no trace of the guesses that were right.
	assert.Len(t, audit.entries, 1)
	assert.Equal(t, base.AuditLogResultAllowed, audit.entries[0].Result)
}

// An action that could not be recorded does not happen.
func TestUpdateSecuritySettingsFailsClosedWhenNotRecordable(t *testing.T) {
	uc, audit, hpApp := newSecurityUCTest(t)
	audit.err = errors.New("audit store is down")

	_, err := uc.UpdateSecuritySettings(context.Background(), adminAuth(),
		&hpappsettingsdto.UpdateSecuritySettingsReq{
			AppSecret:           testAppSecret,
			ReturnSecretsViaAPI: true,
		})

	assert.Error(t, err)
	assert.False(t, config.Current().Security.ReturnSecretsViaAPI)
	assert.False(t, savedFlag(t))
	assert.Equal(t, 0, hpApp.reloads)
}

// A reload that did not happen leaves the running replicas on the old flags, so
// the operator has to hear about it rather than see a success.
func TestUpdateSecuritySettingsReportsAFailedReload(t *testing.T) {
	uc, _, hpApp := newSecurityUCTest(t)
	hpApp.err = errors.New("docker is unreachable")

	_, err := uc.UpdateSecuritySettings(context.Background(), adminAuth(),
		&hpappsettingsdto.UpdateSecuritySettingsReq{
			AppSecret:           testAppSecret,
			ReturnSecretsViaAPI: true,
		})

	assert.Error(t, err)
	assert.True(t, savedFlag(t), "the settings were still saved, and the error must say so")
}

func TestUpdateSecuritySettingsRefusesTheDemoUser(t *testing.T) {
	uc, audit, hpApp := newSecurityUCTest(t)
	auth := adminAuth()
	auth.User.User.ID = "demo-user"

	_, err := uc.UpdateSecuritySettings(context.Background(), auth,
		&hpappsettingsdto.UpdateSecuritySettingsReq{
			AppSecret:           testAppSecret,
			ReturnSecretsViaAPI: true,
		})

	assert.ErrorIs(t, err, hperrors.ErrUserDemoUnauthorized)
	assert.Empty(t, audit.entries)
	assert.Equal(t, 0, hpApp.reloads)
}

func TestUpdateSecuritySettingsCountsWrongAppSecrets(t *testing.T) {
	uc, _, _ := newSecurityUCTest(t)

	for i := 1; i <= 3; i++ {
		_, err := uc.UpdateSecuritySettings(context.Background(), adminAuth(),
			updateReq("wrong"))
		assert.ErrorIs(t, err, hperrors.ErrAppSecretMismatched)
		assert.Equal(t, i, attemptsOf(uc).attempt.Fails)
	}
}

// The point of counting: past the threshold the secret is not even looked at, so
// a caller who has been guessing cannot get through by finally guessing right.
func TestUpdateSecuritySettingsLocksOutAfterTooManyFailures(t *testing.T) {
	uc, audit, hpApp := newSecurityUCTest(t)
	attemptsOf(uc).attempt = &cacheentity.AppSecretAttempt{
		Fails:      appSecretMaxFailsInARow,
		LastFailAt: timeutil.NowUTC(),
	}

	_, err := uc.UpdateSecuritySettings(context.Background(), adminAuth(),
		updateReq(testAppSecret)) // the correct one
	assert.ErrorIs(t, err, hperrors.ErrTooManyAppSecretFailures)

	assert.False(t, config.Current().Security.ReturnSecretsViaAPI)
	assert.Equal(t, 0, hpApp.reloads)
	assert.Len(t, audit.entries, 1)
	assert.Equal(t, base.AuditLogResultDenied, audit.entries[0].Result)
}

// The wait is measured from the last failure, so waiting it out works.
func TestUpdateSecuritySettingsAllowsAfterTheWaitElapses(t *testing.T) {
	uc, _, hpApp := newSecurityUCTest(t)
	attemptsOf(uc).attempt = &cacheentity.AppSecretAttempt{
		Fails:      appSecretMaxFailsInARow,
		LastFailAt: timeutil.NowUTC().Add(-time.Hour),
	}

	_, err := uc.UpdateSecuritySettings(context.Background(), adminAuth(),
		updateReq(testAppSecret))
	assert.NoError(t, err)
	assert.Equal(t, 1, hpApp.reloads)
}

func TestUpdateSecuritySettingsClearsTheCountOnSuccess(t *testing.T) {
	uc, _, _ := newSecurityUCTest(t)
	attemptsOf(uc).attempt = &cacheentity.AppSecretAttempt{Fails: 2}

	_, err := uc.UpdateSecuritySettings(context.Background(), adminAuth(),
		updateReq(testAppSecret))
	assert.NoError(t, err)
	assert.Equal(t, 1, attemptsOf(uc).dels, "a correct answer ends the run")
	assert.Nil(t, attemptsOf(uc).attempt)
}

// No stored count means no backoff, and this endpoint is not one to run unmetered
// because a cache is unreachable.
func TestUpdateSecuritySettingsFailsClosedWhenTheCountCannotBeRead(t *testing.T) {
	uc, _, hpApp := newSecurityUCTest(t)
	attemptsOf(uc).getErr = errors.New("redis is unreachable")

	_, err := uc.UpdateSecuritySettings(context.Background(), adminAuth(),
		updateReq(testAppSecret))
	assert.Error(t, err)
	assert.False(t, savedFlag(t))
	assert.Equal(t, 0, hpApp.reloads)
}
