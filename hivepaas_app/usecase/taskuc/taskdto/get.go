package taskdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
)

type GetTaskReq struct {
	Scope *entity.ObjectScope `json:"-" mapstructure:"-"`
	ID    string              `json:"-"`
}

func NewGetTaskReq() *GetTaskReq {
	return &GetTaskReq{}
}

func (req *GetTaskReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	// TODO: add validation
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetTaskResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *TaskResp     `json:"data"`
}

type TaskResp struct {
	ID        string             `json:"id"`
	Type      base.TaskType      `json:"type"`
	Status    base.TaskStatus    `json:"status"`
	Config    entity.TaskConfig  `json:"config"` // NOTE: use entity's type directly here, may need refactor
	TargetJob *TaskTargetJobResp `json:"targetJob"`
	LastError string             `json:"lastError"`
	UpdateVer int                `json:"updateVer"`
	// SequenceRun is a job sequence's run: each step and how it went.
	SequenceRun *entity.SchedJobSeqRun `json:"sequenceRun,omitempty" copy:"-"`
	// Trigger is what fired a scheduled job's run, when a trigger did.
	Trigger *TaskTriggerResp `json:"trigger,omitempty" copy:"-"`
	// DataBackup is the snapshot a data backup's run took.
	DataBackup *entity.SchedJobDataBackupResult `json:"dataBackup,omitempty" copy:"-"`
	// FunctionInvoke is the response a function's call got; its body is in
	// base64.
	FunctionInvoke *entity.SchedJobFunctionInvokeResult `json:"functionInvoke,omitempty" copy:"-"`
	// BackupRestore is what a restore's task restored, from where, and how.
	BackupRestore *TaskBackupRestoreResp `json:"backupRestore,omitempty" copy:"-"`

	ScopeProject *projectdto.ProjectBaseResp `json:"scopeProject,omitempty"`
	ScopeApp     *appdto.AppBaseResp         `json:"scopeApp,omitempty"`
	ScopeUser    *basedto.UserBaseResp       `json:"scopeUser,omitempty"`

	RunAt     *time.Time `json:"runAt" copy:",nilonzero"`
	RetryAt   *time.Time `json:"retryAt" copy:",nilonzero"`
	StartedAt *time.Time `json:"startedAt" copy:",nilonzero"`
	EndedAt   *time.Time `json:"endedAt" copy:",nilonzero"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TaskBackupRestoreResp is a restore, as it was asked for; the file or volume it
// went to is in its log.
type TaskBackupRestoreResp struct {
	// SnapshotRecordID is the record's; SnapshotID the repository's.
	SnapshotRecordID string                 `json:"snapshotRecordId"`
	RepoID           string                 `json:"repoId"`
	SnapshotID       string                 `json:"snapshotId"`
	SnapshotPath     string                 `json:"snapshotPath,omitempty"`
	FileName         string                 `json:"fileName,omitempty"`
	Mode             base.BackupRestoreMode `json:"mode,omitempty"`
	StopApp          bool                   `json:"stopApp,omitempty"`
}

type TaskTargetJobResp struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func TransformTask(
	task *entity.Task,
	taskInfo *cacheentity.TaskInfo,
	refObjects *entity.RefObjects,
) (resp *TaskResp, err error) {
	if err = copier.Copy(&resp, &task); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if taskInfo != nil {
		resp.Status = taskInfo.Status
		if taskInfo.Status == base.TaskStatusInProgress {
			resp.StartedAt = &taskInfo.StartedAt
		}
	}

	if resp.Status == base.TaskStatusInProgress || resp.Status == base.TaskStatusFailed {
		resp.LastError = task.GetLastError()
	}

	TransformTaskScopeObject(task, refObjects, resp)

	if task.Type == base.TaskTypeSchedJobExec {
		// A plain job's task has no output; one that does not parse is not shown
		// rather than failing the whole answer. A data backup's output reads as a
		// sequence run that never started, and a sequence's as no snapshot.
		if run, _ := task.OutputAsSchedJobSeqRun(); run != nil && run.Started {
			resp.SequenceRun = run
		}
		resp.DataBackup, _ = task.OutputAsDataBackup()
		resp.FunctionInvoke, _ = task.OutputAsFunctionInvoke()
		resp.Trigger = transformTaskTrigger(task, refObjects)
	}
	if task.Type == base.TaskTypeBackupRestore {
		// Arguments that do not parse are not shown rather than failing the answer.
		if args, _ := task.ArgsAsBackupRestore(); args != nil {
			resp.BackupRestore = &TaskBackupRestoreResp{
				SnapshotRecordID: task.TargetID, RepoID: args.RepoID, SnapshotID: args.SnapshotID,
				SnapshotPath: args.SnapshotPath, FileName: args.FileName, Mode: args.Mode, StopApp: args.StopApp,
			}
		}
	}

	return resp, nil
}

func TransformTaskScopeObject(
	task *entity.Task,
	refObjects *entity.RefObjects,
	resp *TaskResp,
) {
	if task.ObjectID == "" {
		return
	}

	var projectID, appID, userID string
	switch task.Scope {
	case base.ObjectScopeProject:
		projectID = task.ObjectID
	case base.ObjectScopeProjectEnv:
		projectID, _ = projecthelper.ParseProjectEnvID(task.ObjectID)
	case base.ObjectScopeApp:
		appID = task.ObjectID
	case base.ObjectScopeUser:
		userID = task.ObjectID
	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
	}

	if appID != "" {
		ref := refObjects.RefApps[appID]
		refResp := appdto.TransformAppBase(ref)
		if refResp == nil {
			refResp = appdto.NewMissingApp(appID)
		}
		resp.ScopeApp = refResp
		if ref != nil && projectID == "" {
			projectID = ref.ProjectID
		}
	}
	if projectID != "" {
		refResp := projectdto.TransformProjectBase(refObjects.RefProjects[projectID])
		if refResp == nil {
			refResp = projectdto.NewMissingProject(projectID)
		}
		resp.ScopeProject = refResp
	}
	if userID != "" {
		refResp := basedto.TransformUserBase(refObjects.RefUsers[userID])
		if refResp == nil {
			refResp = basedto.NewMissingUser(userID)
		}
		resp.ScopeUser = refResp
	}
}

// TaskTriggerResp is what fired a scheduled job's run.
type TaskTriggerResp struct {
	Event        base.SchedJobTriggerEvent `json:"event"`
	App          *basedto.NamedObjectResp  `json:"app"`
	DeploymentID string                    `json:"deploymentId,omitempty"`
}

func transformTaskTrigger(task *entity.Task, refObjects *entity.RefObjects) *TaskTriggerResp {
	args, err := task.ArgsAsSchedJobExec()
	if err != nil || args == nil || args.Trigger == nil {
		return nil
	}
	app := &basedto.NamedObjectResp{ID: args.Trigger.AppID}
	if refObjects != nil {
		if refApp := refObjects.RefApps[args.Trigger.AppID]; refApp != nil {
			app.Name = refApp.Name
		}
	}
	return &TaskTriggerResp{Event: args.Trigger.Event, App: app, DeploymentID: args.Trigger.DeploymentID}
}
