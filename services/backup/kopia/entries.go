package kopia

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

// entryLineRegex is a line of `kopia ls -l`: mode, size, date, time, zone,
// object ID, then the name, which may hold spaces.
var entryLineRegex = regexp.MustCompile(`^(\S+)\s+(\d+)\s+\S+\s+\S+\s+\S+\s+(\S+)\s+(.+)$`)

// listedEntry is an entry with the object that holds its content.
type listedEntry struct {
	backupmodel.SnapshotEntry
	objectID string
}

// ListEntries lists the directory path of the snapshot; "" for its root.
func (c *Client) ListEntries(
	ctx context.Context,
	snapshotID string,
	path string,
) ([]backupmodel.SnapshotEntry, error) {
	listed, err := c.listEntries(ctx, snapshotID, path)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	entries := make([]backupmodel.SnapshotEntry, 0, len(listed))
	for _, entry := range listed {
		entries = append(entries, entry.SnapshotEntry)
	}
	return entries, nil
}

func (c *Client) listEntries(ctx context.Context, snapshotID string, path string) ([]listedEntry, error) {
	var stdout bytes.Buffer
	_, err := c.execCommand(ctx, []string{"ls", "-l", snapshotObjectPath(snapshotID, path)},
		func(o *execOptions) {
			o.stdout = &stdout
		})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	entries, err := parseEntries(stdout.String())
	return entries, hperrors.Wrap(err)
}

// parseEntries reads what `kopia ls -l` prints, directories first, then by name.
func parseEntries(out string) ([]listedEntry, error) {
	var entries []listedEntry
	for line := range strings.Lines(out) {
		line = strings.TrimRight(line, "\r\n")
		if strings.TrimSpace(line) == "" {
			continue
		}
		match := entryLineRegex.FindStringSubmatch(line)
		if match == nil {
			return nil, fmt.Errorf("%w: unexpected kopia ls output: %q", backupmodel.ErrCommandFailed, line)
		}
		size, err := strconv.ParseInt(match[2], 10, 64)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		name := strings.TrimSpace(match[4])
		entries = append(entries, listedEntry{
			SnapshotEntry: backupmodel.SnapshotEntry{
				Name: strings.TrimSuffix(name, "/"), Dir: strings.HasPrefix(match[1], "d"), SizeBytes: size,
			},
			objectID: match[3],
		})
	}
	// Directories first, then by name: kopia's own order is not one to rely on.
	slices.SortStableFunc(entries, func(a, b listedEntry) int {
		if a.Dir != b.Dir {
			if a.Dir {
				return -1
			}
			return 1
		}
		return strings.Compare(a.Name, b.Name)
	})
	return entries, nil
}

// snapshotObjectPath is a path inside a snapshot, as `ls` and `snapshot
// restore` name it: the snapshot's ID, then the path.
func snapshotObjectPath(snapshotID, path string) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return snapshotID
	}
	return snapshotID + "/" + path
}
