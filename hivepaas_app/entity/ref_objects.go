package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func NewRefObjects() *RefObjects {
	return &RefObjects{
		RefSettings:    make(map[string]*Setting, 5), //nolint:mnd
		RefApps:        make(map[string]*App, 2),     //nolint:mnd
		RefProjects:    make(map[string]*Project),
		RefProjectEnvs: make(map[string]*ProjectEnv),
		RefUsers:       make(map[string]*User),
	}
}

type RefObjects struct {
	RefSettings    map[string]*Setting    `json:"settings"`
	RefApps        map[string]*App        `json:"apps"`
	RefProjects    map[string]*Project    `json:"projects"`
	RefProjectEnvs map[string]*ProjectEnv `json:"projectEnvs"`
	RefUsers       map[string]*User       `json:"users"`
}

func (r *RefObjects) AddRefObjects(refObjects *RefObjects) {
	if refObjects == nil {
		return
	}

	for _, refSetting := range refObjects.RefSettings {
		r.RefSettings[refSetting.ID] = refSetting
	}
	for _, refApp := range refObjects.RefApps {
		r.RefApps[refApp.ID] = refApp
	}
	for _, refProject := range refObjects.RefProjects {
		r.RefProjects[refProject.ID] = refProject
	}
	for _, refProjectEnv := range refObjects.RefProjectEnvs {
		r.RefProjectEnvs[refProjectEnv.ID] = refProjectEnv
	}
	for _, refUser := range refObjects.RefUsers {
		r.RefUsers[refUser.ID] = refUser
	}
}

func (r *RefObjects) AddObjectScope(scope *ObjectScope) {
	if scope == nil {
		return
	}
	if scope.App != nil {
		r.RefApps[scope.App.ID] = scope.App
	}
	if scope.ParentApp != nil {
		r.RefApps[scope.ParentApp.ID] = scope.ParentApp
	}
	if scope.Project != nil {
		r.RefProjects[scope.Project.ID] = scope.Project
	}
	if scope.ProjectEnv != nil {
		r.RefProjectEnvs[scope.ProjectEnv.ID] = scope.ProjectEnv
	}
	if scope.User != nil {
		r.RefUsers[scope.User.ID] = scope.User
	}
}

//nolint:gocognit
func (r *RefObjects) GetObjectScope(
	scope base.ObjectScopeType,
	objectID string,
	requireActive bool,
) (*ObjectScope, error) {
	switch scope {
	case base.ObjectScopeApp:
		app := r.RefApps[objectID]
		if app == nil {
			return nil, apperrors.Wrap(apperrors.ErrAppNotFound).WithParam("Name", objectID)
		}
		if app.ProjectEnv == nil {
			app.ProjectEnv = r.RefProjectEnvs[app.ProjectEnvID]
		}
		if app.Project == nil {
			app.Project = r.RefProjects[app.ProjectID]
		}
		if requireActive {
			if app.Status != base.AppStatusActive {
				return nil, apperrors.Wrap(apperrors.ErrAppInactive).WithParam("Name", app.Name)
			}
			if app.ProjectEnv == nil {
				return nil, apperrors.Wrap(apperrors.ErrProjectEnvNotFound).WithParam("Name", app.ProjectEnvID)
			}
			if app.ProjectEnv.Status != base.ProjectStatusActive {
				return nil, apperrors.Wrap(apperrors.ErrProjectEnvInactive).WithParam("Name", app.ProjectEnv.Name)
			}
			if app.Project == nil {
				return nil, apperrors.Wrap(apperrors.ErrProjectNotFound).WithParam("Name", app.ProjectID)
			}
			if app.Project.Status != base.ProjectStatusActive {
				return nil, apperrors.Wrap(apperrors.ErrProjectInactive).WithParam("Name", app.Project.Name)
			}
		}
		return app.GetObjectScope(), nil

	case base.ObjectScopeProject:
		project := r.RefProjects[objectID]
		if project == nil {
			return nil, apperrors.Wrap(apperrors.ErrProjectNotFound).WithParam("Name", objectID)
		}
		if requireActive {
			if project.Status != base.ProjectStatusActive {
				return nil, apperrors.Wrap(apperrors.ErrProjectInactive).WithParam("Name", project.Name)
			}
		}
		return project.GetObjectScope(), nil

	case base.ObjectScopeProjectEnv:
		projectEnv := r.RefProjectEnvs[objectID]
		if projectEnv == nil {
			return nil, apperrors.Wrap(apperrors.ErrProjectEnvNotFound).WithParam("Name", objectID)
		}
		if projectEnv.Project == nil {
			projectEnv.Project = r.RefProjects[projectEnv.ProjectID]
		}
		if requireActive {
			if projectEnv.Status != base.ProjectStatusActive {
				return nil, apperrors.Wrap(apperrors.ErrProjectEnvInactive).WithParam("Name", projectEnv.Name)
			}
			if projectEnv.Project == nil {
				return nil, apperrors.Wrap(apperrors.ErrProjectNotFound).WithParam("Name", projectEnv.ProjectID)
			}
			if projectEnv.Project.Status != base.ProjectStatusActive {
				return nil, apperrors.Wrap(apperrors.ErrProjectInactive).WithParam("Name", projectEnv.Project.Name)
			}
		}
		return projectEnv.GetObjectScope(), nil

	case base.ObjectScopeUser:
		user := r.RefUsers[objectID]
		if user == nil {
			return nil, apperrors.Wrap(apperrors.ErrUserNotFound).WithParam("Name", objectID)
		}
		if requireActive {
			if user.Status != base.UserStatusActive || user.IsAccessExpired() {
				return nil, apperrors.Wrap(apperrors.ErrUserUnavailable).
					WithParam("Name", gofn.Coalesce(user.FullName, user.Username))
			}
		}
		return user.GetObjectScope(), nil

	case base.ObjectScopeGlobal:
		return NewObjectScopeGlobal(), nil

	case base.ObjectScopeHivepaas:
		return NewObjectScopeHivepaas(), nil
	}
	return nil, nil
}

