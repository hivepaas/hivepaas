package permission

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

type AccessCheck struct {
	SubjectType        base.SubjectType
	SubjectID          string
	ResourceModule     base.ResourceModule
	ResourceType       base.ResourceType
	ResourceID         string
	ParentResourceType base.ResourceType
	ParentResourceID   string

	// The below are mutual exclusive
	Action base.ActionType
	AllOf  []base.ActionType
	AnyOf  []base.ActionType
}

func (ac *AccessCheck) IsValid() bool {
	return gofn.If(ac.Action != "", 1, 0)+
		gofn.If(len(ac.AllOf) > 0, 1, 0)+
		gofn.If(len(ac.AnyOf) > 0, 1, 0) == 1
}
