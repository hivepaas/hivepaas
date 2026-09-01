package kopia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

const (
	shortIDLen = 8
)

func (c *Client) ListSnapshots(
	ctx context.Context,
	opts *backupmodel.ListSnapshotsOptions,
) (res backupmodel.ListSnapshotsResult, err error) {
	var outBuf, errBuf bytes.Buffer
	_, err = c.execCommand(ctx, []string{cmdSnapshot, cmdList, cmdFlagJSON, "--all"}, func(o *execOptions) {
		o.stdout = &outBuf
		o.stderr = &errBuf
	})
	if err != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		if errMsg != "" {
			return res, hperrors.Wrap(fmt.Errorf("kopia snapshot list failed: %s (err: %w)", errMsg, err))
		}
		return res, hperrors.Wrap(fmt.Errorf("kopia snapshot list failed: %w", err))
	}

	var rawManifests []kopiaSnapshotManifest
	if err := json.Unmarshal(outBuf.Bytes(), &rawManifests); err != nil {
		return res, hperrors.Wrap(fmt.Errorf("failed to parse kopia snapshots json: %w", err))
	}

	var result []*backupmodel.Snapshot
	for _, m := range rawManifests {
		snap := toStandardSnapshot(&m)
		if opts != nil {
			if len(opts.Tags) > 0 && !hasMatchingTag(snap.Tags, opts.Tags) {
				continue
			}
			if opts.Hostname != "" && snap.Hostname != opts.Hostname {
				continue
			}
			if opts.Path != "" && !hasMatchingPath(snap.Paths, opts.Path) {
				continue
			}
			if opts.Since != nil && snap.Time.Before(*opts.Since) {
				continue
			}
			if opts.Until != nil && snap.Time.After(*opts.Until) {
				continue
			}
		}
		result = append(result, snap)
		if opts != nil && opts.Limit > 0 && len(result) >= opts.Limit {
			break
		}
	}

	res.Items = result
	return res, nil
}

func (c *Client) GetSnapshot(
	ctx context.Context,
	snapshotID string,
) (res backupmodel.GetSnapshotResult, err error) {
	var outBuf, errBuf bytes.Buffer
	_, err = c.execCommand(ctx, []string{cmdSnapshot, cmdList, snapshotID, cmdFlagJSON}, func(o *execOptions) {
		o.stdout = &outBuf
		o.stderr = &errBuf
	})
	if err != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		if errMsg != "" {
			return res, hperrors.Wrap(fmt.Errorf("kopia get snapshot failed: %s (err: %w)", errMsg, err))
		}
		return res, hperrors.Wrap(fmt.Errorf("kopia get snapshot failed: %w", err))
	}

	var rawManifests []kopiaSnapshotManifest
	if err := json.Unmarshal(outBuf.Bytes(), &rawManifests); err != nil || len(rawManifests) == 0 {
		return res, hperrors.Wrap(backupmodel.ErrSnapshotNotFound).WithNTParam("Name", snapshotID)
	}

	res.Item = toStandardSnapshot(&rawManifests[0])
	return res, nil
}

// kopiaUserTagPrefix is how kopia namespaces user-supplied tags inside a snapshot manifest.
const kopiaUserTagPrefix = "tag:"

func toStandardSnapshot(m *kopiaSnapshotManifest) *backupmodel.Snapshot {
	var tags []string
	for k, v := range m.Tags {
		// Kopia namespaces user tags as "tag:<key>" in the manifest. Store the key the user wrote.
		k = strings.TrimPrefix(k, kopiaUserTagPrefix)
		if v != "" {
			tags = append(tags, k+":"+v)
		} else {
			tags = append(tags, k)
		}
	}
	sort.Strings(tags)

	shortID := m.ID
	if len(shortID) > shortIDLen {
		shortID = shortID[:shortIDLen]
	}

	return &backupmodel.Snapshot{
		ID:        m.ID,
		ShortID:   shortID,
		Time:      m.StartTime,
		Tags:      tags,
		Paths:     []string{m.Source.Path},
		Hostname:  m.Source.Host,
		SizeBytes: m.Stats.TotalSize,
	}
}

func hasMatchingTag(snapshotTags []string, filterTags []string) bool {
	tagMap := make(map[string]bool, len(snapshotTags))
	for _, t := range snapshotTags {
		tagMap[t] = true
	}
	for _, f := range filterTags {
		if tagMap[f] {
			return true
		}
	}
	return false
}

func hasMatchingPath(paths []string, targetPath string) bool {
	cleanTarget := strings.TrimRight(targetPath, "/")
	for _, p := range paths {
		if strings.TrimRight(p, "/") == cleanTarget {
			return true
		}
	}
	return false
}
