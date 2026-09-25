package settingmountdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListSettingMountReq struct {
	settings.ListSettingReq
}

func NewListSettingMountReq() *ListSettingMountReq {
	return &ListSettingMountReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListSettingMountReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.ListSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListSettingMountResp struct {
	Meta *basedto.ListMeta   `json:"meta"`
	Data []*SettingMountResp `json:"data"`
}

// TransformSettingMounts renders entries with their states, by setting id.
func TransformSettingMounts(
	entries []*entity.Setting,
	refObjects *entity.RefObjects,
	states map[string]*settingmountservice.EntryState,
) ([]*SettingMountResp, error) {
	out := make([]*SettingMountResp, 0, len(entries))
	for _, setting := range entries {
		item, err := TransformSettingMount(setting, refObjects, states[setting.ID])
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		out = append(out, item)
	}
	return out, nil
}
