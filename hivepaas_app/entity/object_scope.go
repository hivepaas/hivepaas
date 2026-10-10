package entity

import (
	"fmt"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
)

type ObjectScope struct {
	ScopeType    base.ObjectScopeType
	AppID        string
	ParentAppID  string
	ProjectID    string
	ProjectEnvID string
	UserID       string

	App        *App
	ParentApp  *App
	ProjectEnv *ProjectEnv
	Project    *Project
	User       *User

	LockScopeObject  bool
	NotRequireActive bool
	NoInherited      bool
	// IncludeEnvApps makes a project env's scope reach the settings of the apps
	// in the env, next to the env's own, and nothing inherited or shared: the
	// env's scheduled jobs are listed with its apps'.
	IncludeEnvApps bool
	// ProjectEnvIDs, set on a project's scope, keeps it to these of its envs: a
	// project read through a grant on some of its envs shows those alone.
	ProjectEnvIDs []string
}

// FilteredBy is the scope a list reached through s shows once filtered to a
// project, an env or an app: the filter's when s holds it, s itself when the
// filter holds s, and nil when the two do not meet. A filter narrows what s
// shows and never widens it; a project kept to some of its envs holds those
// alone.
func (s *ObjectScope) FilteredBy(filter *ObjectScope) *ObjectScope {
	if filter == nil {
		return s
	}
	switch s.ScopeType {
	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		return filter
	case base.ObjectScopeProject:
		switch {
		case filter.ProjectID != s.ProjectID:
			return nil
		case filter.ScopeType == base.ObjectScopeProject:
			return s
		case len(s.ProjectEnvIDs) > 0 && !slices.Contains(s.ProjectEnvIDs, filter.ProjectEnvID):
			return nil
		}
		return filter
	case base.ObjectScopeProjectEnv:
		switch {
		case filter.ScopeType == base.ObjectScopeProject && filter.ProjectID == s.ProjectID:
			return s
		case filter.ProjectEnvID != s.ProjectEnvID:
			return nil
		case filter.ScopeType == base.ObjectScopeApp:
			return filter
		}
		return s
	case base.ObjectScopeApp:
		switch {
		case filter.ScopeType == base.ObjectScopeApp && filter.AppID == s.AppID,
			filter.ScopeType == base.ObjectScopeProjectEnv && filter.ProjectEnvID == s.ProjectEnvID,
			filter.ScopeType == base.ObjectScopeProject && filter.ProjectID == s.ProjectID:
			return s
		}
	case base.ObjectScopeUser:
	}
	return nil
}

func (s *ObjectScope) IsGlobalScope() bool {
	return s.ScopeType == base.ObjectScopeGlobal
}

func (s *ObjectScope) IsHivepaasScope() bool {
	return s.ScopeType == base.ObjectScopeHivepaas
}

func (s *ObjectScope) IsAppScope() bool {
	return s.ScopeType == base.ObjectScopeApp
}

func (s *ObjectScope) IsProjectEnvScope() bool {
	return s.ScopeType == base.ObjectScopeProjectEnv
}

func (s *ObjectScope) IsProjectScope() bool {
	return s.ScopeType == base.ObjectScopeProject
}

func (s *ObjectScope) IsUserScope() bool {
	return s.ScopeType == base.ObjectScopeUser
}

func (s *ObjectScope) ScopeObjectID() string {
	switch s.ScopeType {
	case base.ObjectScopeApp:
		return s.AppID
	case base.ObjectScopeProjectEnv:
		return s.ProjectEnvID
	case base.ObjectScopeProject:
		return s.ProjectID
	case base.ObjectScopeUser:
		return s.UserID
	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		return ""
	default:
		return ""
	}
}

func (s *ObjectScope) CalcProjectEnvKey() string {
	_, envKey := projecthelper.ParseProjectEnvID(s.ProjectEnvID)
	if envKey != "" {
		return envKey
	}
	return projecthelper.CalcProjectEnvKey(s.ProjectEnvID)
}

