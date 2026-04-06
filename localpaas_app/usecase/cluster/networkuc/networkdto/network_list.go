package networkdto

import (
	"github.com/docker/docker/api/types/network"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type ListNetworkReq struct {
	ProjectID string `json:"-"`
	ListAll   bool   `json:"-" mapstructure:"listAll"`
	Search    string `json:"-" mapstructure:"search"`

	Paging basedto.Paging `json:"-"`
}

func NewListNetworkReq() *ListNetworkReq {
	return &ListNetworkReq{
		Paging: basedto.Paging{
			Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
		},
	}
}

func (req *ListNetworkReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListNetworkResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*NetworkResp    `json:"data"`
}

func TransformNetworks(networks []network.Summary) []*NetworkResp {
	return gofn.MapSlice(networks, func(net network.Summary) *NetworkResp {
		return TransformNetwork(&net)
	})
}
