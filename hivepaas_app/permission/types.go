package permission

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type AccessCheck interface {
	IsValid() bool
	GetBase() *BaseAccessCheck
	InitSubject(auth *basedto.Auth)
}

type BaseAccessCheck struct {
	AccessCheck
	SubjectType base.SubjectType
	SubjectID   string

	// The below are mutual exclusive
	Action base.ActionType
	AllOf  []base.ActionType
	AnyOf  []base.ActionType
}

func (ac *BaseAccessCheck) IsValid() bool {
	return gofn.If(ac.Action != "", 1, 0)+
		gofn.If(len(ac.AllOf) > 0, 1, 0)+
		gofn.If(len(ac.AnyOf) > 0, 1, 0) == 1
}

func (ac *BaseAccessCheck) GetBase() *BaseAccessCheck {
	return ac
}

func (ac *BaseAccessCheck) InitSubject(auth *basedto.Auth) {
	if ac.SubjectID == "" {
		ac.SubjectType = base.SubjectTypeUser
		ac.SubjectID = auth.User.ID
	}
}

type ModuleAccessCheck struct {
	BaseAccessCheck

	Module base.ResourceModule
}

type GeneralResourceAccessCheck struct {
	BaseAccessCheck

	Module       base.ResourceModule
	ResourceType base.ResourceType
	ResourceID   string
}

type ProjectAccessCheck struct {
	BaseAccessCheck

	ProjectID  string
	ProjectEnv *string
}

type AppAccessCheck struct {
	BaseAccessCheck

	ProjectID  string
	ProjectEnv string
	ParentID   string
	AppID      string
}
