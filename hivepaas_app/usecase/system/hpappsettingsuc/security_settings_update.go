package hpappsettingsuc

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/failbackoff"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

const (
	// appSecretMaxFailsInARow is tighter than the login threshold next door. This
	// is a value the operator already knows being re-entered on a page they went
	// looking for, not a password typed from memory at the start of a day, so a
	// run of wrong answers means something different here.
	appSecretMaxFailsInARow = 5
	appSecretBackoffStep    = 2 * time.Minute

	// appSecretAttemptExp outlives the longest wait the policy can impose, or the
	// count would expire before the wait it caused ever elapsed.
	appSecretAttemptExp = 4 * time.Hour
)

// appSecretBackoff makes each wrong app secret cost more than the last.
//
// It counts against the user, not the address. A network rate limit - which the
// API paths also carry - stops a flood from one place; it does nothing about a
// patient caller on a hijacked admin session who can change address at will. Only
// the account is constant across those attempts.
var appSecretBackoff = failbackoff.Policy{
	MaxFailsInARow: appSecretMaxFailsInARow,
	Step:           appSecretBackoffStep,
}

// UpdateSecuritySettings changes the operator-level security switches.
//
// These do not live in the database. They gate whether stored secrets may leave
// the server at all, and a switch the app can rewrite in its own tables is a
// switch that falls with the database - so they go to the managed file, which is
// applied over everything else at load. See config.SaveSecuritySettings.
//
// The caller re-enters the app secret. The endpoint is already admin-only, so
// this is not authentication; it is the difference between an admin and the
// operator. Reaching an admin session - a stolen cookie, a borrowed laptop, an
// admin API key - is not supposed to be enough to switch on the retrieval of
// every secret in the system, and before this endpoint existed it was not:
// the flag could only be set in a file on the host. Requiring the secret is what
// keeps that property now that there is a way in over the network.
func (uc *UC) UpdateSecuritySettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.UpdateSecuritySettingsReq,
) (*hpappsettingsdto.UpdateSecuritySettingsResp, error) {
	if auth.User.IsDemoUser() {
		return nil, hperrors.Wrap(hperrors.ErrUserDemoUnauthorized)
	}

	current := config.Current().Security
	newSettings := req.ToConfig()

	if err := uc.authorizeSecuritySettingsUpdate(ctx, auth, req.AppSecret, &current, newSettings); err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Nothing to write and nothing to tell the other replicas about. The attempt
	// is still recorded above: a caller who reaches this line has just had an app
	// secret confirmed, and that is worth knowing whether or not they changed
	// anything with it.
	if *newSettings == current {
		return &hpappsettingsdto.UpdateSecuritySettingsResp{}, nil
	}

	if err := config.SaveSecuritySettings(newSettings); err != nil {
		return nil, hperrors.Wrap(err)
	}

	if err := uc.applySecuritySettings(ctx); err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &hpappsettingsdto.UpdateSecuritySettingsResp{}, nil
}

// authorizeSecuritySettingsUpdate checks the app secret and records the attempt,
// allowed or refused.
//
// The record is written before anything is applied and its failure aborts the
// change, for the same reason the reveal path does it that way: an unrecorded
// flip of these flags is exactly the event the record exists to capture. The
// refusals matter at least as much - a wrong app secret against an admin-only
// endpoint is somebody working from a session they should not have.
func (uc *UC) authorizeSecuritySettingsUpdate(
	ctx context.Context,
	auth *basedto.Auth,
	appSecret string,
	current, newSettings *config.Security,
) error {
	allowed, denyErr := uc.verifyOperator(ctx, auth, appSecret)

	result := base.AuditLogResultAllowed
	if !allowed {
		result = base.AuditLogResultDenied
	}
	err := uc.auditService.Record(ctx, uc.db, &auditservice.Entry{
		Type:    base.AuditLogTypeSecuritySettingsUpdate,
		Scope:   base.ObjectScopeHivepaas,
		Source:  base.AuditLogSourceAPIUpdate,
		Result:  result,
		Auth:    auth,
		ResType: base.ResourceTypeSecuritySettings,
		ResName: "security settings",
		Detail:  securitySettingsDetail(current, newSettings),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}

	return denyErr
}

// verifyOperator checks the app secret, and makes a wrong one cost time.
//
// A refusal is counted before it is returned, and the count is what a later
// attempt is measured against, so guessing gets slower on its own. A correct
// answer clears the count: the wait exists to price failures, not to punish
// somebody who mistyped once.
func (uc *UC) verifyOperator(
	ctx context.Context,
	auth *basedto.Auth,
	appSecret string,
) (bool, error) {
	userID := actorID(auth)
	if userID == "" {
		// Nothing to count the failures against, so the check cannot be metered.
		// Refuse rather than run it for free.
		return false, hperrors.Wrap(hperrors.ErrUnauthorized).
			WithMsgLog("no actor to meter app secret attempts against")
	}

	attempt, err := uc.loadAppSecretAttempt(ctx, userID)
	if err != nil {
		// Fail closed. Without the stored count there is no backoff, and an
		// endpoint that hands out the ability to read every secret is not one to
		// leave unmetered because a cache is unreachable.
		return false, hperrors.Wrap(err)
	}

	if wait := appSecretBackoff.Wait(attempt, timeutil.NowUTC()); wait > 0 {
		return false, hperrors.Wrap(hperrors.ErrTooManyAppSecretFailures).
			WithParam("WaitDuration", int(wait.Seconds()))
	}

	if !verifyAppSecret(appSecret) {
		uc.saveAppSecretFailure(ctx, userID, attempt)
		return false, hperrors.Wrap(hperrors.ErrAppSecretMismatched)
	}

	uc.clearAppSecretFailures(ctx, userID, attempt)
	return true, nil
}

// actorID is who the failures are counted against.
func actorID(auth *basedto.Auth) string {
	if auth == nil {
		return ""
	}
	if user := auth.User.Entity(); user != nil {
		return user.ID
	}
	return ""
}

// loadAppSecretAttempt returns the stored count, or nil when there is none. A
// missing entry is the normal case and not an error.
func (uc *UC) loadAppSecretAttempt(
	ctx context.Context,
	userID string,
) (*cacheentity.AppSecretAttempt, error) {
	attempt, err := uc.cacheAppSecretAttemptRepo.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, nil //nolint:nilnil // no attempt stored is not a failure
		}
		return nil, hperrors.Wrap(err)
	}
	return attempt, nil
}

