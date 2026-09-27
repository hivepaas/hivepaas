package entity

import (
	"encoding/json"
	"maps"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// SchedJobDataBackup is what a data-backup job backs up, and the repository it
// takes the snapshot into.
type SchedJobDataBackup struct {
	Source base.SchedJobDataBackupSource `json:"source"`
	// SourceCommand runs in the job's app; its stdout is the backup, the file
	// SourceFileName of the snapshot.
	SourceCommand  *CommandTemplate `json:"sourceCommand,omitempty"`
	SourceFileName string           `json:"sourceFileName,omitempty"`
	// SourceVolume is a volume the app mounts, and SourceVolumeSubpath a path
	// inside what the app sees of it; "" for all of it.
	SourceVolume        ObjectID `json:"sourceVolume,omitzero"`
	SourceVolumeSubpath string   `json:"sourceVolumeSubpath,omitempty"`
	TargetRepository    ObjectID `json:"targetRepository"`
	// Tags go onto every snapshot the job takes.
	Tags map[string]string `json:"tags,omitempty"`
}

func (b *SchedJobDataBackup) refSettingIDs() []string {
	if b == nil {
		return nil
	}
	var ids []string
	if b.SourceVolume.ID != "" {
		ids = append(ids, b.SourceVolume.ID)
	}
	if b.TargetRepository.ID != "" {
		ids = append(ids, b.TargetRepository.ID)
	}
	return ids
}

// Snapshot tags HivePaaS puts on every snapshot a data backup takes.
const (
	DataBackupTagJob = "hivepaas.job"
	DataBackupTagApp = "hivepaas.app"
)

// SnapshotTags are the tags of a snapshot the job takes, as kopia takes them
// (key:value): the job's and the app's, then the job's own, sorted.
func (b *SchedJobDataBackup) SnapshotTags(jobID, appID string) []string {
	tags := []string{DataBackupTagJob + ":" + jobID, DataBackupTagApp + ":" + appID}
	if b == nil {
		return tags
	}
	for _, key := range slices.Sorted(maps.Keys(b.Tags)) {
		tags = append(tags, key+":"+b.Tags[key])
	}
	return tags
}

// SchedJobDataBackupResult is what a data backup's run took: its task's output.
type SchedJobDataBackupResult struct {
	SnapshotID string `json:"snapshotId"`
	SizeBytes  int64  `json:"sizeBytes"`
}

// OutputAsDataBackup is the snapshot a data backup's run took; nil for a task
// that holds none, a job sequence's run among them.
func (t *Task) OutputAsDataBackup() (*SchedJobDataBackupResult, error) {
	if t.Output == "" {
		return nil, nil //nolint:nilnil // no output, no snapshot
	}
	// Read apart from the task's parsed output, which a job sequence's run
	// caches as its own type: the task API asks for both.
	result := &SchedJobDataBackupResult{}
	if err := json.Unmarshal([]byte(t.Output), result); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if result.SnapshotID == "" {
		return nil, nil //nolint:nilnil // no snapshot: another job's output
	}
	return result, nil
}
