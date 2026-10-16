package licenseuc

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/licenseuc/licensedto"
)

// RefreshLicense asks the portal now for a newer key than the one stored, and
// answers with what it said and the license in force after it, as GET answers
// it. The portal is asked whether or not the daily check is switched off: a
// person asking is the one deciding to.
//
// It is asked outside any transaction, as that may take a while. A check that
// failed is recorded, and then answered as the error it failed with: 503 for a
// portal that could not be reached.
func (uc *UC) RefreshLicense(
	ctx context.Context,
	auth *basedto.Auth,
	_ *licensedto.RefreshLicenseReq,
) (*licensedto.RefreshLicenseResp, error) {
	fetched, err := uc.licenseService.FetchRefresh(ctx, &licenseservice.FetchRefreshReq{Manual: true})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	var applied *licenseservice.ApplyRefreshResp
	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		resp, err := uc.licenseService.ApplyRefresh(ctx, db, fetched)
		if err != nil {
			return hperrors.Wrap(err)
		}
		// In the transaction, so a key whose install cannot be recorded is not
		// installed either.
		if resp.Changed {
			if err = uc.recordRefresh(ctx, db, auth, resp); err != nil {
				return hperrors.Wrap(err)
			}
		}
		applied = resp
		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if applied.Result == base.LicenseRefreshFailed {
		return nil, hperrors.Wrap(applied.Err)
	}

	state, err := uc.licenseService.Current(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data, err := uc.licenseResp(ctx, state)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &licensedto.RefreshLicenseResp{
		Meta: &basedto.Meta{},
		Data: &licensedto.RefreshLicenseDataResp{Result: applied.Result, License: data},
	}, nil
}

// recordRefresh records a refresh that changed the license - a newer key
// installed, or a move learned of - as a write of the license setting, made by
// asking rather than by writing it.
//
// As for an install, the detail names the key by its license ID, plan, trial
// and expiry, and never holds the key or the customer.
func (uc *UC) recordRefresh(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	applied *licenseservice.ApplyRefreshResp,
) error {
	detail := auditdetail.New().Set("result", applied.Result)
	if state := applied.State; state != nil {
		if key := state.Key; key != nil {
			detail.Set("licenseId", key.LicenseID).
				Set("plan", key.Plan).
				Set("trial", key.Trial).
				Set("expiresAt", key.ExpiresAt.UTC().Format(time.RFC3339))
		}
		if !state.MovedAt.IsZero() {
			detail.Set("movedAt", state.MovedAt.UTC().Format(time.RFC3339))
		}
	}

	entry := &auditservice.Entry{
		Type:    base.AuditLogTypeSettingUpdate,
		Scope:   base.ObjectScopeGlobal,
		Source:  base.AuditLogSourceAPIAction,
		Section: "refresh",
		Auth:    auth,
		ResType: base.ResourceType(base.SettingTypeLicense),
		Detail:  detail.String(),
	}
	if applied.Setting != nil {
		entry.ResID, entry.ResName = applied.Setting.ID, applied.Setting.Name
	}
	if err := auditservice.RecordAllowed(ctx, uc.auditService, db, entry); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
