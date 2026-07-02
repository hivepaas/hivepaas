package basedto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// ObjectAccessReq request input for requesting access to an object
type ObjectAccessReq struct {
	ObjectIDReq
	Access base.AccessActions `json:"access"`
}

type ObjectAccessSliceReq []*ObjectAccessReq

func (req ObjectAccessSliceReq) ToIDStringSlice() []string {
	result := make([]string, 0, len(req))
	for _, obj := range req {
		result = append(result, obj.ID)
	}
	return result
}

// ModuleAccessReq request input for requesting access to a module
type ModuleAccessReq struct {
	ModuleIDReq
	Access base.AccessActions `json:"access"`
}

type ModuleIDReq struct {
	ID string `json:"id"` // could be module name
}

type ModuleAccessSliceReq []*ModuleAccessReq

type ObjectAccessResp struct {
	NamedObjectResp
	Access base.AccessActions `json:"access"`
}

type ObjectAccessSliceResp []*ObjectAccessResp