// saveAppSecretFailure records one more wrong answer.
//
// Best effort, and logged when it fails. It cannot be fail-closed the way the
// read is: the caller is already being refused, and a write that did not land
// shows up as a read failure on the next attempt, which does fail closed.
func (uc *UC) saveAppSecretFailure(
	ctx context.Context,
	userID string,
	attempt *cacheentity.AppSecretAttempt,
) {
	updated := appSecretBackoff.Fail(attempt, timeutil.NowUTC())
	if err := uc.cacheAppSecretAttemptRepo.Set(ctx, userID, updated, appSecretAttemptExp); err != nil {
		logging.Warnf("failed to record an app secret failure for user %s: %v", userID, err)
	}
}

// clearAppSecretFailures forgets the run of wrong answers a correct one ended.
func (uc *UC) clearAppSecretFailures(
	ctx context.Context,
	userID string,
	attempt *cacheentity.AppSecretAttempt,
) {
	if attempt == nil {
		return
	}
	if err := uc.cacheAppSecretAttemptRepo.Del(ctx, userID); err != nil {
		logging.Warnf("failed to clear app secret failures for user %s: %v", userID, err)
	}
}

// verifyAppSecret reports whether the given value is the app secret in use.
//
// The comparison is constant time. The endpoint is admin-only and rate limiting
// would be the real defense, but a byte-at-a-time comparison against a value that
// unlocks every stored secret is not worth leaving on the table.
func verifyAppSecret(appSecret string) bool {
	return subtle.ConstantTimeCompare([]byte(appSecret), []byte(config.Current().Secret)) == 1
}

// securitySettingsDetail describes the change for the audit record. It carries
// the flags only - never the app secret that authorized it.
func securitySettingsDetail(current, newSettings *config.Security) string {
	detail, err := json.Marshal(map[string]any{
		"from": map[string]any{"returnSecretsViaApi": current.ReturnSecretsViaAPI},
		"to":   map[string]any{"returnSecretsViaApi": newSettings.ReturnSecretsViaAPI},
	})
	if err != nil {
		return ""
	}
	return string(detail)
}

// applySecuritySettings makes the saved settings take effect everywhere.
//
// A reload, not a restart. The flags are read live out of config.Current on every
// request, so re-reading the config is enough to apply them - and a restart would
// take down the very replica serving this request, so the operator would see a
// dropped connection instead of an answer and have no way to tell whether the
// change landed.
//
// Unlike the app secret rotation next door this is not best effort. A reload that
// did not happen leaves the running replicas still enforcing the old flags, and
// in the direction that matters - switching secret retrieval off - that means it
// is still on. Saying so is the only way the operator knows to restart by hand.
func (uc *UC) applySecuritySettings(ctx context.Context) error {
	if err := uc.hpAppService.ReloadHpAppConfig(ctx); err != nil {
		return hperrors.Wrap(err).
			WithMsgLog("the security settings were saved but could not be applied to the running app")
	}
	return nil
}
