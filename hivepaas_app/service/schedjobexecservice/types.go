package schedjobexecservice

import (
	"fmt"
	"io"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type SchedJobExecReq struct {
	*queue.TaskExecData
	SchedJobSetting        *entity.Setting
	DestApp                *entity.App
	TaskMinRunningDuration time.Duration
	TaskFindRetryMax       int
	TaskFindRetryDelay     time.Duration
	// Sequence is set when the job runs as a step of a job sequence.
	Sequence *SequenceStep
	// Command runs in place of the job's own: a data backup's source command.
	Command *entity.CommandTemplate
	// StdoutWriter takes the command's stdout in place of the job's command output,
	// without a TTY: a data backup streams it into its repository.
	StdoutWriter io.Writer
}

type RunCommandReq struct {
	*queue.TaskExecData
	App     *entity.App
	Command *entity.CommandTemplate
	// Stdin is what the command reads; its end is the command's input's end.
	Stdin io.Reader
}

type RunCommandResp struct {
	// ExitCode is the command's, when it ran to an end.
	ExitCode *int
}

type SchedJobExecResp struct {
	SkipResultNotification bool
	// ExitCode is the command's, when it ran to an end.
	ExitCode *int
	// Outputs is what a step of a sequence wrote to its output file.
	Outputs map[string]string
}

// SequenceStep is the part a job plays in a job sequence's run: which step it
// is, and how the steps before it went. The job is told so in its environment,
// and may write outputs for the steps after it to OutputFile.
type SequenceStep struct {
	Step       int // 1-based
	Steps      int
	Earlier    []*entity.SchedJobSeqStepResult
	OutputFile string
}

// OutputFilePath is where the step of a run writes its outputs, in its
// container.
func OutputFilePath(taskID string, step int) string {
	return fmt.Sprintf("/tmp/hivepaas-output-%s-%d", taskID, step)
}
