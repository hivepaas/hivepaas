package commandpipeexecserviceimpl

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

func (s *service) calcCommand(
	ctx context.Context,
	command *entity.CommandTemplate,
	data *execData,
) (cmd []string, err error) {
	if command == nil || (command.Command == "" && !command.Script.IsValid()) {
		_ = data.LogStore.Add(ctx, tasklog.NewErrFrame(
			"Execution command/script is empty, aborted", tasklog.TsNow))
		return nil, apperrors.Wrap(apperrors.ErrInternal).WithMsgLog("command/script is empty")
	}

	if command.Script.IsValid() { //nolint:nestif
		script := command.Script.Value
		if script == "" && command.Script.ID != "" {
			scriptSetting := data.RefObjects.RefSettings[command.Script.ID]
			if scriptSetting == nil {
				return nil, apperrors.NewNotFound("Script object")
			}
			script = scriptSetting.MustAsScript().Data
		}

		encodedScript := base64.StdEncoding.EncodeToString(reflectutil.UnsafeStrToBytes(script))
		tmpFilePath := fmt.Sprintf("/tmp/hivepaas_pipe_%s.sh", data.Task.ID)

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
			return nil, apperrors.Wrap(err)
		}
	}
	return cmd, nil
}
