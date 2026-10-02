package base

import "github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"

type SchedJobType string

const (
	SchedJobTypeContainerCommand  SchedJobType = "container-command"
	SchedJobTypeSystemCleanup     SchedJobType = "system-cleanup"
	SchedJobTypeSystemBackup      SchedJobType = "system-backup"
	SchedJobTypeSSLRenewal        SchedJobType = "ssl-renewal"
	SchedJobTypeBackupRepoCleanup SchedJobType = "backup-repo-cleanup"
	// SchedJobTypeJobSequence runs other scheduled jobs, one after another.
	SchedJobTypeJobSequence SchedJobType = "job-sequence"
	// SchedJobTypeDataBackup takes a snapshot of an app's data into a backup repository.
	SchedJobTypeDataBackup SchedJobType = "data-backup"
	// SchedJobTypeFunctionInvoke calls a function once: its runtime's invoke, in a
	// running task of the function.
	SchedJobTypeFunctionInvoke SchedJobType = "function-invoke"
	// SchedJobTypeRegistryAuthRenewal hands the services that pull from Amazon ECR
	// a token fresh enough to outlive the next run.
	SchedJobTypeRegistryAuthRenewal SchedJobType = "registry-auth-renewal"
)

var (
	AllSchedJobTypes = []SchedJobType{SchedJobTypeContainerCommand, SchedJobTypeSystemCleanup,
		SchedJobTypeSystemBackup, SchedJobTypeSSLRenewal, SchedJobTypeBackupRepoCleanup, SchedJobTypeJobSequence,
		SchedJobTypeDataBackup, SchedJobTypeFunctionInvoke, SchedJobTypeRegistryAuthRenewal}
)

// SchedJobDataBackupSource is what a data backup reads: a command's output, or a
// volume the app mounts.
type SchedJobDataBackupSource string

const (
	SchedJobDataBackupSourceCommand SchedJobDataBackupSource = "command"
	SchedJobDataBackupSourceVolume  SchedJobDataBackupSource = "volume"
)

var AllSchedJobDataBackupSources = []SchedJobDataBackupSource{SchedJobDataBackupSourceCommand,
	SchedJobDataBackupSourceVolume}

// BackupRestoreMode is how a restore writes a snapshot into a directory.
type BackupRestoreMode string

const (
	// BackupRestoreModeReplace moves the directory aside and restores into an
	// empty one: the state at backup time, and the old one kept.
	BackupRestoreModeReplace BackupRestoreMode = "replace"
	// BackupRestoreModeOverwrite writes the snapshot's files over what is there,
	// keeping what it does not have.
	BackupRestoreModeOverwrite BackupRestoreMode = "overwrite"
)

var AllBackupRestoreModes = []BackupRestoreMode{BackupRestoreModeReplace, BackupRestoreModeOverwrite}

// SchedJobSeqMode is how a job sequence runs its steps. Sequential is the only
// mode for now; the results are kept per step so a parallel one can follow.
type SchedJobSeqMode string

const (
	SchedJobSeqModeSequential SchedJobSeqMode = "sequential"
)

var AllSchedJobSeqModes = []SchedJobSeqMode{SchedJobSeqModeSequential}

// SchedJobSeqOnFailure is what a job sequence does when a step fails for good.
type SchedJobSeqOnFailure string

const (
	SchedJobSeqOnFailureStop     SchedJobSeqOnFailure = "stop"
	SchedJobSeqOnFailureContinue SchedJobSeqOnFailure = "continue"
)

var AllSchedJobSeqOnFailures = []SchedJobSeqOnFailure{SchedJobSeqOnFailureStop, SchedJobSeqOnFailureContinue}

// SchedJobSeqStepStatus is where one step of a job sequence's run stands.
type SchedJobSeqStepStatus string

const (
	SchedJobSeqStepPending SchedJobSeqStepStatus = "pending"
	SchedJobSeqStepRunning SchedJobSeqStepStatus = "running"
	SchedJobSeqStepDone    SchedJobSeqStepStatus = "done"
	SchedJobSeqStepFailed  SchedJobSeqStepStatus = "failed"
	SchedJobSeqStepSkipped SchedJobSeqStepStatus = "skipped"
)

// SchedJobSeqMaxSteps is the most steps a job sequence has.
const SchedJobSeqMaxSteps = 50

// SchedJobTriggerEvent is something that happens to an app and runs the
// scheduled jobs that listen to it.
type SchedJobTriggerEvent string

const (
	SchedJobTriggerPreDeploy    SchedJobTriggerEvent = "pre-deploy"
	SchedJobTriggerPostDeploy   SchedJobTriggerEvent = "post-deploy"
	SchedJobTriggerDeployFailed SchedJobTriggerEvent = "deploy-failed"
	SchedJobTriggerHealthDown   SchedJobTriggerEvent = "health-down"
	SchedJobTriggerHealthUp     SchedJobTriggerEvent = "health-up"
	SchedJobTriggerAppEnabled   SchedJobTriggerEvent = "app-enabled"
	SchedJobTriggerAppDisabled  SchedJobTriggerEvent = "app-disabled"
)

var AllSchedJobTriggerEvents = []SchedJobTriggerEvent{SchedJobTriggerPreDeploy, SchedJobTriggerPostDeploy,
	SchedJobTriggerDeployFailed, SchedJobTriggerHealthDown, SchedJobTriggerHealthUp, SchedJobTriggerAppEnabled,
	SchedJobTriggerAppDisabled}

// SchedJobMaxTriggers is the most triggers a scheduled job has.
const SchedJobMaxTriggers = 10

const (
	ExecCommandMaxSize = 300 * unit.KB
)
