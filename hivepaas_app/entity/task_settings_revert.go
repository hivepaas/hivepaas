package entity

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// TaskSettingsRevertArgs is the record of a settings change on probation.
//
// It carries the whole known-good snapshot rather than a pointer to it, because
// the point of the task is to survive the change it guards: anything it would
// have to look up could itself be part of what went wrong.
type TaskSettingsRevertArgs struct {
	AppID       string           `json:"appId"`
	SettingID   string           `json:"settingId"`
	SettingType base.SettingType `json:"settingType"`

	// ProbationVer is Setting.UpdateVer as it stands right after the change was
	// applied. The executor reverts only while the setting still carries it - see
	// the staleness check there.
	ProbationVer int `json:"probationVer"`

	// Snapshot is the setting as it was before the change: the state to restore.
	Snapshot SettingSnapshot `json:"snapshot"`

	AppliedBy string    `json:"appliedBy,omitempty"`
	AppliedAt time.Time `json:"appliedAt"`

	// ConfirmableAt is when a confirmation starts being worth something. See
	// SettingsProbationSettleDelay for what decides it.
	ConfirmableAt time.Time `json:"confirmableAt"`
	DeadlineAt    time.Time `json:"deadlineAt"`
}

// SettingSnapshot is a setting's payload frozen at a point in time.
//
// Version travels with Data because the two are read together: Migrate() decides
// what to do with the payload by looking at the version it was written under.
type SettingSnapshot struct {
	Data    string `json:"data"`
	Version int    `json:"version"`
}

// SettingsProbationSettleDelay is how long a confirmation is worth nothing.
//
// The danger it guards is narrow and specific: a confirmation that the *previous*
// configuration served. That one succeeds, cancels the trial, and leaves the new
// configuration live with nothing left to undo it. A confirmation that fails
// costs nothing - the operator retries - so this only has to outlast the moment
// the old configuration stops answering, not the moment the new one is fully up.
//
// Two things decide that moment, and the slower one wins:
//
//   - A label change is picked up by traefik's swarm provider, which polls every
//     15s. Until it does, the old routers are the live ones.
//   - A change to traefik's container spec replaces its task. Measured on a
//     single-node swarm, three runs within 50ms of each other: the old task stops
//     answering 3.2s after the API call, and the new one is serving at 4.2s - an
//     outage of about one second.
//
// So the poll is the binding constraint, and 30s is twice it. The task
// replacement being the *faster* of the two is worth stating plainly, because the
// opposite was assumed here for some time: this was 120s, sized against traefik's
// start_period of 60s and the app's of 120s. That was measuring the wrong thing.
// start_period is a grace window for the healthcheck, not a delay anything waits
// out - swarm marked the replacement task healthy at 9.6s in the same runs, and
// it was serving five seconds before that.
const SettingsProbationSettleDelay = 30 * time.Second

// SettingsProbationAppRestartSettleDelay is the same idea for a change that takes
// the HivePaaS app itself down and brings it back.
//
// Not a correctness bound - the one above already covers that, and an app that is
// not running cannot serve a wrong confirmation either. This is about what the
// dashboard tells the operator while it waits. Before ConfirmableFrom, a probe
// that cannot reach HivePaaS means "still restarting"; after it, the dashboard
// says the change may have locked them out and advises them to sit out the
// countdown - advice that would throw away a perfectly good change if the app was
// merely still booting.
//
// So it is sized against the app coming back rather than traefik: its start_period
// is 120s because SystemInstallation is the slowest thing in that path.
//
// Only the caller knows whether a given request restarts the app - the same
// endpoint can carry proxy fields alone, or a replica count with them - which is
// why this is passed in at arm time rather than derived from the setting type.
const SettingsProbationAppRestartSettleDelay = 120 * time.Second

// ConfirmableFrom is the earliest moment a confirmation of this change means
// anything.
//
// Read from the record rather than recomputed, because what the change disturbs
// is known when it is applied and not afterwards. A task written before this
// field existed carries the zero time and falls back to the floor.
func (a *TaskSettingsRevertArgs) ConfirmableFrom() time.Time {
	if a.ConfirmableAt.IsZero() {
		return a.AppliedAt.Add(SettingsProbationSettleDelay)
	}
	return a.ConfirmableAt
}

type TaskSettingsRevertOutput struct {
	Reverted bool   `json:"reverted"`
	Reason   string `json:"reason,omitempty"`
}

func (t *Task) ArgsAsSettingsRevert() (*TaskSettingsRevertArgs, error) {
	return parseTaskArgsAs(t, func() *TaskSettingsRevertArgs { return &TaskSettingsRevertArgs{} })
}

func (t *Task) OutputAsSettingsRevert() (*TaskSettingsRevertOutput, error) {
	return parseTaskOutputAs(t, func() *TaskSettingsRevertOutput { return &TaskSettingsRevertOutput{} })
}

// SettingSnapshotOf freezes the setting's current payload.
func SettingSnapshotOf(setting *Setting) SettingSnapshot {
	if setting == nil {
		return SettingSnapshot{}
	}
	return SettingSnapshot{Data: setting.Data, Version: setting.Version}
}

// RestoreTo writes the snapshot back over the setting, leaving identity and
// bookkeeping columns to the caller.
//
// parsedData is dropped along with it: it is a cache of the payload being
// replaced, and leaving it in place would let a later MustAsX() hand back the
// settings that are being reverted away from.
func (s SettingSnapshot) RestoreTo(setting *Setting) {
	setting.Data = s.Data
	setting.Version = s.Version
	setting.parsedData = nil
}
