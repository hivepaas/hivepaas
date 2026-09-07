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

	AppliedBy  string    `json:"appliedBy,omitempty"`
	AppliedAt  time.Time `json:"appliedAt"`
	DeadlineAt time.Time `json:"deadlineAt"`
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
// Traefik's swarm provider polls every 15s, so a request arriving sooner may well
// have traveled through the configuration being replaced: it would vouch for a
// router that is not live yet, and the real one would go live afterwards with
// nothing left to undo it.
//
// It lives here rather than with the usecase because the answer is reported to
// callers as ConfirmableFrom and enforced on the way back in, and those two have
// to be the same number.
const SettingsProbationSettleDelay = 25 * time.Second

// ConfirmableFrom is the earliest moment a confirmation of this change means
// anything.
func (a *TaskSettingsRevertArgs) ConfirmableFrom() time.Time {
	return a.AppliedAt.Add(SettingsProbationSettleDelay)
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

// SnapshotOf freezes the setting's current payload.
func SnapshotOf(setting *Setting) SettingSnapshot {
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
