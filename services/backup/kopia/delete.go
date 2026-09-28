package kopia

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

func (c *Client) DeleteSnapshot(
	ctx context.Context,
	snapshotID string,
) (res backupmodel.DeleteSnapshotResult, err error) {
	var errBuf bytes.Buffer
	_, err = c.execCommand(ctx, []string{cmdSnapshot, "delete", snapshotID, "--delete"}, func(o *execOptions) {
		o.stderr = &errBuf
	})
	if err != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		// A snapshot already gone - deleted with kopia directly, or by an earlier
		// try - is what the caller asked for.
		if strings.Contains(errMsg, "no snapshots matched") {
			return res, hperrors.Wrap(fmt.Errorf("%w: %s", backupmodel.ErrSnapshotNotFound, snapshotID))
		}
		if errMsg != "" {
			return res, hperrors.Wrap(fmt.Errorf("kopia delete snapshot failed: %s (err: %w)", errMsg, err))
		}
		return res, hperrors.Wrap(fmt.Errorf("kopia delete snapshot failed: %w", err))
	}
	return res, nil
}
