package apptemplateuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

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
