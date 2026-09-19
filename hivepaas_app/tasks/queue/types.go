package queue

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

type TaskExecData struct {
	Task *entity.Task

	// RefObjects can be used as a cache to store objects
	RefObjects *entity.RefObjects
	LogStore   *tasklog.Store

	TaskNonCancelable bool
	TaskNonRetryable  bool
	TaskCanceled      bool
	TaskDone          bool
	CancelFunc        context.CancelFunc

	// Callback functions
	OnCommandFunc func(base.TaskCommand, ...any)
	OnEndTxFunc   func()
	OnPostTxFunc  func()

	// owner is the task whose transaction this one runs inside, when this data
	// belongs to work that is a task of its own but not a task of the queue's: a
	// workflow step, or the clone a preview makes. Only the queue calls these
	// callbacks, and only on the data it created, so a callback registered on
	// such a child has to reach the owner or it is never called at all. See
	// SubTask, which is the only way to set this.
	owner *TaskExecData
}

// SubTask is exec data for work that is a task of its own but runs inside this
// one: a workflow step, the clone a preview makes, the image build inside a
// deployment. It carries this task's references and log store, and anything
// registered on it through OnCommand, OnEndTx or OnPostTx is
// registered on this task - because the transaction, and its end, are this
// task's, and nothing else will ever call the child's.
//
// Build child exec data with this rather than with a literal. A literal's
// callbacks go nowhere, and say nothing about it.
func (t *TaskExecData) SubTask(task *entity.Task) *TaskExecData {
	if t == nil {
		return &TaskExecData{Task: task}
	}
	return &TaskExecData{
		Task:       task,
		RefObjects: t.RefObjects,
		LogStore:   t.LogStore,
		owner:      t,
	}
}

func (t *TaskExecData) IsTaskCanceled() bool {
	return t.TaskCanceled
}

func (t *TaskExecData) IsTaskDone() bool {
	return t.TaskDone
}

func (t *TaskExecData) AddRefObjects(refObjects *entity.RefObjects) {
	if t.RefObjects == nil {
		t.RefObjects = refObjects
	} else {
		t.RefObjects.AddRefObjects(refObjects)
	}
}

func (t *TaskExecData) OnCommand(fn func(base.TaskCommand, ...any)) {
	if t.owner != nil {
		t.owner.OnCommand(fn)
		return
	}
	if t.OnCommandFunc == nil {
		t.OnCommandFunc = fn
		return
	}
	currFunc := t.OnCommandFunc
	t.OnCommandFunc = func(cmd base.TaskCommand, args ...any) {
		currFunc(cmd, args...)
		fn(cmd, args...)
	}
}

func (t *TaskExecData) OnEndTx(fn func()) {
	if t.owner != nil {
		t.owner.OnEndTx(fn)
		return
	}
	if fn == nil {
		return
	}
	if t.OnEndTxFunc == nil {
		t.OnEndTxFunc = fn
		return
	}
	currFunc := t.OnEndTxFunc
	t.OnEndTxFunc = func() {
		currFunc()
		fn()
	}
}

func (t *TaskExecData) OnPostTx(fn func()) {
	if t.owner != nil {
		t.owner.OnPostTx(fn)
		return
	}
	if fn == nil {
		return
	}
	if t.OnPostTxFunc == nil {
		t.OnPostTxFunc = fn
		return
	}
	currFunc := t.OnPostTxFunc
	t.OnPostTxFunc = func() {
		currFunc()
		fn()
	}
}

type TaskExecFunc func(context.Context, database.Tx, *TaskExecData) error

type PeriodicExecData struct {
	PeriodicSetting *entity.Setting
	Task            *entity.Task
	Scope           *entity.ObjectScope

	// RefObjects can be used as a store of objects
	RefObjects *entity.RefObjects

	// SaveTask save task to DB if true, the executor should set this value
	SaveTask bool
}

func (t *PeriodicExecData) AddRefObjects(refObjects *entity.RefObjects) {
	if t.RefObjects == nil {
		t.RefObjects = refObjects
	} else {
		t.RefObjects.AddRefObjects(refObjects)
	}
}

type PeriodicExecFunc func(context.Context, *PeriodicExecData) error
