package apptemplateuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

func (uc *UC) GetAppTemplateIcon(
	ctx context.Context,
	req *apptemplatedto.GetAppTemplateIconReq,
) (*apptemplatedto.GetAppTemplateIconResp, error) {
	icon, err := uc.appTemplateService.Icon(ctx, &apptemplateservice.IconReq{
		Name:         req.Name,
		SHA256Prefix: req.SHA256Prefix,
		Ext:          req.Ext,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &apptemplatedto.GetAppTemplateIconResp{Content: icon.Content, ContentType: icon.ContentType}, nil
}
