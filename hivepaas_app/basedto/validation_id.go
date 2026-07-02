package basedto

import (
	"math"

	vld "github.com/tiendc/go-validator"
)

func ValidateID(id *string, required bool, field string) (res []vld.Validator) {
	if required {
		res = append(res, vld.Required(id).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_OBJECT_ID_REQUIRED"), // use default error key
		))
	}
	return res
}

func ValidateIDSlice(ids []string, unique bool, minLen int, field string) (result []vld.Validator) {
	if unique {
		result = append(result, vld.SliceUnique(ids).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_OBJECT_IDS_NON_UNIQUE"), // use default error key
		))
	}
	if minLen > 0 {
		result = append(result, vld.SliceLen(ids, minLen, math.MaxInt).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_OBJECT_IDS_REQUIRED"), // use default error key
		))
	}
	return result
}

func ValidateObjectIDReq(objID *ObjectIDReq, required bool, field string) (res []vld.Validator) {
	var id *string
	if objID != nil {
		id = &objID.ID
	}
	if required {
		res = append(res, vld.Required(id).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_OBJECT_ID_REQUIRED"), // use default error key
		))
	}
	return res
}

func ValidateObjectIDSliceReq(ids []*ObjectIDReq, unique bool, minLen int, field string) (result []vld.Validator) {
	if unique {
		result = append(result, vld.SliceUniqueBy(ids, func(item *ObjectIDReq) string {
			return item.ID
		}).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_OBJECT_IDS_NON_UNIQUE"), // use default error key
		))
	}
	if minLen > 0 {
		result = append(result, vld.SliceLen(ids, minLen, math.MaxInt).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_OBJECT_IDS_REQUIRED"), // use default error key
		))
	}
	return result
}
