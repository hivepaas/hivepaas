package sysupdateservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type SysUpdateReq struct {
	*queue.TaskExecData

	// PauseTaskQueues, when set, is called once the images are pulled, right
	// before the app and the worker are scaled down.
	PauseTaskQueues func() error
}

type SysUpdateResp struct {
}

// UpdatePlan is what an update to a release would do.
type UpdatePlan struct {
	// Components are in the order the update reaches them.
	Components []*ComponentChange
	// Blocked is whether a component would refuse to move, which fails the
	// update as a whole: it would stop at that component.
	Blocked bool
	// RequiresBackup is whether a component is moved from the database backup,
	// so the update cannot be run without one.
	RequiresBackup bool
}

// Change is what an update does to one component.
type Change string

const (
	// ChangeNone leaves the component as it is: already on the image, or the
	// image is not newer, or the release names none.
	ChangeNone Change = "none"
	// ChangeSettings leaves the component's image as it is, and brings the rest
	// of its spec to what the release writes, which restarts it.
	//
	// It is judged by the running version, against what the running version
	// writes, while the update runs on the target's image and writes what the
	// target does. A target that writes something the running version does not
	// know of - other access log fields for Traefik - restarts the component with
	// the plan saying ChangeNone, so the screen says the proxy may still restart.
	ChangeSettings Change = "settings"
	// ChangeUpdate moves the component to a newer image of the same major.
	ChangeUpdate Change = "update"
	// ChangeMajor moves it across a major version.
	ChangeMajor Change = "major"
	// ChangeBlocked is a major move the release says the updater must not make.
	ChangeBlocked Change = "blocked"
	// ChangeNotDeployed is a component this installation does not run.
	ChangeNotDeployed Change = "not-deployed"
)

// ComponentChange is what an update does to one component.
type ComponentChange struct {
	// Key is the release's name for it: base.HivepaasDbKey and the rest.
	Key          string
	CurrentImage string
	TargetImage  string
	Change       Change
	// Reason is the update's own words for its decision.
	Reason string
	// RequiresBackup is a move made from the database backup: a postgres major
	// loads the new cluster from it.
	RequiresBackup bool
	// InterruptsTraffic is a change that restarts the proxy every app is
	// reached through, so apps are unreachable for a moment.
	InterruptsTraffic bool
}
