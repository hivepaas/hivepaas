package networkdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetNetworkReq struct {
	settings.GetSettingReq
}

func NewGetNetworkReq() *GetNetworkReq {
	return &GetNetworkReq{}
}

func (req *GetNetworkReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetNetworkResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *NetworkResp  `json:"data"`
}

type NetworkResp struct {
	*settings.BaseSettingResp

	Driver     string            `json:"driver"`
	Internal   bool              `json:"internal"`
	Attachable bool              `json:"attachable"`
	Ingress    bool              `json:"ingress"`
	EnableIPv4 bool              `json:"enableIPv4"`
	EnableIPv6 bool              `json:"enableIPv6"`
	Options    map[string]string `json:"options"`
	Labels     map[string]string `json:"labels"`
}

func TransformNetwork(
	setting *entity.Setting,
	_ *entity.RefObjects,
	refClusterObjects *entity.RefClusterObjects,
) (resp *NetworkResp, err error) {
	netEnt := setting.MustAsClusterNetwork()
	if err = copier.Copy(&resp, netEnt); err != nil {
		return nil, apperrors.New(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	net := refClusterObjects.RefNetworks[netEnt.NetworkID]

	resp.Driver = net.Driver
	resp.Internal = net.Internal
	resp.Attachable = net.Attachable
	resp.Ingress = net.Ingress
	resp.EnableIPv4 = net.EnableIPv4
	resp.EnableIPv6 = net.EnableIPv6
	resp.Options = net.Options
	resp.Labels = net.Labels
	resp.CreatedAt = net.Created

	return resp, nil
}
