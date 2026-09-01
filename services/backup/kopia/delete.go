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
		if errMsg != "" {
			return res, hperrors.Wrap(fmt.Errorf("kopia delete snapshot failed: %s (err: %w)", errMsg, err))
		}
		return res, hperrors.Wrap(fmt.Errorf("kopia delete snapshot failed: %w", err))
	}
	return res, nil
}
