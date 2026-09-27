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
)

var (
	AllSchedJobTypes = []SchedJobType{SchedJobTypeContainerCommand, SchedJobTypeSystemCleanup,
		SchedJobTypeSystemBackup, SchedJobTypeSSLRenewal, SchedJobTypeBackupRepoCleanup, SchedJobTypeJobSequence}
)

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

const (
	ExecCommandMaxSize = 300 * unit.KB
)
