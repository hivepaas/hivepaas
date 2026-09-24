package specuc

import (
	"context"
	"errors"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc/specdto"
)

// ApplyImport applies what a validate of the same request planned, in three
// phases: the database in one transaction, recorded as spec-import; then the
// running services of the apps it updated; then the deployments it queued.
//
// The transaction is all or nothing, and what provisioning created in docker for
// the new apps is removed when it does not commit. The second phase is not: a
// running service cannot be rolled back, so an app that fails there is reported
// as failed and the rest go on, and importing again retries it.
func (uc *UC) ApplyImport(
	ctx context.Context,
	auth *basedto.Auth,
	req *specdto.ApplyImportReq,
) (_ *specdto.ApplyImportResp, err error) {
	serviceReq := &specservice.ApplyImportReq{
		ValidateImportReq: *uc.importReq(auth, &req.ValidateImportReq),
		OperatorID:        auth.User.ID,
		PlanHash:          req.PlanHash,
		AcceptIssues:      req.AcceptIssues,
	}

	var applied *specservice.ApplyImportResp
	committed := false
	cleanup := func() error {
		if applied == nil || applied.Cleanup == nil {
			return nil
		}
		cleanupErr := applied.Cleanup(context.WithoutCancel(ctx))
		applied = nil
		return hperrors.Wrap(cleanupErr)
	}
	defer func() {
		if rec := recover(); rec != nil {
			err = errors.Join(err, hperrors.NewPanic(rec))
		}
		if err != nil && !committed {
			err = errors.Join(err, cleanup())
		}
	}()

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		// A transaction retried has to start from nothing in docker as well.
		if cleanupErr := cleanup(); cleanupErr != nil {
			return hperrors.Wrap(cleanupErr)
		}
		var txErr error
		applied, txErr = uc.applyInTx(ctx, db, auth, req, serviceReq)
		return txErr
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	committed = true
	return uc.afterImport(ctx, uc.db, applied), nil
}

// applyInTx is phase 1: the import written and recorded in the transaction.
// What it created in docker comes back with an error too, for the caller to
// remove.
func (uc *UC) applyInTx(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *specdto.ApplyImportReq,
	serviceReq *specservice.ApplyImportReq,
) (*specservice.ApplyImportResp, error) {
	applied, err := uc.specService.ApplyImport(ctx, db, serviceReq)
	if err != nil {
		return applied, hperrors.Wrap(err)
	}
	return applied, uc.recordSpecImport(ctx, db, auth, req, applied)
}

// afterImport is phases 2 and 3, once committed. What fails from here is
// reported rather than returned: the import itself stands, and an app that
// failed says so in its outcome.
func (uc *UC) afterImport(
	ctx context.Context,
	db database.IDB,
	applied *specservice.ApplyImportResp,
) *specdto.ApplyImportResp {
	var warnings []string
	if err := applied.AfterCommit(ctx, db); err != nil {
		warnings = append(warnings, "writing certificate files: "+err.Error())
	}
	// After phase 2, so that a deployment runs on the service it updated.
	if err := uc.taskQueue.ScheduleTask(ctx, applied.Tasks...); err != nil {
		warnings = append(warnings, "scheduling deployments: "+err.Error())
	}
	return &specdto.ApplyImportResp{
		Meta: &basedto.Meta{Warning: strings.Join(warnings, "\n")},
		Data: &specdto.ApplyImportData{Plan: applied.Plan, Deployments: applied.Deployments},
	}
}

// recordSpecImport records an import in its transaction: what was imported
// where, and what it did, counted. A failure to record aborts the import.
func (uc *UC) recordSpecImport(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *specdto.ApplyImportReq,
	applied *specservice.ApplyImportResp,
) error {
	plan := applied.Plan
	detail := auditdetail.New().
		Set("digest", plan.Bundle.Digest).
		Set("bundleScope", plan.Bundle.Scope).
		Set("secretsMode", string(plan.Bundle.SecretsMode)).
		Set("include", req.Selection.Include).
		Set("exclude", req.Selection.Exclude).
		Set("existing", string(req.Options.Existing)).
		Set("deployCreated", req.Options.DeployCreated).
		Set("deployChangedSource", req.Options.DeployChangedSource).
		Set("summary", plan.Summary)

	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:     base.AuditLogTypeSpecImport,
		Scope:    req.Scope.ScopeType,
		ObjectID: req.Scope.ScopeObjectID(),
		Source:   base.AuditLogSourceAPIAction,
		Auth:     auth,
		ResType:  resourceTypeOfScope(req.Scope.ScopeType),
		ResID:    req.Scope.ScopeObjectID(),
		Detail:   detail.String(),
	})
	return hperrors.Wrap(err)
}
