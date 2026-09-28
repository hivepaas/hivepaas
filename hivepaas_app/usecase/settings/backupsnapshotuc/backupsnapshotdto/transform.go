package backupsnapshotdto

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

// BackupSnapshotResp is a snapshot a repository holds, with where it came from.
type BackupSnapshotResp struct {
	// ID is the record's; SnapshotID the repository's.
	ID          string    `json:"id"`
	SnapshotID  string    `json:"snapshotId"`
	ShortID     string    `json:"shortId"`
	Time        time.Time `json:"time"`
	SizeBytes   int64     `json:"sizeBytes"`
	Description string    `json:"description,omitempty"`
	Paths       []string  `json:"paths"`
	Hostname    string    `json:"hostname,omitempty"`
	Tags        []string  `json:"tags"`
	// Source is command or volume; empty when not known.
	Source base.SchedJobDataBackupSource `json:"source,omitempty"`
	Repo   *settings.BaseSettingResp     `json:"repo"`
	App    *SnapshotAppResp              `json:"app,omitempty"`
	Job    *SnapshotJobResp              `json:"job,omitempty"`
	// RunID is the task of the run that took the snapshot.
	RunID string `json:"runId,omitempty"`
}

// SnapshotAppResp is the app a snapshot is of; Deleted when it is gone.
type SnapshotAppResp struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Env     string `json:"env,omitempty"`
	Deleted bool   `json:"deleted,omitempty"`
}

// SnapshotJobResp is the job that took a snapshot; Deleted when it is gone.
type SnapshotJobResp struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Deleted bool   `json:"deleted,omitempty"`
}

// SnapshotRefs are what snapshots name, by ID.
type SnapshotRefs struct {
	Repos map[string]*entity.Setting
	Apps  map[string]*entity.App
	Jobs  map[string]*entity.Setting
}

// TransformBackupSnapshot is a snapshot record with its tags, named from refs.
func TransformBackupSnapshot(record *entity.Setting, tags []string, refs *SnapshotRefs) *BackupSnapshotResp {
	snapshot, _ := record.AsBackupSnapshot()
	if snapshot == nil {
		snapshot = &entity.BackupSnapshot{}
	}
	if tags == nil {
		tags = []string{}
	}
	paths := snapshot.Paths
	if paths == nil {
		paths = []string{}
	}
	parsed := entity.ParseDataBackupSnapshotTags(tags)
	resp := &BackupSnapshotResp{
		ID:          record.ID,
		SnapshotID:  snapshot.ID,
		ShortID:     snapshot.ShortID,
		Time:        snapshot.Time,
		SizeBytes:   snapshot.SizeBytes,
		Description: snapshot.Description,
		Paths:       paths,
		Hostname:    snapshot.Hostname,
		Tags:        tags,
		Source:      parsed.Source,
		RunID:       parsed.RunID,
	}

	repoResp, _ := settings.TransformSettingBase(refs.Repos[record.RefID])
	if repoResp == nil {
		repoResp = settings.NewMissingSetting(record.RefID, base.SettingTypeBackupRepo)
	}
	resp.Repo = repoResp

	if parsed.AppID != "" {
		resp.App = &SnapshotAppResp{ID: parsed.AppID, Deleted: true}
		if app := refs.Apps[parsed.AppID]; app != nil && app.DeletedAt.IsZero() {
			_, env := projecthelper.ParseProjectEnvID(app.ProjectEnvID)
			resp.App = &SnapshotAppResp{ID: app.ID, Name: app.Name, Env: env}
		}
	}
	if parsed.JobID != "" {
		resp.Job = &SnapshotJobResp{ID: parsed.JobID, Deleted: true}
		if job := refs.Jobs[parsed.JobID]; job != nil && job.DeletedAt.IsZero() {
			resp.Job = &SnapshotJobResp{ID: job.ID, Name: job.Name}
			// A snapshot taken before the source tag: its job says.
			if resp.Source == "" {
				if schedJob, _ := job.AsSchedJob(); schedJob != nil && schedJob.DataBackup != nil {
					resp.Source = schedJob.DataBackup.Source
				}
			}
		}
	}
	return resp
}
