package entity

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentDeploymentVersion = 1
)

var (
	DeploymentUpsertingConflictCols = []string{"id"}
	DeploymentUpsertingUpdateCols   = []string{"app_id", "status", "settings", "trigger", "output",
		"version", "update_ver", "started_at", "ended_at", "updated_at", "deleted_at"}
)

type Deployment struct {
	ID        string                 `bun:",pk" json:"id"`
	AppID     string                 `json:"appId"`
	Status    base.DeploymentStatus  `json:"status"`
	Settings  *AppDeploymentSettings `json:"settings"`
	Trigger   *AppDeploymentTrigger  `json:"trigger"`
	Output    *AppDeploymentOutput   `json:"output"`
	Version   int                    `json:"version"`
	UpdateVer int                    `json:"updateVer"`

	StartedAt time.Time `bun:",nullzero" json:"startedAt"`
	EndedAt   time.Time `bun:",nullzero" json:"endedAt"`
	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`

	App   *App    `bun:"rel:belongs-to,join:app_id=id" json:"app,omitempty"`
	Tasks []*Task `bun:"rel:has-many,join:id=target_id" json:"tasks,omitempty"`
}

// GetID implements IDEntity interface
func (d *Deployment) GetID() string {
	return d.ID
}

func (d *Deployment) IsDone() bool {
	return d.Status == base.DeploymentStatusDone
}

func (d *Deployment) IsFailed() bool {
	return d.Status == base.DeploymentStatusFailed
}

func (d *Deployment) IsCanceled() bool {
	return d.Status == base.DeploymentStatusCanceled
}

func (d *Deployment) IsNotStarted() bool {
	return d.Status == base.DeploymentStatusNotStarted
}

func (d *Deployment) IsInProgress() bool {
	return d.Status == base.DeploymentStatusInProgress
}

func (d *Deployment) CanCancel() bool {
	if d.Status == base.DeploymentStatusDone ||
		d.Status == base.DeploymentStatusCanceled ||
		d.Status == base.DeploymentStatusFailed {
		return false
	}
	return true
}

func (d *Deployment) IsDeleted() bool {
	return !d.DeletedAt.IsZero()
}

func (d *Deployment) GetDuration() time.Duration {
	if d.EndedAt.IsZero() || d.EndedAt.Before(d.StartedAt) {
		return time.Since(d.StartedAt)
	}
	return d.EndedAt.Sub(d.StartedAt)
}

func (d *Deployment) GetRefObjectIDs() *RefObjectIDs {
	return d.Settings.GetRefObjectIDs()
}

type AppDeploymentTrigger struct {
	Source   base.DeploymentTriggerSource `json:"source"`
	SourceID string                       `json:"sourceId"`
	ChangeID string                       `json:"changeId"`
}

type AppDeploymentOutput struct {
	Error         string   `json:"error,omitempty"`
	CommitHash    string   `json:"commitHash,omitempty"`
	CommitTitle   string   `json:"commitTitle,omitempty"`
	CommitMessage string   `json:"commitMessage,omitempty"`
	CommitAuthor  string   `json:"commitAuthor,omitempty"`
	ImageTags     []string `json:"imageTags,omitempty"`
}
