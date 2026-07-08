package schedjobexecserviceimpl

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

func (s *service) calcCommandHelper(
	ctx context.Context,
	command *entity.CommandTemplate,
	taskID string,
	logStore *tasklog.Store,
) (cmd []string, err error) {
	if command == nil || (command.Command == "" && command.Script == "") { // can't continue if this happens
		_ = logStore.Add(ctx, tasklog.NewErrFrame(
			"Execution command/script is empty, aborted", tasklog.TsNow))
		return nil, apperrors.New(apperrors.ErrInternal).WithMsgLog("schedule job command/script is empty")
	}

	if command.Script != "" {
		encodedScript := base64.StdEncoding.EncodeToString(reflectutil.UnsafeStrToBytes(command.Script))
		tmpFilePath := fmt.Sprintf("/tmp/hivepaas_job_%s.sh", taskID)

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

func (s *service) calcCommand(
	ctx context.Context,
	data *execData,
) (cmd []string, err error) {
	cmd, err = s.calcCommandHelper(ctx, data.SchedJob.Command, data.Task.ID, data.LogStore)
	if err != nil {
		data.TaskNonRetryable = true
		return nil, apperrors.New(err)
	}
	return cmd, nil
}
