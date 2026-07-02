package basedto

import (
	"fmt"
	"math"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func ValidateObjectAccessReq(access *ObjectAccessReq, required bool, field string) (res []vld.Validator) {
	var id *string
	if access != nil {
		id = &access.ID
	}
	if required {
		res = append(res, vld.Required(id).OnError(
			vld.SetField(field+".id", nil),
			vld.SetCustomKey("ERR_VLD_OBJECT_ID_REQUIRED"), // use default error key
		))
	}
	return res
}

func ValidateObjectAccessSliceReq(access ObjectAccessSliceReq, unique bool, minLen int, field string) (
	res []vld.Validator) {
	if unique {
		res = append(res, vld.SliceUniqueBy(access, func(item *ObjectAccessReq) string {
			return item.ID
		}).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_OBJECT_IDS_NON_UNIQUE"),
		))
	}
	if minLen > 0 {
		res = append(res, vld.SliceLen(access, minLen, math.MaxInt).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_OBJECT_IDS_REQUIRED"),
		))
	}
	return res
}

func ValidateModuleAccessReq(access *ModuleAccessReq, required bool, allowedValues []base.ResourceModule,
	field string) (res []vld.Validator) {
	var id *string
	if access != nil {
		id = &access.ID
	}
	if required {
		res = append(res, vld.Required(id).OnError(
			vld.SetField(field+".id", nil),
			vld.SetCustomKey("ERR_VLD_OBJECT_ID_REQUIRED"), // use default error key
		))
	}
	if id != nil && len(allowedValues) > 0 {
		allowedVals := gofn.ToStringSlice[string](allowedValues)
		res = append(res, vld.StrIn(id, allowedVals...).OnError(
			vld.SetField(field+".id", nil),
			vld.SetCustomKey("ERR_VLD_VALUE_NOT_IN_LIST"),
		))
	}
	return res
}

func ValidateModuleAccessSliceReq(access ModuleAccessSliceReq, unique bool, minLen int,
	allowedValues []base.ResourceModule, field string) (res []vld.Validator) {
	if unique {
		res = append(res, vld.SliceUniqueBy(access, func(item *ModuleAccessReq) string {
			return item.ID
		}).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_OBJECT_IDS_NON_UNIQUE"),
		))
	}
	if minLen > 0 {
		res = append(res, vld.SliceLen(access, minLen, math.MaxInt).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_OBJECT_IDS_REQUIRED"),
		))
	}
	if len(allowedValues) > 0 {
		allowedVals := gofn.ToStringSlice[string](allowedValues)
		res = append(res,
			vld.Slice(access).ForEach(func(elem *ModuleAccessReq, index int, elemValidator vld.ItemValidator) {
				elemValidator.Validate(
					vld.StrIn(&elem.ID, allowedVals...).OnError(
						vld.SetField(fmt.Sprintf("%s[%d]", field, index), nil),
						vld.SetCustomKey("ERR_VLD_VALUE_NOT_IN_LIST"),
					),
				)
			}),
		)
	}

	return res
}
