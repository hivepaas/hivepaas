package base

type ObjectScopeType string

const (
	ObjectScopeGlobal     ObjectScopeType = ""
	ObjectScopeHivepaas   ObjectScopeType = "hivepaas"
	ObjectScopeUser       ObjectScopeType = "user"
	ObjectScopeProjectEnv ObjectScopeType = "project-env"
	ObjectScopeProject    ObjectScopeType = "project"
	ObjectScopeApp        ObjectScopeType = "app"
)
