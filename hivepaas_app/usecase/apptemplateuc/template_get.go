package apptemplateuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

func (uc *UC) GetAppTemplate(
	ctx context.Context,
	_ *basedto.Auth,
	req *apptemplatedto.GetAppTemplateReq,
) (*apptemplatedto.GetAppTemplateResp, error) {
	tmpl, err := uc.appTemplateService.Template(ctx, req.Name)
	if errors.Is(err, hperrors.ErrAppTemplateIncompatible) {
		return uc.getIncompatibleAppTemplate(ctx, req.Name, err)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &apptemplatedto.GetAppTemplateResp{
		Data: apptemplatedto.TransformAppTemplate(tmpl, base.CurrentVersion),
	}, nil
}

// getIncompatibleAppTemplate answers for a template written for a newer HivePaaS,
// whose file this one does not read. The store still shows it, as the listing
// does, from what the index says of it: with no form, and marked as needing a
// newer HivePaaS. A dependency that needs one is refused as it was.
func (uc *UC) getIncompatibleAppTemplate(
	ctx context.Context,
	name string,
	refused error,
) (*apptemplatedto.GetAppTemplateResp, error) {
	index, err := uc.appTemplateService.Index(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	entry := index.Index.FindTemplate(name)
	if entry == nil || templatemodel.IsCompatible(entry.Requires, base.CurrentVersion) {
		return nil, hperrors.Wrap(refused)
	}
	return &apptemplatedto.GetAppTemplateResp{
		Data: apptemplatedto.TransformIncompatibleAppTemplate(index, entry, base.CurrentVersion),
	}, nil
}
