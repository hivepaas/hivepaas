package apptemplateuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// PreflightAppFromTemplate answers what creating this request would run into,
// without creating anything.
//
// Today it answers one question: which of the apps would be given a directory
// that already holds something. It is a list rather than a refusal because a
// template creates up to eight apps at once, and the answer people want is one
// decision about all of them.
func (uc *UC) PreflightAppFromTemplate(
	ctx context.Context,
	_ *basedto.Auth,
	req *apptemplatedto.PreflightAppFromTemplateReq,
) (*apptemplatedto.PreflightAppFromTemplateResp, error) {
	rendered, err := uc.appTemplateService.Render(ctx, &apptemplateservice.RenderReq{
		Name:             req.Template,
		AppName:          req.Name,
		Version:          req.Version,
		Variant:          req.Variant,
		Params:           req.Params,
		DependencyParams: req.DependencyParams,
		ImageTag:         req.ImageTag,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	create := &req.CreateAppFromTemplateReq
	plan, err := uc.planStorage(ctx, uc.db, create, planApps(create, rendered))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &apptemplatedto.PreflightAppFromTemplateResp{Data: &apptemplatedto.PreflightAppRe{
		Storage:          storageResults(plan.Findings),
		StorageUnchecked: storageResults(plan.Unchecked),
	}}, nil
}

func storageResults(findings []*storageFinding) []*apptemplatedto.PreflightStorageRes {
	results := make([]*apptemplatedto.PreflightStorageRes, 0, len(findings))
	for _, finding := range findings {
		item := &apptemplatedto.PreflightStorageRes{
			App:        finding.AppName,
			AppKey:     finding.AppKey,
			IsDatabase: finding.IsDatabase,
			Path:       finding.Path,
		}
		item.Volume.ID, item.Volume.Name = finding.VolumeID, finding.VolumeName
		results = append(results, item)
	}
	return results
}
