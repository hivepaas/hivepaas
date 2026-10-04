package specserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func (s *service) PlanBundle(
	ctx context.Context,
	db database.IDB,
	req *specservice.PlanBundleReq,
) (*specmodel.ImportPlan, error) {
	if req.Doc == nil {
		return nil, hperrors.Wrap(hperrors.ErrSpecBundleInvalid)
	}
	p, err := s.planBundleWith(ctx, db, &req.ValidateImportReq, copyBundle(req.Doc), req.Issues)
	if err != nil {
		return nil, err
	}
	return p.result, nil
}

func (s *service) ApplyBundle(
	ctx context.Context,
	db database.IDB,
	req *specservice.ApplyBundleReq,
) (*specservice.ApplyImportResp, error) {
	if req.Doc == nil {
		return nil, hperrors.Wrap(hperrors.ErrSpecBundleInvalid)
	}
	return s.applyBundleWith(ctx, db, &specservice.ApplyImportReq{
		ValidateImportReq: req.ValidateImportReq,
		OperatorID:        req.OperatorID,
		PlanHash:          req.PlanHash,
		AcceptIssues:      req.AcceptIssues,
	}, copyBundle(req.Doc), req.Issues)
}

// copyBundle is a bundle whose maps planning may restrict to a scope without
// touching the caller's.
func copyBundle(b *specmodel.ImportBundle) *specmodel.ImportBundle {
	out := *b
	out.Projects = make(map[string]*specmodel.ProjectDoc, len(b.Projects))
	for key, doc := range b.Projects {
		out.Projects[key] = doc
	}
	out.Envs = make(map[string]map[string]*specmodel.EnvDoc, len(b.Envs))
	for key, envs := range b.Envs {
		out.Envs[key] = envs
	}
	return &out
}
