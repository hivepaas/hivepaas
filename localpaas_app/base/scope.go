package base

type ObjectScopeType string

const (
	ObjectScopeGlobal  ObjectScopeType = ""
	ObjectScopeUser    ObjectScopeType = "user"
	ObjectScopeProject ObjectScopeType = "project"
	ObjectScopeApp     ObjectScopeType = "app"
)

type ObjectScope struct {
	AppID            string
	ParentAppID      string
	ProjectID        string
	UserID           string
	NotRequireActive bool
}

func (s *ObjectScope) ScopeType() ObjectScopeType {
	switch {
	case s.AppID != "":
		return ObjectScopeApp
	case s.ProjectID != "":
		return ObjectScopeProject
	case s.UserID != "":
		return ObjectScopeUser
	default:
		return ObjectScopeGlobal
	}
}

func (s *ObjectScope) IsGlobalScope() bool {
	return s.ScopeType() == ObjectScopeGlobal
}

func (s *ObjectScope) IsAppScope() bool {
	return s.AppID != ""
}

func (s *ObjectScope) IsProjectScope() bool {
	return s.ProjectID != ""
}

func (s *ObjectScope) IsUserScope() bool {
	return s.UserID != ""
}

func (s *ObjectScope) MainObjectID() string {
	switch {
	case s.AppID != "":
		return s.AppID
	case s.ProjectID != "":
		return s.ProjectID
	case s.UserID != "":
		return s.UserID
	default:
		return ""
	}
}

func NewObjectScopeGlobal() *ObjectScope {
	return &ObjectScope{}
}

func NewObjectScopeApp(appID, projectID string) *ObjectScope {
	return &ObjectScope{
		AppID:     appID,
		ProjectID: projectID,
	}
}

func NewObjectScopeProject(projectID string) *ObjectScope {
	return &ObjectScope{
		ProjectID: projectID,
	}
}

func NewObjectScopeUser(userID string) *ObjectScope {
	return &ObjectScope{
		UserID: userID,
	}
}
