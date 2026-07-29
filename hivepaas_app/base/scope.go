package base

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
)

type ObjectScopeType string

const (
	ObjectScopeGlobal     ObjectScopeType = ""
	ObjectScopeUser       ObjectScopeType = "user"
	ObjectScopeProjectEnv ObjectScopeType = "project-env"
	ObjectScopeProject    ObjectScopeType = "project"
	ObjectScopeApp        ObjectScopeType = "app"
)

type ObjectScope struct {
	ScopeType    ObjectScopeType
	AppID        string
	ParentAppID  string
	ProjectID    string
	ProjectEnvID string
	UserID       string

	NotRequireActive bool
	NoInherited      bool
}

func (s *ObjectScope) IsGlobalScope() bool {
	return s.ScopeType == ObjectScopeGlobal
}

func (s *ObjectScope) IsAppScope() bool {
	return s.ScopeType == ObjectScopeApp
}

func (s *ObjectScope) IsProjectEnvScope() bool {
	return s.ScopeType == ObjectScopeProjectEnv
}

func (s *ObjectScope) IsProjectScope() bool {
	return s.ScopeType == ObjectScopeProject
}

func (s *ObjectScope) IsUserScope() bool {
	return s.ScopeType == ObjectScopeUser
}

func (s *ObjectScope) ScopeObjectID() string {
	switch s.ScopeType {
	case ObjectScopeGlobal:
		return ""
	case ObjectScopeApp:
		return s.AppID
	case ObjectScopeProjectEnv:
		return s.ProjectEnvID
	case ObjectScopeProject:
		return s.ProjectID
	case ObjectScopeUser:
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
	return s.ScopeObjectID() != ""
}

func NewObjectScopeGlobal() *ObjectScope {
	return &ObjectScope{ScopeType: ObjectScopeGlobal}
}

func NewObjectScopeApp(appID, parentAppID, projectID, env string) *ObjectScope {
	return &ObjectScope{
		ScopeType:    ObjectScopeApp,
		AppID:        appID,
		ParentAppID:  parentAppID,
		ProjectID:    projectID,
		ProjectEnvID: projecthelper.CalcProjectEnvID(projectID, env),
	}
}

func NewObjectScopeProjectEnv(projectID string, env string) *ObjectScope {
	return &ObjectScope{
		ScopeType:    ObjectScopeProjectEnv,
		ProjectID:    projectID,
		ProjectEnvID: projecthelper.CalcProjectEnvID(projectID, env),
	}
}

func NewObjectScopeProject(projectID string) *ObjectScope {
	return &ObjectScope{
		ScopeType: ObjectScopeProject,
		ProjectID: projectID,
	}
}

func NewObjectScopeUser(userID string) *ObjectScope {
	return &ObjectScope{
		ScopeType: ObjectScopeUser,
		UserID:    userID,
	}
}
