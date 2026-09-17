package apptemplateuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

func (uc *UC) ListAppTemplates(
	ctx context.Context,
	_ *basedto.Auth,
	req *apptemplatedto.ListAppTemplatesReq,
) (*apptemplatedto.ListAppTemplatesResp, error) {
	index, err := uc.appTemplateService.Index(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	page, pagingMeta := pageTemplates(filterTemplates(index.Index.Templates, req), req.Paging)
	return &apptemplatedto.ListAppTemplatesResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: apptemplatedto.TransformAppTemplateSummaries(page, base.CurrentVersion),
	}, nil
}

// GetAppTemplateCatalog returns what the store needs once when it opens: the source,
// its revision, and the categories and tags to filter the list by.
func (uc *UC) GetAppTemplateCatalog(
	ctx context.Context,
	_ *basedto.Auth,
	_ *apptemplatedto.GetAppTemplateCatalogReq,
) (*apptemplatedto.GetAppTemplateCatalogResp, error) {
	index, err := uc.appTemplateService.Index(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &apptemplatedto.GetAppTemplateCatalogResp{
		Data: apptemplatedto.TransformAppTemplateCatalog(index),
	}, nil
}

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

func (uc *UC) GetAppTemplateIcon(
	ctx context.Context,
	req *apptemplatedto.GetAppTemplateIconReq,
) (*apptemplatedto.GetAppTemplateIconResp, error) {
	icon, err := uc.appTemplateService.Icon(ctx, req.SHA256)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &apptemplatedto.GetAppTemplateIconResp{Content: icon.Content, ContentType: icon.ContentType}, nil
}
