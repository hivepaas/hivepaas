package schedjobtriggerservice

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// TriggerInfo is what an event is about beyond its app.
type TriggerInfo struct {
	// DeploymentID is the deployment a deploy event is about.
	DeploymentID string
}

// FireResult is the runs an event started.
type FireResult struct {
	Runs []*Run
}

// Run is one job an event runs: its task, and whether a deploy waits for it.
type Run struct {
	Task    *entity.Task
	JobName string
	// Wait is true when a trigger that fired it holds the deploy.
	Wait bool
	// Timeout is how long a waiting deploy gives it: the job's timeout, or the
	// configured one for a job without a timeout and for a job sequence.
	Timeout time.Duration
}

// WaitedRuns are the runs a deploy waits for.
func (r *FireResult) WaitedRuns() []*Run {
	if r == nil {
		return nil
	}
	var waited []*Run
	for _, run := range r.Runs {
		if run.Wait {
			waited = append(waited, run)
		}
	}
	return waited
}
