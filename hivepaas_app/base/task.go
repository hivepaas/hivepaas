package base

import "github.com/tiendc/gofn"

type TaskType string

const (
	TaskTypeDummy        TaskType = "task:dummy"
	TaskTypeAppDeploy    TaskType = "task:app-deploy"
	TaskTypeAppClone     TaskType = "task:app-clone"
	TaskTypeAppPreview   TaskType = "task:app-preview"
	TaskTypeSchedJobExec TaskType = "task:sched-job-exec"
	TaskTypePeriodicExec TaskType = "task:periodic-exec"
	TaskTypeSystemUpdate TaskType = "task:system-update"
	TaskTypeWorkflow     TaskType = "task:workflow"

	// TaskTypeSettingsRevert undoes a settings change that was never confirmed.
	// See usecase/system/hpappsettingsuc/routing_settings_probation.go.
	TaskTypeSettingsRevert TaskType = "task:settings-revert"

	// TaskTypeAppLabelsSweep pushes a confirmed proxy topology onto the apps that
	// were left out while the change was on trial.
	// See usecase/system/hpappsettingsuc/settings_probation_confirm.go.
	TaskTypeAppLabelsSweep TaskType = "task:app-labels-sweep"
)

var (
	AllTaskTypes = []TaskType{TaskTypeDummy, TaskTypeAppDeploy, TaskTypeAppClone,
		TaskTypeAppPreview, TaskTypeSchedJobExec, TaskTypePeriodicExec,
		TaskTypeSystemUpdate, TaskTypeWorkflow, TaskTypeSettingsRevert, TaskTypeAppLabelsSweep}

	// These are listing types for front-end to show
	AllGlobalTaskTypes   = gofn.Drop(AllTaskTypes, TaskTypeDummy)
	AllHivepaasTaskTypes = []TaskType{}
	AllProjectTaskTypes  = []TaskType{TaskTypeAppDeploy, TaskTypeAppClone,
		TaskTypeAppPreview, TaskTypeSchedJobExec, TaskTypePeriodicExec}
	AllAppTaskTypes  = AllProjectTaskTypes
	AllUserTaskTypes = []TaskType{}
)

type TaskStatus string

const (
	TaskStatusNotStarted TaskStatus = "not-started"
	TaskStatusInProgress TaskStatus = "in-progress"
	TaskStatusCanceled   TaskStatus = "canceled"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusFailed     TaskStatus = "failed"
)

var (
	AllTaskStatuses = []TaskStatus{TaskStatusNotStarted, TaskStatusInProgress, TaskStatusCanceled,
		TaskStatusDone, TaskStatusFailed}
	AllTaskSettableStatuses = []TaskStatus{TaskStatusCanceled}
)

type TaskPriority string

const (
	TaskPriorityLow      TaskPriority = "low"
	TaskPriorityDefault  TaskPriority = "default"
	TaskPriorityCritical TaskPriority = "critical"
)

var (
	AllTaskPriorities = []TaskPriority{TaskPriorityLow, TaskPriorityDefault, TaskPriorityCritical}

	//nolint:mnd
	mapPriorityValues = map[TaskPriority]int{
		TaskPriorityLow:      3,
		TaskPriorityDefault:  6,
		TaskPriorityCritical: 10,
	}
)

func (p TaskPriority) Cmp(priority TaskPriority) int {
	if priority == "" {
		priority = TaskPriorityDefault
	}
	if p == priority {
		return 0
	}
	return mapPriorityValues[p] - mapPriorityValues[priority]
}

type TaskCommand string

const (
	TaskCommandCancel TaskCommand = "cancel"
)

type TaskTargetType string

const (
	TaskTargetTypeSchedJob    TaskTargetType = "sched-job"
	TaskTargetTypePeriodicJob TaskTargetType = "periodic-job"
)

var (
	AllTaskTargetTypes = []TaskTargetType{TaskTargetTypeSchedJob, TaskTargetTypePeriodicJob}
)
