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
		if policy.KeepHourly > 0 {
			args = append(args, "--keep-hourly="+strconv.Itoa(policy.KeepHourly))
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

	// Setting the policy does not remove anything on its own, and neither does maintenance:
	// `snapshot expire` is what actually applies retention, and only with --delete - without it
	// the command is a dry run that reports what it would remove and changes nothing.
	var expireErrBuf bytes.Buffer
	_, err = c.execCommand(ctx, []string{cmdSnapshot, "expire", "--all", "--delete"},
		func(o *execOptions) {
			o.stderr = &expireErrBuf
		})
	if err != nil {
		return res, hperrors.Wrap(fmt.Errorf("%w: kopia snapshot expire: %s",
			backupmodel.ErrCommandFailed, strings.TrimSpace(expireErrBuf.String())))
	}

	// Maintenance reclaims the blobs the expired snapshots were the last reference to.
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
