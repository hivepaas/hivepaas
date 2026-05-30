package entity

import (
	"time"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

var (
	FileUpsertingConflictCols = []string{"id"}
	FileUpsertingUpdateCols   = []string{"scope", "object_id", "type", "status", "key", "name", "path",
		"size", "mimetype", "storage_type", "deleted", "storage_id", "bucket", "update_ver",
		"updated_at", "deleted_at"}
)

type File struct {
	ID          string               `bun:",pk" json:"id"`
	Scope       base.ObjectScopeType `json:"scope"`
	ObjectID    string               `bun:",nullzero" json:"objectId,omitempty"`
	Type        base.FileType        `json:"type"`
	Status      base.FileStatus      `json:"status"`
	Key         string               `bun:",nullzero" json:"key,omitempty"`
	Name        string               `bun:",nullzero" json:"name"`
	Path        string               `json:"path"`
	Size        int64                `json:"size"`
	Mimetype    string               `bun:",nullzero" json:"mimetype"`
	Deleted     bool                 `bun:",nullzero" json:"deleted,omitempty"`
	StorageType base.FileStorageType `json:"storageType"`
	StorageID   string               `bun:",nullzero" json:"storageId,omitempty"`
	Bucket      string               `bun:",nullzero" json:"bucket,omitempty"`
	UpdateVer   int                  `json:"updateVer"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`

	Storage         *Setting `bun:"rel:has-one,join:storage_id=id" json:"storage,omitempty"`
	BelongToUser    *User    `bun:"rel:belongs-to,join:object_id=id" json:"belongToUser,omitempty"`
	BelongToProject *Project `bun:"rel:belongs-to,join:object_id=id" json:"belongToProject,omitempty"`
	BelongToApp     *App     `bun:"rel:belongs-to,join:object_id=id" json:"belongToApp,omitempty"`
}

// GetID implements IDEntity interface
func (f *File) GetID() string {
	return f.ID
}

// GetName implements NamedEntity interface
func (f *File) GetName() string {
	return f.Name
}

func (f *File) IsActive() bool {
	return f.Status == base.FileStatusActive
}

func (f *File) IsInLocalStorage() bool {
	return f.StorageType == base.FileStorageLocal
}

func (f *File) IsInCloudStorage() bool {
	return f.StorageType == base.FileStorageCloud
}
