package kopia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

const bytesPerMB = 1024 * 1024

// kopiaRepoStatus is the part of `repository status --json` that matters here. Pack size is a
// repository format parameter, so it is the same for every client that connects.
type kopiaRepoStatus struct {
	ContentFormat struct {
		MaxPackSize int64 `json:"maxPackSize"`
	} `json:"contentFormat"`
}

// kopiaGlobalPolicy is the part of `policy get --global --json` that matters here.
type kopiaGlobalPolicy struct {
	Compression struct {
		CompressorName string `json:"compressorName"`
	} `json:"compression"`
	Retention struct {
		KeepLatest  int `json:"keepLatest"`
		KeepHourly  int `json:"keepHourly"`
		KeepDaily   int `json:"keepDaily"`
		KeepWeekly  int `json:"keepWeekly"`
		KeepMonthly int `json:"keepMonthly"`
	} `json:"retention"`
}

func (c *Client) ReadRepoConfig(ctx context.Context) (res backupmodel.RepoConfig, err error) {
	var status kopiaRepoStatus
	if err := c.readJSON(ctx, []string{cmdRepository, "status", cmdFlagJSON}, &status); err != nil {
		return res, hperrors.Wrap(err)
	}
	res.PackSizeMB = int(status.ContentFormat.MaxPackSize / bytesPerMB)

	var policy kopiaGlobalPolicy
	if err := c.readJSON(ctx, []string{cmdPolicy, "get", cmdFlagGlobal, cmdFlagJSON}, &policy); err != nil {
		return res, hperrors.Wrap(err)
	}
	res.Compression = policy.Compression.CompressorName
	res.Retention = &backupmodel.RetentionPolicy{
		KeepLast:    policy.Retention.KeepLatest,
		KeepHourly:  policy.Retention.KeepHourly,
		KeepDaily:   policy.Retention.KeepDaily,
		KeepWeekly:  policy.Retention.KeepWeekly,
		KeepMonthly: policy.Retention.KeepMonthly,
	}

	return res, nil
}

func (c *Client) readJSON(ctx context.Context, args []string, out any) error {
	var outBuf bytes.Buffer
	_, err := c.execCommand(ctx, args, func(o *execOptions) {
		o.stdout = &outBuf
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	if err := json.Unmarshal(outBuf.Bytes(), out); err != nil {
		return hperrors.Wrap(fmt.Errorf("%w: %w", backupmodel.ErrRepoConfigUnreadable, err))
	}
	return nil
}
