package apptemplateuc

import (
	"context"
	"errors"

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
	auth *basedto.Auth,
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
	apps := planApps(create, rendered)

	plan, err := uc.planStorage(ctx, uc.db, create, apps)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	issues, err := uc.collectIssues(ctx, auth, req, rendered, apps)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &apptemplatedto.PreflightAppFromTemplateResp{Data: &apptemplatedto.PreflightAppResult{
		Storage:          storageResults(plan.Findings),
		StorageUnchecked: storageResults(plan.Unchecked),
		Issues:           issues,
	}}, nil
}

// collectIssues runs the refusals the creation runs and reports them instead of
// raising them.
//
// All of them run, not the first that fails: the creation stops at the first
// because it has nothing to do with the rest, but a dialog that says "and also"
// three times is three round trips through the same form. They share no state,
// so running them all is only running them all.
func (uc *UC) collectIssues(
	ctx context.Context,
	auth *basedto.Auth,
	req *apptemplatedto.PreflightAppFromTemplateReq,
	rendered *apptemplateservice.RenderResp,
	apps []*appToProvision,
) ([]*apptemplatedto.PreflightIssueResult, error) {
	create := &req.CreateAppFromTemplateReq
	checks := []func() error{
		func() error { return uc.checkAppRefs(ctx, create, rendered) },
		func() error { return uc.checkCapabilities(ctx, auth, apps) },
		func() error { return uc.checkDockerAPI(ctx, auth, apps) },
		func() error { return uc.checkSharedMounts(ctx, auth, create, apps) },
		func() error { return uc.checkPublishedPorts(ctx, apps) },
		func() error { return uc.checkDomains(ctx, create, apps) },
	}

	issues := make([]*apptemplatedto.PreflightIssueResult, 0, len(checks))
	for _, check := range checks {
		err := check()
		if err == nil {
			continue
		}
		var hpErr hperrors.HPError
		if !errors.As(err, &hpErr) {
			// Not a refusal but a failure - a database that could not be read.
			// Reporting it as the caller's problem would be a lie, so the whole
			// check fails and the screen goes on to create, which raises it
			// properly.
			return nil, hperrors.Wrap(err)
		}
		info := hpErr.Build(req.Lang)
		issues = append(issues, &apptemplatedto.PreflightIssueResult{Code: info.Code, Detail: info.Detail})
	}
	return issues, nil
}

func storageResults(findings []*storageFinding) []*apptemplatedto.PreflightStorageResult {
	results := make([]*apptemplatedto.PreflightStorageResult, 0, len(findings))
	for _, finding := range findings {
		item := &apptemplatedto.PreflightStorageResult{
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
