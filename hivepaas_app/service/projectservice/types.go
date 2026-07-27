package projectservice

import "github.com/hivepaas/hivepaas/hivepaas_app/entity"

type PersistingProjectData struct {
	UpsertingProjects       []*entity.Project
	UpsertingProjectEnvs    []*entity.ProjectEnv
	UpsertingApps           []*entity.App
	UpsertingTags           []*entity.Tag
	UpsertingSettings       []*entity.Setting
	UpsertingACLPermissions []*entity.ACLPermission

	ProjectsToDeleteTags []string
}
