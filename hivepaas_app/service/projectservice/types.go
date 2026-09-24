package projectservice

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
)

type PersistingProjectData struct {
	UpsertingProjects       []*entity.Project
	UpsertingProjectEnvs    []*entity.ProjectEnv
	UpsertingApps           []*entity.App
	UpsertingTags           []*entity.Tag
	UpsertingSettings       []*entity.Setting
	UpsertingACLPermissions []*entity.ACLPermission

	ProjectsToDeleteTags []string
}

// NewProjectReq is a project being created, before anything of it is written.
type NewProjectReq struct {
	// Project has its ID, Key, Name, Note, Status and OwnerID set.
	Project *entity.Project
	Envs    []*NewProjectEnvReq
	Tags    []string
	TimeNow time.Time
}

type NewProjectEnvReq struct {
	Name  string
	Color string
}

// NewProjectEnv is the row of an env being added to a project, at a position
// among its envs. Its id and key are derived from its name, and never change.
func NewProjectEnv(project *entity.Project, name, color string, index int, now time.Time) *entity.ProjectEnv {
	return &entity.ProjectEnv{
		ID:        projecthelper.CalcProjectEnvID(project.ID, name),
		ProjectID: project.ID,
		Name:      name,
		Key:       projecthelper.CalcProjectEnvKey(name),
		Status:    base.ProjectStatusActive,
		Color:     color,
		Index:     index,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewProjectTags are a project's tag rows, numbered from startIndex.
func NewProjectTags(project *entity.Project, tags []string, startIndex int) []*entity.Tag {
	out := make([]*entity.Tag, 0, len(tags))
	for i, tag := range tags {
		out = append(out, &entity.Tag{ObjectID: project.ID, Tag: tag, Index: startIndex + i})
	}
	return out
}
