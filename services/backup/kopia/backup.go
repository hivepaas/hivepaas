package kopia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

type kopiaSnapshotManifest struct {
	ID          string            `json:"id"`
	Source      kopiaSource       `json:"source"`
	StartTime   time.Time         `json:"startTime"`
	EndTime     time.Time         `json:"endTime"`
	Description string            `json:"description"`
	Tags        map[string]string `json:"tags"`
	Stats       kopiaStats        `json:"stats"`
}

type kopiaSource struct {
	Host     string `json:"host"`
	UserName string `json:"userName"`
	Path     string `json:"path"`
}

type kopiaStats struct {
	TotalSize      int64 `json:"totalSize"`
	TotalFileCount int64 `json:"totalFileCount"`
}

func (c *Client) BackupDirectory(
	ctx context.Context,
	dirPath string,
	opts *backupmodel.BackupOptions,
) (res backupmodel.BackupResult, err error) {
	args := []string{cmdSnapshot, cmdCreate, dirPath, cmdFlagJSON}
	args = append(args, c.formatBackupFlags(opts)...)

	var outBuf, errBuf bytes.Buffer
	_, err = c.execCommand(ctx, args, func(o *execOptions) {
		o.stdout = &outBuf
		o.stderr = &errBuf
	})
	if err != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		if errMsg != "" {
			return res, hperrors.Wrap(fmt.Errorf("kopia snapshot create failed: %s (err: %w)", errMsg, err))
		}
		return res, hperrors.Wrap(fmt.Errorf("kopia snapshot create failed: %w", err))
	}

	manifest, err := parseSnapshotOutput(outBuf.Bytes())
	if err != nil {
		return res, hperrors.Wrap(err)
	}
	res.Item = toStandardSnapshot(manifest)
	return res, nil
}

func (c *Client) BackupStream(
	ctx context.Context,
	stdin io.Reader,
	filename string,
	opts *backupmodel.BackupOptions,
) (res backupmodel.BackupResult, err error) {
	args := []string{cmdSnapshot, cmdCreate, "--stdin-file=" + filename, cmdFlagJSON}
	args = append(args, c.formatBackupFlags(opts)...)

	var outBuf, errBuf bytes.Buffer
	_, err = c.execCommand(ctx, args, func(o *execOptions) {
		o.stdin = stdin
		o.stdout = &outBuf
		o.stderr = &errBuf
	})
	if err != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		if errMsg != "" {
			return res, hperrors.Wrap(fmt.Errorf("kopia snapshot stream failed: %s (err: %w)", errMsg, err))
		}
		return res, hperrors.Wrap(fmt.Errorf("kopia snapshot stream failed: %w", err))
	}

	manifest, err := parseSnapshotOutput(outBuf.Bytes())
	if err != nil {
		return res, hperrors.Wrap(err)
	}
	res.Item = toStandardSnapshot(manifest)
	return res, nil
}

// parseSnapshotOutput reads the manifest that `snapshot create --json` prints. That command emits
// a single object, unlike `snapshot list --json` which emits an array, so both shapes are accepted.
func parseSnapshotOutput(output []byte) (*kopiaSnapshotManifest, error) {
	trimmed := bytes.TrimSpace(output)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("%w: kopia returned nothing", backupmodel.ErrSnapshotManifestInvalid)
	}

	if trimmed[0] == '[' {
		var manifests []kopiaSnapshotManifest
		if err := json.Unmarshal(trimmed, &manifests); err != nil {
			return nil, fmt.Errorf("%w: %w", backupmodel.ErrSnapshotManifestInvalid, err)
		}
		if len(manifests) == 0 {
			return nil, fmt.Errorf("%w: kopia returned an empty list", backupmodel.ErrSnapshotManifestInvalid)
		}
		return &manifests[0], nil
	}

	var manifest kopiaSnapshotManifest
	if err := json.Unmarshal(trimmed, &manifest); err != nil {
		return nil, fmt.Errorf("%w: %w", backupmodel.ErrSnapshotManifestInvalid, err)
	}
	return &manifest, nil
}

func (c *Client) formatBackupFlags(opts *backupmodel.BackupOptions) []string {
	var flags []string
	if opts == nil {
		return flags
	}

	for _, tag := range opts.Tags {
		if strings.TrimSpace(tag) != "" {
			flags = append(flags, "--tags="+tag)
		}
	}
	if opts.Hostname != "" {
		flags = append(flags, "--override-source="+opts.Hostname+":/")
	}

	return flags
}
