package kopia

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

// RestoreDirectory restores the snapshot, or the directory opts.Path of it, into
// targetDir. What is there already is overwritten where the snapshot has it, and
// kept where it has not: kopia never deletes.
func (c *Client) RestoreDirectory(
	ctx context.Context,
	snapshotID string,
	targetDir string,
	opts *backupmodel.RestoreOptions,
) (res backupmodel.RestoreResult, err error) {
	inside := ""
	if opts != nil {
		inside = opts.Path
	}
	_, err = c.execCommand(ctx, []string{cmdSnapshot, "restore", snapshotObjectPath(snapshotID, inside), targetDir})
	if err != nil {
		return res, hperrors.Wrap(err)
	}
	return res, nil
}

// RestoreStream writes the file filename of the snapshot to stdout: what a
// stream backup took, as it was taken.
//
// `show` reads an object, which a snapshot's ID is not: the file's object is
// found by listing the snapshot, which `ls` does by its ID.
func (c *Client) RestoreStream(
	ctx context.Context,
	snapshotID string,
	filename string,
	stdout io.Writer,
	_ *backupmodel.RestoreOptions,
) (res backupmodel.RestoreResult, err error) {
	dir, name := path.Split(strings.Trim(filename, "/"))
	entries, err := c.listEntries(ctx, snapshotID, dir)
	if err != nil {
		return res, hperrors.Wrap(err)
	}
	objectID := ""
	for _, entry := range entries {
		if entry.Name == name && !entry.Dir {
			objectID = entry.objectID
		}
	}
	if objectID == "" {
		return res, hperrors.Wrap(fmt.Errorf("%w: the snapshot holds no file %s",
			backupmodel.ErrSnapshotNotFound, filename))
	}
	_, err = c.execCommand(ctx, []string{"show", objectID},
		func(o *execOptions) {
			o.stdout = stdout
		})
	if err != nil {
		return res, hperrors.Wrap(err)
	}
	return res, nil
}
