package specserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func (s *service) ApplyImport(
	ctx context.Context,
	db database.IDB,
	req *specservice.ApplyImportReq,
) (*specservice.ApplyImportResp, error) {
	bundle, err := readBundle(req.Bundle, req.Passphrase)
	if err != nil {
		return nil, err
	}
	if mode := bundle.Manifest.SecretsMode; mode.RevealsSecrets() && req.AuthorizeSecrets != nil {
		if err = req.AuthorizeSecrets(ctx, mode); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	return s.applyBundle(ctx, db, req, bundle)
}

// applyBundle applies a bundle already read. It is the part tests reach
// directly, so they can change a document before applying it.
func (s *service) applyBundle(
	ctx context.Context,
	db database.IDB,
	req *specservice.ApplyImportReq,
	bundle *specmodel.ImportBundle,
) (*specservice.ApplyImportResp, error) {
	// Planned again, on the transaction the writes happen in: the plan applied
	// is the plan of this installation as it is now, not as it was at validate.
	p, err := s.planBundle(ctx, db, &req.ValidateImportReq, bundle)
	if err != nil {
		return nil, err
	}
	if err = refuseToApply(p.result, req); err != nil {
		return nil, err
	}
	w := newWriter(p, req.OperatorID)
	if err = w.write(ctx); err != nil {
		return nil, err
	}
	return &specservice.ApplyImportResp{Plan: p.result, AfterCommit: w.afterCommit}, nil
}

// refuseToApply refuses a plan other than the one the operator saw, a plan
// with something blocked, and a plan with issues nobody accepted.
func refuseToApply(plan *specmodel.ImportPlan, req *specservice.ApplyImportReq) error {
	if plan.PlanHash != req.PlanHash {
		return hperrors.Wrap(hperrors.ErrSpecImportPlanChanged)
	}
	hasIssues := false
	for _, node := range plan.Nodes {
		if !node.Selected {
			continue
		}
		for _, issue := range node.Issues {
			if issue.Severity == specmodel.SeverityBlocked {
				return hperrors.Wrap(hperrors.ErrSpecImportBlocked).WithExtraDetail("%s: %s", node.Path, issue.Code)
			}
			hasIssues = true
		}
	}
	if hasIssues && !req.AcceptIssues {
		return hperrors.Wrap(hperrors.ErrSpecImportIssuesNotAccepted)
	}
	return nil
}
