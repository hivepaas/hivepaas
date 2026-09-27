package schedjobexecserviceimpl

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
)

const (
	// outputsMaxSize is the most of a step's output file that is read.
	outputsMaxSize = 64 * 1024
)

var outputKeyRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type seqStepResultEnv struct {
	Job        string `json:"job"`
	Name       string `json:"name,omitempty"`
	Status     string `json:"status"`
	ExitCode   *int   `json:"exitCode,omitempty"`
	Error      string `json:"error,omitempty"`
	DurationMs int64  `json:"durationMs,omitempty"`
}

type seqStepOutputsEnv struct {
	Job     string            `json:"job"`
	Name    string            `json:"name,omitempty"`
	Outputs map[string]string `json:"outputs"`
}

// sequenceEnv is what a step of a job sequence is told: which step it is, how
// the earlier steps went, what they output, and where to write its own
// outputs. Nothing outside a sequence. HIVEPAAS_, not HP_: HP_ is the app's
// own configuration.
func sequenceEnv(step *schedjobexecservice.SequenceStep) []string {
	if step == nil {
		return nil
	}
	results := make([]*seqStepResultEnv, 0, len(step.Earlier))
	outputs := make([]*seqStepOutputsEnv, 0, len(step.Earlier))
	for _, earlier := range step.Earlier {
		result := &seqStepResultEnv{
			Job:      earlier.Job.ID,
			Name:     earlier.Name,
			Status:   string(earlier.Status),
			ExitCode: earlier.ExitCode,
			Error:    earlier.Error,
		}
		if !earlier.StartedAt.IsZero() && !earlier.EndedAt.IsZero() {
			result.DurationMs = earlier.EndedAt.Sub(earlier.StartedAt).Milliseconds()
		}
		results = append(results, result)
		stepOutputs := earlier.Outputs
		if stepOutputs == nil {
			stepOutputs = map[string]string{}
		}
		outputs = append(outputs, &seqStepOutputsEnv{Job: earlier.Job.ID, Name: earlier.Name, Outputs: stepOutputs})
	}
	resultsJSON, _ := json.Marshal(results)
	outputsJSON, _ := json.Marshal(outputs)

	env := []string{
		"HIVEPAAS_SEQ_STEP=" + strconv.Itoa(step.Step),
		"HIVEPAAS_SEQ_STEPS=" + strconv.Itoa(step.Steps),
		"HIVEPAAS_SEQ_RESULTS=" + string(resultsJSON),
		"HIVEPAAS_SEQ_OUTPUTS=" + string(outputsJSON),
		"HIVEPAAS_OUTPUT=" + step.OutputFile,
	}
	var prev *entity.SchedJobSeqStepResult
	if n := len(step.Earlier); n > 0 {
		prev = step.Earlier[n-1]
	}
	if prev == nil {
		return append(env, "HIVEPAAS_SEQ_PREV_STATUS=")
	}
	env = append(env, "HIVEPAAS_SEQ_PREV_STATUS="+string(prev.Status))
	for key, value := range prev.Outputs {
		env = append(env, "HIVEPAAS_SEQ_OUTPUT_"+strings.ToUpper(key)+"="+value)
	}
	return env
}

// parseOutputs reads a step's output file: KEY=value lines, the value as it
// stands after the first '='. Blank lines and # comments are skipped; a line
// that breaks the rules is left out with a warning, and so is what lies past
// the first 64 KB.
func parseOutputs(data []byte) (map[string]string, []string) {
	var warnings []string
	tooLong := len(data) > outputsMaxSize
	if tooLong {
		data = data[:outputsMaxSize]
		if i := bytes.LastIndexByte(data, '\n'); i >= 0 {
			data = data[:i+1]
		}
	}
	outputs := map[string]string{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(nil, outputsMaxSize+1)
	for n := 1; scanner.Scan(); n++ {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found || !outputKeyRegex.MatchString(key) {
			warnings = append(warnings, fmt.Sprintf("output line %d is not KEY=value with a KEY of letters, "+
				"digits and _, not starting with a digit; left out", n))
			continue
		}
		outputs[key] = value
	}
	if tooLong {
		warnings = append(warnings, "the output file is over 64 KB; what follows the first 64 KB is left out")
	}
	return outputs, warnings
}

// readOutputs reads and deletes a step's output file, with a second command in
// the container the step ran in. A file that cannot be read - no shell in the
// container, a container gone - is a warning and no outputs, never a failed
// step.
func (s *service) readOutputs(
	ctx context.Context,
	req *schedjobexecservice.SchedJobExecReq,
	execResp *containerexecservice.ContainerExecResp,
) map[string]string {
	if execResp == nil || execResp.ContainerID == "" {
		return nil
	}
	buf := &limitedBuffer{max: outputsMaxSize + 1}
	_, err := s.containerExecService.ContainerExec(ctx, &containerexecservice.ContainerExecReq{
		App:          req.DestApp,
		ContainerID:  execResp.ContainerID,
		NodeID:       execResp.NodeID,
		LogStore:     tasklog.NewNullStore(),
		StdoutWriter: buf,
		ExecOptions: func(opts *client.ExecCreateOptions) {
			opts.AttachStdout = true
			opts.AttachStderr = true
			opts.Cmd = []string{"sh", "-c", `cat "$1" 2>/dev/null; rm -f "$1"`, "sh", req.Sequence.OutputFile}
		},
	})
	if err != nil {
		_ = req.LogStore.Add(ctx, tasklog.NewWarnFrame(
			"The step's outputs could not be read: "+err.Error()+"\n", tasklog.TsNow))
		return nil
	}
	outputs, warnings := parseOutputs(buf.Bytes())
	for _, warning := range warnings {
		_ = req.LogStore.Add(ctx, tasklog.NewWarnFrame("Outputs: "+warning+"\n", tasklog.TsNow))
	}
	return outputs
}

// limitedBuffer keeps what is written to it up to max bytes; past that it
// keeps nothing more, which parseOutputs then reports.
type limitedBuffer struct {
	bytes.Buffer
	max int
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	room := b.max - b.Len()
	if room <= 0 {
		return len(p), nil
	}
	if len(p) > room {
		_, _ = b.Buffer.Write(p[:room])
		return len(p), nil
	}
	return b.Buffer.Write(p) //nolint:wrapcheck // bytes.Buffer never fails
}
