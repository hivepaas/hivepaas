package backuprepodto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListBackupRepoReq struct {
	settings.ListSettingReq
}

func NewListBackupRepoReq() *ListBackupRepoReq {
	return &ListBackupRepoReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListBackupRepoReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.ListSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListBackupRepoResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*BackupRepoResp `json:"data"`
}

func TransformBackupRepos(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) (resp []*BackupRepoResp, err error) {
	resp = make([]*BackupRepoResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformBackupRepo(setting, refObjects)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
