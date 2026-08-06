package entity

import (
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

	NotRequireActive bool
	NoInherited      bool
}

func (s *ObjectScope) IsGlobalScope() bool {
	return s.ScopeType == base.ObjectScopeGlobal
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
	case base.ObjectScopeGlobal:
		return ""
	case base.ObjectScopeApp:
		return s.AppID
	case base.ObjectScopeProjectEnv:
		return s.ProjectEnvID
	case base.ObjectScopeProject:
		return s.ProjectID
	case base.ObjectScopeUser:
		return s.UserID
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
	case base.ObjectScopeGlobal:
		return true
	case base.ObjectScopeApp:
		return s.App != nil && s.ProjectEnv != nil && s.Project != nil &&
			(s.ParentAppID == "" || s.ParentApp != nil)
	case base.ObjectScopeProjectEnv:
		return s.ProjectEnv != nil && s.Project != nil
	case base.ObjectScopeProject:
		return s.Project != nil
	case base.ObjectScopeUser:
		return s.User != nil
	default:
		return false
	}
}

func NewObjectScopeGlobal() *ObjectScope {
	return &ObjectScope{ScopeType: base.ObjectScopeGlobal}
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
