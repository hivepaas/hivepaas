package accessiblebyprojectsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/accessiblebyprojectsuc/accessiblebyprojectsdto"
)

func (uc *UC) GetAccessibleByProjects(
	ctx context.Context,
	auth *basedto.Auth,
	req *accessiblebyprojectsdto.GetAccessibleByProjectsReq,
) (*accessiblebyprojectsdto.GetAccessibleByProjectsResp, error) {
	setting, err := uc.SettingRepo.GetByID(ctx, uc.DB, nil, "", req.SettingID, false,
		bunex.SelectRelation("AccessibleByProjects"),
		bunex.SelectRelation("AccessibleByProjects.Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &accessiblebyprojectsdto.GetAccessibleByProjectsResp{
		Data: accessiblebyprojectsdto.TransformAccessibleByProjects(setting),
	}, nil
}
