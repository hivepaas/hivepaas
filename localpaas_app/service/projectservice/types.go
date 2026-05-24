package projectservice

import "github.com/localpaas/localpaas/localpaas_app/entity"

type PersistingProjectData struct {
	UpsertingProjects       []*entity.Project
	UpsertingApps           []*entity.App
	UpsertingTags           []*entity.ProjectTag
	UpsertingSettings       []*entity.Setting
	UpsertingACLPermissions []*entity.ACLPermission

	ProjectsToDeleteTags []string
}