type RefObjectIDs struct {
	RefSettingIDs    []string `json:"settingIds"`
	RefAppIDs        []string `json:"appIds"`
	RefProjectIDs    []string `json:"projectIds"`
	RefProjectEnvIDs []string `json:"projectEnvIds"`
	RefUserIDs       []string `json:"userIds"`
}

func (r *RefObjectIDs) HasData() bool {
	return len(r.RefSettingIDs) > 0 || len(r.RefAppIDs) > 0 ||
		len(r.RefProjectIDs) > 0 || len(r.RefProjectEnvIDs) > 0 || len(r.RefUserIDs) > 0
}

func (r *RefObjectIDs) AddRefIDs(refIDs *RefObjectIDs) {
	if refIDs == nil {
		return
	}
	r.RefSettingIDs = append(r.RefSettingIDs, refIDs.RefSettingIDs...)
	r.RefAppIDs = append(r.RefAppIDs, refIDs.RefAppIDs...)
	r.RefProjectIDs = append(r.RefProjectIDs, refIDs.RefProjectIDs...)
	r.RefProjectEnvIDs = append(r.RefProjectEnvIDs, refIDs.RefProjectEnvIDs...)
	r.RefUserIDs = append(r.RefUserIDs, refIDs.RefUserIDs...)
}

func (r *RefObjectIDs) AddScopeObjectIDOfSettings(settings ...*Setting) {
	for _, s := range settings {
		if s.ObjectID == "" {
			continue
		}
		switch s.Scope {
		case base.ObjectScopeApp:
			r.RefAppIDs = append(r.RefAppIDs, s.ObjectID)
		case base.ObjectScopeProject:
			r.RefProjectIDs = append(r.RefProjectIDs, s.ObjectID)
		case base.ObjectScopeProjectEnv:
			r.RefProjectEnvIDs = append(r.RefProjectEnvIDs, s.ObjectID)
		case base.ObjectScopeUser:
			r.RefUserIDs = append(r.RefUserIDs, s.ObjectID)
		case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		}
	}
}

func (r *RefObjectIDs) GetRecursiveRefObjectIDs(refObjects *RefObjects) *RefObjectIDs {
	newRefIDs := &RefObjectIDs{}
	for _, setting := range refObjects.RefSettings {
		newRefIDs.AddRefIDs(setting.MustGetRefObjectIDs())
	}
	res := &RefObjectIDs{}
	for _, settingID := range newRefIDs.RefSettingIDs {
		if !gofn.Contain(r.RefSettingIDs, settingID) {
			res.RefSettingIDs = append(res.RefSettingIDs, settingID)
		}
	}
	for _, appID := range newRefIDs.RefAppIDs {
		if !gofn.Contain(r.RefAppIDs, appID) {
			res.RefAppIDs = append(res.RefAppIDs, appID)
		}
	}
	for _, projectID := range newRefIDs.RefProjectIDs {
		if !gofn.Contain(r.RefProjectIDs, projectID) {
			res.RefProjectIDs = append(res.RefProjectIDs, projectID)
		}
	}
	for _, projectEnvID := range newRefIDs.RefProjectEnvIDs {
		if !gofn.Contain(r.RefProjectEnvIDs, projectEnvID) {
			res.RefProjectEnvIDs = append(res.RefProjectEnvIDs, projectEnvID)
		}
	}
	for _, userID := range newRefIDs.RefUserIDs {
		if !gofn.Contain(r.RefUserIDs, userID) {
			res.RefUserIDs = append(res.RefUserIDs, userID)
		}
	}
	return res
}

func (r *RefObjectIDs) GetResourceLinks(srcType base.ResourceType, srcID string) []*ResLink {
	resLinks := make([]*ResLink, 0, len(r.RefSettingIDs)+len(r.RefAppIDs)+len(r.RefProjectIDs)+
		len(r.RefProjectEnvIDs)+len(r.RefUserIDs))
	timeNow := timeutil.NowUTC()
	for _, refSettingID := range r.RefSettingIDs {
		resLinks = append(resLinks, &ResLink{
			SrcType:   srcType,
			SrcID:     srcID,
			DstType:   base.ResourceTypeSetting,
			DstID:     refSettingID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}
	for _, refAppID := range r.RefAppIDs {
		resLinks = append(resLinks, &ResLink{
			SrcType:   srcType,
			SrcID:     srcID,
			DstType:   base.ResourceTypeApp,
			DstID:     refAppID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}
	for _, refProjectID := range r.RefProjectIDs {
		resLinks = append(resLinks, &ResLink{
			SrcType:   srcType,
			SrcID:     srcID,
			DstType:   base.ResourceTypeProject,
			DstID:     refProjectID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}
	for _, refProjectEnvID := range r.RefProjectEnvIDs {
		resLinks = append(resLinks, &ResLink{
			SrcType:   srcType,
			SrcID:     srcID,
			DstType:   base.ResourceTypeProjectEnv,
			DstID:     refProjectEnvID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}
	for _, refUserID := range r.RefUserIDs {
		resLinks = append(resLinks, &ResLink{
			SrcType:   srcType,
			SrcID:     srcID,
			DstType:   base.ResourceTypeUser,
			DstID:     refUserID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}
	return resLinks
}
