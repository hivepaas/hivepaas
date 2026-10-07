package hpappuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc/hpappdto"
)

// GetHpProject is the project HivePaaS runs in, for the way to its apps - and
// their logs - the projects list does not show.
func (uc *UC) GetHpProject(
	ctx context.Context,
	_ *basedto.Auth,
	_ *hpappdto.GetHpProjectReq,
) (*hpappdto.GetHpProjectResp, error) {
	project, err := uc.projectRepo.GetByKey(ctx, uc.db, base.HivepaasProjectKey,
		bunex.SelectColumns("id", "name", "key"))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &hpappdto.GetHpProjectResp{
		Data: &hpappdto.HpProjectResp{ID: project.ID, Name: project.Name, Key: project.Key},
	}, nil
}
