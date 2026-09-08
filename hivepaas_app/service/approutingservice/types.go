package approutingservice

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type ApplyAppRoutingReq struct {
	App                       *entity.App
	RoutingSettings           *entity.AppRoutingSettings
	RefObjects                *entity.RefObjects
	Service                   *swarm.Service
	SkipApplyingSslCerts      bool
	ForceRecreateSslCertFiles bool
	SkipApplyingNetworks      bool
	SkipUpdatingService       bool

	// SkipMissingRefObjects keeps a reference that has gone since from being fatal.
	//
	// Set it on the way back, never on the way forward. Undoing a change is the
	// path somebody takes when they cannot get in, so a middleware that can no
	// longer be resolved has to be dropped rather than allowed to abort the only
	// way back. Applying a change is the opposite: dropping the reference there
	// would quietly take the basic auth off an app that is meant to have it, and
	// leaving that app on a stale ip-strategy depth is the smaller harm.
	SkipMissingRefObjects bool
}

type ApplyAppRoutingResp struct {
	Service *swarm.Service
}

// RevertSettingsReq asks for a setting to be put back the way the snapshot has it.
//
// ProbationVer is what makes a stale revert harmless: the change is undone only
// while the setting still carries the version it had right after being applied.
// Anything that has moved on since - a confirmed change followed by a new one, or
// a revert that already ran - leaves the setting on a different version, and the
// request turns into a no-op instead of overwriting work it knows nothing about.
type RevertSettingsReq struct {
	App          *entity.App
	SettingID    string
	ProbationVer int
	Snapshot     entity.SettingSnapshot
}

type RevertSettingsResp struct {
	Reverted bool
	// Reason says why nothing was done, when nothing was done.
	Reason  string
	Service *swarm.Service
}

type ReapplyClientIPStrategyReq struct {
	// SkipMissingRefObjects is passed through to each app's apply - see the field
	// of the same name on ApplyAppRoutingReq for why it is set only on a revert.
	SkipMissingRefObjects bool

	// PrimaryOnly limits the sweep to PrimaryAppID.
	//
	// Set while a change is on trial. The trial asks one question - can the
	// operator still get in - and only the HivePaaS app answers it, so only that
	// app has to carry the new depth while the clock runs. The rest are swept once
	// the change is confirmed, which keeps the undo down to a single app and keeps
	// a fan-out off the path between applying a change and being able to confirm
	// it.
	//
	// Deferring costs those apps nothing they were not already paying: an operator
	// editing the proxy settings is doing so because the current topology is
	// described wrongly, and an app reading the wrong forwarded position is
	// already refusing the callers its allowlist was meant to admit. Reverting
	// them would only put back a configuration that is wrong in the other
	// direction.
	PrimaryOnly bool

	// PrimaryAppID is applied first and its failure aborts the rest.
	//
	// It is the HivePaaS app: the trial that this sweep runs inside exists to
	// guard reachability of exactly that one, so leaving it until last would mean
	// the operator could confirm a change that had not been applied to the thing
	// they are confirming through.
	PrimaryAppID string
}

type ReapplyClientIPStrategyResp struct {
	// Applied counts the apps whose labels were regenerated.
	Applied int
	// Skipped counts apps that carry no ip-strategy depth to begin with.
	Skipped int
	// Failed names the apps that could not be updated, with the reason. The sweep
	// keeps going past these: one app's docker hiccup must not stop the others
	// from getting the depth they need.
	Failed map[string]string
}
