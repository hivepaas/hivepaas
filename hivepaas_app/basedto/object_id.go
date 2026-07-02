package basedto

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// ObjectIDReq request input for an object id
type ObjectIDReq struct {
	ID string `json:"id"`
}

func (req *ObjectIDReq) ToIDString() string {
	if req == nil {
		return ""
	}
	return req.ID
}

func (req *ObjectIDReq) ToEntity() *entity.ObjectID {
	if req == nil {
		return nil
	}
	return &entity.ObjectID{ID: req.ID}
}

type ObjectIDSliceReq []*ObjectIDReq

func (req ObjectIDSliceReq) ToIDStringSlice() []string {
	result := make([]string, 0, len(req))
	for _, obj := range req {
		result = append(result, obj.ID)
	}
	return result
}

func (req ObjectIDSliceReq) ToEntity() entity.ObjectIDSlice {
	result := make(entity.ObjectIDSlice, 0, len(req))
	for _, obj := range req {
		result = append(result, obj.ToEntity())
	}
	return result
}

func (req ObjectIDSliceReq) HasID(id string) bool {
	for _, obj := range req {
		if obj.ID == id {
			return true
		}
	}
	return false
}

func (req *ObjectIDSliceReq) AppendID(id string) {
	*req = append(*req, &ObjectIDReq{ID: id})
}

// ObjectIDResp response for object with id
type ObjectIDResp struct {
	ID string `json:"id"`
}

type ObjectIDSliceResp []*ObjectIDResp

// NamedObjectResp response for object with name
type NamedObjectResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ObjectUpdatedAtResp struct {
	UpdatedAt time.Time `json:"updatedAt"`
}

type NamedObjectSliceResp []*NamedObjectResp
