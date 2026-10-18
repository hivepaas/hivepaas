package schedjobexecserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
)

// outputFileFor is where a run writes its outputs, and the environment that
// tells it so: a sequence step's file, which sequenceEnv names, or for a job its
// app's engine made, a file of the run's own - its outputs are what the engine
// reads the app's state from. Any other run writes none.
func outputFileFor(req *schedjobexecservice.SchedJobExecReq, job *entity.SchedJob) (string, []string) {
	switch {
	case req.Sequence != nil:
		return req.Sequence.OutputFile, nil
	case job.IsManaged() && req.TaskExecData != nil && req.Task != nil:
		file := schedjobexecservice.OutputFilePath(req.Task.ID, 0)
		return file, []string{"HIVEPAAS_OUTPUT=" + file}
	}
	return "", nil
}
