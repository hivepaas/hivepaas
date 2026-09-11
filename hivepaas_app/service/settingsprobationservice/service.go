// Package settingsprobationservice is confirm-or-revert: the machinery that
// applies a settings change immediately and then undoes it unless somebody comes
// back through the new configuration and vouches for it.
//
// It guards everything a request cannot be read for - a middleware reference that
// no longer resolves, a domain whose DNS is not ready, a rate limit set too low,
// a traefik argument that parses but discovers no routers. None of those can be
// recognized before applying them; all of them are obvious the moment somebody
// tries to come back in.
//
// The load-bearing assumption is that whoever runs the undo survives the change
// it guards. That holds for every kind of change put on trial: routing settings
// rewrite swarm service labels, which does not recreate a task at all, and a
// traefik change recreates traefik's task while leaving the HivePaaS process -
// the one holding the timer and the docker socket - untouched.
//
// This is the arming half. settingsrevertservice is the undoing half, and the two
// are deliberately separate: the undo has three triggers that must not drift
// apart (the task queue, the timer this package arms, and the startup
// reconciler), while the arming has only ever one.
package settingsprobationservice

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	// Arm records the change as being on trial and builds its undo.
	//
	// The task is built but not scheduled: it is handed to upsertTask so it goes
	// into the same transaction as the change, and the deadline is committed with
	// the thing it guards. Handing it to the queue is Schedule's job, after the
	// commit.
	Arm(ctx context.Context, db database.Tx, auth *basedto.Auth,
		in *ArmReq, out *ArmResult, upsertTask func(task *entity.Task)) error

	// Schedule hands the committed deadline to everything that can enforce it.
	//
	// Called after the commit, and never allowed to fail the change: by that point
	// the row exists whether or not anything managed to schedule it, so a failure
	// costs lateness rather than the guarantee.
	Schedule(ctx context.Context, out *ArmResult)

	// FindPending returns the change of the given kind currently on trial, or nil.
	FindPending(ctx context.Context, db database.IDB, appID string,
		settingType base.SettingType) (*entity.Task, error)

	// Confirm vouches for a change, which is what stops it from being undone.
	Confirm(ctx context.Context, auth *basedto.Auth, in *AnswerReq) error

	// RevertNow undoes the change instead of waiting for its deadline.
	RevertNow(ctx context.Context, auth *basedto.Auth,
		in *AnswerReq) (*entity.TaskSettingsRevertOutput, error)

	// Reconcile picks up trials that nothing was left to enforce.
	//
	// The queue's own scan only looks in a window of TaskCheckInterval around now,
	// so a deadline that passed while the process was down for longer than that is
	// never rediscovered: the task stays not-started for good, and the change it
	// was guarding becomes permanent. This closes that, and re-arms the local timer
	// for trials still in their window.
	Reconcile(ctx context.Context) error
}

// ArmReq is what a caller has to know to put its change on trial.
//
// Snapshot has to be taken before the change is written, and the setting's
// UpdateVer read after - the version it carries once it is the new one.
type ArmReq struct {
	AppID    string
	Setting  *entity.Setting
	Snapshot entity.SettingSnapshot
	Window   time.Duration

	// SettleDelay is how long a confirmation of this change is worth nothing,
	// counted from now. Floored at entity.SettingsProbationSettleDelay, which is
	// the bound every change shares; pass more only for a change that takes the
	// HivePaaS app itself down. See that constant for what decides the number.
	SettleDelay time.Duration
}

// ArmResult is what the caller has to carry out of the transaction so Schedule
// can hand the deadline to the things that enforce it.
type ArmResult struct {
	Probation           *entity.Task
	SupersededProbation *entity.Task
}

// AuditTarget is where the record of an answer is filed, and what it is called.
//
// It has to match where the caller recorded the change itself. The settings
// pages file their writes under the hivepaas scope - the install's own log -
// while the setting row behind a trial belongs to an app, so filing the answer
// under the row's scope puts the two halves of one story in different places:
// the log shows a change being made and never shows whether anybody stood by it.
//
// Section is the page the change was made on, the same value the caller passed
// when it recorded the change, so a reader can follow one page through both.
type AuditTarget struct {
	Scope    base.ObjectScopeType
	ObjectID string
	Section  string
}

// AnswerReq names the trial a caller is answering.
type AnswerReq struct {
	// AppID is the object the trial was filed under - the same one Arm was given.
	AppID       string
	SettingType base.SettingType

	// Audit is where the confirmation or revert is recorded. See AuditTarget.
	Audit AuditTarget

	// ChangeID names the change being answered. It is optional, but sending it is
	// what stops an answer that was in flight during one change from landing on
	// the next one.
	ChangeID string

	// OnConfirmed is work that has to commit with the confirmation, if the caller
	// has any. Tasks it returns are scheduled after the commit.
	//
	// It exists because what confirming entails is not this package's business:
	// confirming the HivePaaS proxy settings releases a change to every other
	// app's labels, which was deliberately deferred until somebody vouched for it,
	// and no other kind of trial has anything of the sort. Running inside the
	// transaction is the point - a confirmation that commits always has its
	// follow-up committed with it.
	//
	// Ignored by RevertNow: there is nothing to release when the change is going
	// away.
	OnConfirmed func(ctx context.Context, db database.Tx,
		args *entity.TaskSettingsRevertArgs) ([]*entity.Task, error)

	// EnsureStillLive checks that what the trial applied is what is running.
	// Returning an error refuses the confirmation.
	//
	// Required. Confirm refuses outright when it is nil, because the alternative -
	// skipping the check for a caller that forgot - is a confirmation granted on
	// no evidence at all. A caller with nothing to verify passes a function that
	// returns nil, which is a decision somebody made rather than one nobody did.
	//
	// It exists because the database is not evidence. Swarm undoes a service
	// update whose task never becomes healthy, restoring the previous command
	// without telling anybody, so a change can be recorded as applied while the
	// cluster quietly runs the one it replaced. The request reaching this endpoint
	// proves HivePaaS is reachable; it does not prove it is reachable *through the
	// change*, and after a rollback those are different statements.
	//
	// Ignored by RevertNow, which is undoing the change either way.
	EnsureStillLive func(ctx context.Context, db database.Tx,
		args *entity.TaskSettingsRevertArgs) error
}
