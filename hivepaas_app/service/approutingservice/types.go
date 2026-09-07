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
