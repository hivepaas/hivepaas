package taskschedjobexec

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// managedSnapshotOutputs is a managed data backup's run's output: the snapshot
// it took, which its run's page links, beside the outputs its engine reads.
type managedSnapshotOutputs struct {
	*entity.SchedJobDataBackupResult
	Outputs map[string]string `json:"outputs"`
}

// keepManagedOutputs makes a managed job's outputs its run's output, whether the
// run failed or not: a health check that fails says why in them, and the run is
// saved either way. Any other job's run keeps what it had.
func keepManagedOutputs(task *entity.Task, job *entity.SchedJob, result *jobResult) {
	if !job.IsManaged() || result == nil || len(result.outputs) == 0 {
		return
	}
	if snapshot, err := task.OutputAsDataBackup(); err == nil && snapshot != nil {
		task.MustSetOutput(&managedSnapshotOutputs{SchedJobDataBackupResult: snapshot, Outputs: result.outputs})
		return
	}
	task.MustSetOutput(&entity.TaskSchedJobOutputs{Outputs: result.outputs})
}
