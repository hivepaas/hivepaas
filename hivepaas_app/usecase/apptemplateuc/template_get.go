package apptemplateuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

func (uc *UC) GetAppTemplate(
	ctx context.Context,
	_ *basedto.Auth,
	req *apptemplatedto.GetAppTemplateReq,
) (*apptemplatedto.GetAppTemplateResp, error) {
	tmpl, err := uc.appTemplateService.Template(ctx, req.Name)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &apptemplatedto.GetAppTemplateResp{
		Data: apptemplatedto.TransformAppTemplate(tmpl, base.CurrentVersion),
	}, nil
}
