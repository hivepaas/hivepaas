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

// SetRetention sets the global retention policy, every rule of it. A rule at zero keeps nothing on
// its own - and with every rule at zero kopia keeps every snapshot - while one left out would go on
// keeping what kopia's default, or the policy set before, said to keep.
func (c *Client) SetRetention(ctx context.Context, policy *backupmodel.RetentionPolicy) error {
	if policy == nil {
		return nil
	}
	var errBuf bytes.Buffer
	_, err := c.execCommand(ctx, []string{cmdPolicy, cmdSet, cmdFlagGlobal,
		"--keep-latest=" + strconv.Itoa(policy.KeepLast),
		"--keep-hourly=" + strconv.Itoa(policy.KeepHourly),
		"--keep-daily=" + strconv.Itoa(policy.KeepDaily),
		"--keep-weekly=" + strconv.Itoa(policy.KeepWeekly),
		"--keep-monthly=" + strconv.Itoa(policy.KeepMonthly),
	}, func(o *execOptions) {
		o.stderr = &errBuf
	})
	if err != nil {
		return hperrors.Wrap(fmt.Errorf("kopia policy set retention failed: %s (err: %w)",
			strings.TrimSpace(errBuf.String()), err))
	}
	return nil
}

func (c *Client) Prune(
	ctx context.Context,
	policy *backupmodel.RetentionPolicy,
) (res backupmodel.PruneResult, err error) {
	if err = c.SetRetention(ctx, policy); err != nil {
		return res, hperrors.Wrap(err)
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

	// Maintenance reclaims the blobs the expired snapshots were the last reference to. kopia runs
	// it for the repository's maintenance owner only: this client becomes it first.
	if res.MaintenanceTakenFrom, err = c.takeMaintenance(ctx); err != nil {
		return res, hperrors.Wrap(err)
	}
	var errBuf bytes.Buffer
	_, err = c.execCommand(ctx, []string{cmdMaintenance, "run", "--full"}, func(o *execOptions) {
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
