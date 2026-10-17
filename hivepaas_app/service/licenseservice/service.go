// Package licenseservice is the license this installation runs on, and the
// plan limits it puts in force. See docs/superpowers/specs/2026-10-16-licensing-design.md.
//
// A limit only ever stops creating: Check refuses what would take a kind past
// its limit, and nothing here stops, scales down or deletes what exists.
package licenseservice

import (
	"context"
	"crypto/sha256"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
)

type Service interface {
	// Current is the license in force: the stored key, verified and evaluated for
	// this installation now. With no key stored it is Bee, status none.
	//
	// The State carries the key's payload, with the customer's email: never log
	// it. Its Key is shared with the service's cache, so it is read, never
	// changed.
	Current(ctx context.Context) (*license.State, error)

	// Usage counts each limited kind as Check counts it, without its locks, and
	// the paid sign-in providers that are active, which Check does not count.
	Usage(ctx context.Context, db database.IDB) (*Usage, error)

	// Check refuses, with ERR_PLAN_LIMIT_REACHED, a creation that would take a
	// kind past the limit of the plan in force. A kind the plan does not limit is
	// passed over without a lock or a count, so paid plans never wait on one.
	//
	// The kinds counted in the database - users, apps, functions and previews -
	// are counted after taking their plan-limit lock, which is held until the
	// transaction ends: db must be the transaction that creates them, before it
	// creates them, on every plan (ErrInternal otherwise). Nodes are counted from
	// the swarm, and sign-in providers are not counted at all, so neither needs
	// one.
	Check(ctx context.Context, db database.IDB, req *CheckReq) error

	// CheckFeatures refuses, with ERR_PLAN_FEATURE_REQUIRED, a creation that needs
	// a feature the license in force does not grant, naming the first one missing.
	// Features are named as templates name them; one this HivePaaS does not know
	// is never granted. Needing none reads nothing.
	//
	// Only creating is refused for want of a feature. What exists keeps running,
	// and restoring is never checked.
	CheckFeatures(ctx context.Context, features []string) error

	// Install verifies a key and stores it as the license, replacing the one
	// stored, whatever its plan. It refuses a key that is invalid, for another
	// installation, or past its grace, and keeps the stored key then.
	Install(ctx context.Context, db database.Tx, req *InstallReq) (*InstallResp, error)

	// InstallationID is the ID a key names this installation by: its swarm's.
	InstallationID(ctx context.Context) (string, error)

	// FetchRefresh asks the portal whether there is a newer key than the one
	// stored. It sends the stored key and this installation's ID, and nothing
	// else, and writes nothing. Asking that fails is an answer, failed, with what
	// went wrong as its Err; only a stored license that cannot be read is an
	// error. With no key stored it asks nothing; nor does the daily check while it
	// is switched off in the config.
	FetchRefresh(ctx context.Context, req *FetchRefreshReq) (*FetchRefreshResp, error)

	// ApplyRefresh stores what the portal answered, with the license setting
	// locked: a newer key, installed as Install installs one, or when the key was
	// moved; and how the check went, in the setting's Refresh. A newer key that
	// Install would refuse is not installed, and the check failed. An answer for a
	// key no longer stored - one installed while the portal was asked - is
	// dropped.
	ApplyRefresh(ctx context.Context, db database.Tx, fetched *FetchRefreshResp) (*ApplyRefreshResp, error)

	// RefreshInfo is whether the daily check is on, and how the last check went.
	RefreshInfo(ctx context.Context) (*RefreshInfo, error)
}

// CheckReq is how many of each kind a creation adds.
type CheckReq struct {
	Adding map[base.LicenseLimit]int
}

// Adding is a CheckReq for n of one kind.
func Adding(limit base.LicenseLimit, n int) *CheckReq {
	return &CheckReq{Adding: map[base.LicenseLimit]int{limit: n}}
}

// Usage is how many of each limited kind there are. For sign-in providers it is
// the paid ones active, which no plan counts against, but which the dashboard
// marks on Bee.
type Usage struct {
	Used map[base.LicenseLimit]int
}

type InstallReq struct {
	// Key is the key's text, as pasted or as the portal sent it.
	Key string
}

type InstallResp struct {
	// Setting is the license setting as stored.
	Setting *entity.Setting
	// Created is the setting written for the first time.
	Created bool
	// State is the license the key puts in force.
	State *license.State
}

// CloneLimit is what a clone counts against. A function's clone stays a function
// only with its deployment settings: without them it is left with no kind, which
// is an app.
func CloneLimit(srcIsFunction, cloneDeploymentSettings bool) base.LicenseLimit {
	if srcIsFunction && cloneDeploymentSettings {
		return base.LicenseLimitFunctions
	}
	return base.LicenseLimitApps
}

type FetchRefreshReq struct {
	// Manual is a person asking now. The daily check is not manual, and asks
	// nothing while it is switched off in the config.
	Manual bool
}

// FetchRefreshResp is what the portal answered about the stored key.
type FetchRefreshResp struct {
	Result base.LicenseRefreshResult
	// SentKeySum is the sha256 of the key asked about, so the answer is applied
	// only to that key.
	SentKeySum [sha256.Size]byte
	// NewKey is the key to install, for newer. It is not verified yet.
	NewKey string
	// MovedAt is when the key was moved to another installation, for moved; zero
	// when the portal did not say.
	MovedAt time.Time
	// CheckedAt is when the portal was asked.
	CheckedAt time.Time
	// Err is why the check failed, for failed.
	Err error
}

// ApplyRefreshResp is what applying a refresh changed.
type ApplyRefreshResp struct {
	// Result is the answer as applied: failed for a newer key that could not be
	// installed, and current for an answer about a key no longer stored.
	Result base.LicenseRefreshResult
	// Changed is the license changed: a key installed, or a move learned of.
	Changed bool
	// Setting is the license setting as stored; nil when nothing was recorded.
	Setting *entity.Setting
	// Installed is the key installed, for newer.
	Installed *InstallResp
	// State is the license in force once it changed; nil when it did not.
	State *license.State
	// Err is why the refresh failed, for failed: the portal's, or why the newer
	// key could not be installed.
	Err error
}

// RefreshInfo is the daily check with the portal: whether it is on, and how the
// last check went.
type RefreshInfo struct {
	Enabled bool
	Refresh entity.LicenseRefresh
}
