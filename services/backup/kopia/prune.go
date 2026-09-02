package kopia

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

func (c *Client) Prune(
	ctx context.Context,
	policy *backupmodel.RetentionPolicy,
) (res backupmodel.PruneResult, err error) {
	if policy != nil {
		args := []string{cmdPolicy, "set", cmdFlagGlobal}
		if policy.KeepLast > 0 {
			args = append(args, "--keep-latest="+strconv.Itoa(policy.KeepLast))
		}
		if policy.KeepDaily > 0 {
			args = append(args, "--keep-daily="+strconv.Itoa(policy.KeepDaily))
		}
		if policy.KeepWeekly > 0 {
			args = append(args, "--keep-weekly="+strconv.Itoa(policy.KeepWeekly))
		}
		if policy.KeepMonthly > 0 {
			args = append(args, "--keep-monthly="+strconv.Itoa(policy.KeepMonthly))
		}

		_, _ = c.execCommand(ctx, args)
	}

	var errBuf bytes.Buffer
	_, err = c.execCommand(ctx, []string{"maintenance", "run", "--full"}, func(o *execOptions) {
		o.stderr = &errBuf
	})
	if err != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		if errMsg != "" {
			return res, hperrors.Wrap(fmt.Errorf("kopia maintenance run failed: %s (err: %w)", errMsg, err))
		}
		return res, hperrors.Wrap(fmt.Errorf("kopia maintenance run failed: %w", err))
	}
	return res, nil
}
