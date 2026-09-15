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

// CapabilityCheck asks whether a subject holds one capability.
//
// Unlike the other checks it takes no action: a capability is a single operation,
// so there is exactly one action it could be checked for, and the check fills it
// in itself. Leave BaseAccessCheck's Action, AllOf and AnyOf unset - anything put
// there is overwritten.
type CapabilityCheck struct {
	BaseAccessCheck

	Capability base.ResourceCapability
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

// RevealSubject identifies the secret being asked for.
//
// It exists because a stored setting is not the only secret the API hands out in
// the clear - the Swarm join token is another, and it is not a setting at all -
// and every one of them has to pass the same two gates and leave the same record.
type RevealSubject struct {
	Scope    base.ObjectScopeType
	ObjectID string
	Source   base.AuditLogSource

	// SecretType is the kind of secret being asked for, and decides whether the
	// operator's ReturnSecretsViaAPI flag can be stood down for it. Empty - which
	// is every stored setting - is always bound by the flag.
	SecretType base.SecretType

	ResType base.ResourceType
	ResID   string
	// ResName is stored as it reads now, because the object may be renamed or
	// deleted long before anyone comes to read the entry.
	ResName string

	// Detail is free-form JSON, and must never carry the secret it describes.
	Detail string
}
