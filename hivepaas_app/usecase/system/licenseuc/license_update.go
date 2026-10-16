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

// UpdateLicense installs a key in place of the one stored, and answers with the
// license it puts in force. The service refuses a key that could never be in
// force here - invalid, for another installation, past its grace - and the stored
// key is then kept.
func (uc *UC) UpdateLicense(
	ctx context.Context,
	auth *basedto.Auth,
	req *licensedto.UpdateLicenseReq,
) (*licensedto.UpdateLicenseResp, error) {
	var installed *licenseservice.InstallResp
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		resp, err := uc.licenseService.Install(ctx, db, &licenseservice.InstallReq{Key: req.Key})
		if err != nil {
			return hperrors.Wrap(err)
		}
		// In the transaction, so a key whose install cannot be recorded is not
		// installed either.
		if err = uc.recordInstall(ctx, db, auth, resp); err != nil {
			return hperrors.Wrap(err)
		}
		installed = resp
		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	data, err := uc.licenseResp(ctx, installed.State)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &licensedto.UpdateLicenseResp{Meta: &basedto.Meta{}, Data: data}, nil
}

// recordInstall records the key installed as a write of the license setting, as
// any other setting's write is recorded.
//
// The detail names the key by its license ID, plan, trial and expiry. It never
// holds the key, nor the customer: a key carries the customer's email, and an
// audit row outlives the key it describes.
func (uc *UC) recordInstall(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	installed *licenseservice.InstallResp,
) error {
	logType := base.AuditLogTypeSettingUpdate
	if installed.Created {
		logType = base.AuditLogTypeSettingCreate
	}
	detail := auditdetail.New()
	if key := installed.State.Key; key != nil {
		detail.Set("licenseId", key.LicenseID).
			Set("plan", key.Plan).
			Set("trial", key.Trial).
			Set("expiresAt", key.ExpiresAt.UTC().Format(time.RFC3339))
	}

	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:    logType,
		Scope:   base.ObjectScopeGlobal,
		Source:  base.AuditLogSourceAPIUpdate,
		Auth:    auth,
		ResType: base.ResourceType(base.SettingTypeLicense),
		ResID:   installed.Setting.ID,
		ResName: installed.Setting.Name,
		Detail:  detail.String(),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
