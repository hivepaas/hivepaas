package syserrordto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type ListSysErrorReq struct {
	Status []int    `json:"-" mapstructure:"status"`
	Code   []string `json:"-" mapstructure:"code"`
	Search string   `json:"-" mapstructure:"search"`

	Paging basedto.Paging `json:"-"`
}

func NewListSysErrorReq() *ListSysErrorReq {
	return &ListSysErrorReq{
		Paging: basedto.Paging{
			// Default paging if unset by client
			Sort: basedto.Orders{{Direction: basedto.DirectionDesc, ColumnName: "created_at"}},
		},
	}
}

func (req *ListSysErrorReq) Validate() hperrors.ValidationErrors {
	return nil
}

type ListSysErrorResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*SysErrorResp   `json:"data"`
}

func TransformSysErrors(sysErrors []*entity.SysError) (resp []*SysErrorResp, err error) {
	resp = make([]*SysErrorResp, 0, len(sysErrors))
	for _, appError := range sysErrors {
		item, err := TransformSysError(appError)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
