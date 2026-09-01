package kopia

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

func (c *Client) RestoreDirectory(
	ctx context.Context,
	snapshotID string,
	targetDir string,
	opts *backupmodel.RestoreOptions,
) (res backupmodel.RestoreResult, err error) {
	var errBuf bytes.Buffer
	_, err = c.execCommand(ctx, []string{cmdSnapshot, "restore", snapshotID, targetDir},
		func(o *execOptions) {
			o.stderr = &errBuf
		})
	if err != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		if errMsg != "" {
			return res, hperrors.Wrap(fmt.Errorf("kopia restore failed: %s (err: %w)", errMsg, err))
		}
		return res, hperrors.Wrap(fmt.Errorf("kopia restore failed: %w", err))
	}
	return res, nil
}

func (c *Client) RestoreStream(
	ctx context.Context,
	snapshotID string,
	filename string,
	stdout io.Writer,
	opts *backupmodel.RestoreOptions,
) (res backupmodel.RestoreResult, err error) {
	var errBuf bytes.Buffer
	_, err = c.execCommand(ctx, []string{"show", snapshotID + "/" + filename},
		func(o *execOptions) {
			o.stdout = stdout
			o.stderr = &errBuf
		})
	if err != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		if errMsg != "" {
			return res, hperrors.Wrap(fmt.Errorf("kopia show dump failed: %s (err: %w)", errMsg, err))
		}
		return res, hperrors.Wrap(fmt.Errorf("kopia show dump failed: %w", err))
	}
	return res, nil
}
