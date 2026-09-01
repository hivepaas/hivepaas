package entity

import (
	"fmt"

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
