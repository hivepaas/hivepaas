package schedjobexecserviceimpl

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/executil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
)

func (s *service) schedJobExecCalcCommand(
	ctx context.Context,
	data *schedJobExecData,
) (cmd []string, err error) {
	command := data.SchedJob.Command
	if command == nil || (command.Command == "" && command.Script == "") { // can't continue if this happens
		data.TaskNonRetryable = true
		_ = data.LogStore.Add(ctx, tasklog.NewErrFrame(
			"Execution command/script is empty, aborted", tasklog.TsNow))
		return nil, apperrors.New(apperrors.ErrInternal).WithMsgLog("schedule job command/script is empty")
	}

	if command.Script != "" {
		encodedScript := base64.StdEncoding.EncodeToString(reflectutil.UnsafeStrToBytes(command.Script))
		tmpFilePath := fmt.Sprintf("/tmp/localpaas_job_%s.sh", data.Task.ID)

		// Sample command format constructed below:
		// sh -c "echo '<base64>' | base64 -d > script-file && chmod +x script-file && script-file; exit_code=$?; \
		// rm -f script-file; exit $exit_code"
		var sb strings.Builder
		sb.Grow(len(encodedScript) + len(tmpFilePath)*5 + 100) //nolint:mnd
		sb.WriteString("echo '")
		sb.WriteString(encodedScript)
		sb.WriteString("' | base64 -d > ")
		sb.WriteString(tmpFilePath)
		sb.WriteString(" && chmod +x ")
		sb.WriteString(tmpFilePath)
		sb.WriteString(" && ")
		sb.WriteString(tmpFilePath)
		sb.WriteString("; exit_code=$?; rm -f ")
		sb.WriteString(tmpFilePath)
		sb.WriteString("; exit $exit_code")

		cmd = []string{"sh", "-c", sb.String()}
	} else {
		cmd, err = executil.CmdSplit(command.Command)
		if err != nil {
			return nil, apperrors.New(err)
		}
	}
	return cmd, nil
}