func (s *ObjectScope) IsValid() bool {
	if s.ScopeType == base.ObjectScopeGlobal {
		return s.ScopeObjectID() == ""
	}
	return s.ScopeObjectID() != ""
}

func (s *ObjectScope) IsObjectLoaded() bool {
	switch s.ScopeType {
	case base.ObjectScopeApp:
		return s.App != nil && s.ProjectEnv != nil && s.Project != nil &&
			(s.ParentAppID == "" || s.ParentApp != nil)
	case base.ObjectScopeProjectEnv:
		return s.ProjectEnv != nil && s.Project != nil
	case base.ObjectScopeProject:
		return s.Project != nil
	case base.ObjectScopeUser:
		return s.User != nil
	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		return true
	default:
		return false
	}
}

func (s *ObjectScope) GetApp() *App {
	if s.ScopeType == base.ObjectScopeApp {
		return s.App
	}
	return nil
}

func (s *ObjectScope) GetProject() *Project {
	if s.ScopeType == base.ObjectScopeProject && s.Project != nil {
		return s.Project
	}
	if s.ScopeType == base.ObjectScopeProjectEnv && s.ProjectEnv != nil && s.ProjectEnv.Project != nil {
		return s.ProjectEnv.Project
	}
	if s.ScopeType == base.ObjectScopeApp && s.App != nil && s.App.Project != nil {
		return s.App.Project
	}
	return nil
}

func (s *ObjectScope) GetProjectEnv() *ProjectEnv {
	if s.ScopeType == base.ObjectScopeProjectEnv && s.ProjectEnv != nil {
		return s.ProjectEnv
	}
	if s.ScopeType == base.ObjectScopeApp && s.App != nil && s.App.ProjectEnv != nil {
		return s.App.ProjectEnv
	}
	return nil
}

func (s *ObjectScope) GetBaseURLPath() string {
	switch s.ScopeType {
	case base.ObjectScopeApp:
		return fmt.Sprintf("projects/%v/%v/apps/%v", s.App.ProjectID, s.App.ProjectEnv.Name, s.App.ID)
	case base.ObjectScopeProjectEnv:
		return fmt.Sprintf("projects/%v/%v", s.ProjectEnv.ProjectID, s.ProjectEnv.Name)
	case base.ObjectScopeProject:
		return fmt.Sprintf("projects/%v", s.Project.ID)
	case base.ObjectScopeUser:
		return fmt.Sprintf("users/%v", s.User.ID)
	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		return ""
	default:
		return ""
	}
}

func NewObjectScopeGlobal() *ObjectScope {
	return &ObjectScope{ScopeType: base.ObjectScopeGlobal}
}

func NewObjectScopeHivepaas() *ObjectScope {
	return &ObjectScope{ScopeType: base.ObjectScopeHivepaas}
}

func NewObjectScopeApp(appID, parentAppID, projectID, env string) *ObjectScope {
	return &ObjectScope{
		ScopeType:    base.ObjectScopeApp,
		AppID:        appID,
		ParentAppID:  parentAppID,
		ProjectID:    projectID,
		ProjectEnvID: projecthelper.CalcProjectEnvID(projectID, env),
	}
}

func NewObjectScopeProjectEnv(projectID string, env string) *ObjectScope {
	return &ObjectScope{
		ScopeType:    base.ObjectScopeProjectEnv,
		ProjectID:    projectID,
		ProjectEnvID: projecthelper.CalcProjectEnvID(projectID, env),
	}
}

func NewObjectScopeProject(projectID string) *ObjectScope {
	return &ObjectScope{
		ScopeType: base.ObjectScopeProject,
		ProjectID: projectID,
	}
}

func NewObjectScopeUser(userID string) *ObjectScope {
	return &ObjectScope{
		ScopeType: base.ObjectScopeUser,
		UserID:    userID,
	}
}
