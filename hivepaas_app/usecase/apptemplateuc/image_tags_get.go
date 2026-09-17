package apptemplateuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// GetAppTemplateImageTags reads the registry, so it is the one catalog call that
// reaches outside this installation. It runs when somebody asks for it and never
// on a schedule.
func (uc *UC) GetAppTemplateImageTags(
	ctx context.Context,
	_ *basedto.Auth,
	req *apptemplatedto.GetAppTemplateImageTagsReq,
) (*apptemplatedto.GetAppTemplateImageTagsResp, error) {
	tags, err := uc.appTemplateService.ImageTags(ctx, &apptemplateservice.ImageTagsReq{
		Name:    req.Name,
		Version: req.Version,
		Variant: req.Variant,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &apptemplatedto.GetAppTemplateImageTagsResp{
		Data: apptemplatedto.TransformAppTemplateImageTags(tags),
	}, nil
}
