package licenseserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

const (
	licenseSettingName = "License"
)

// Install refuses only what could never be in force here: a key that does not
// verify, one for another installation, and one past its grace. Any other key
// replaces the one stored, a lower plan or a trial included: the person pasting
// it is the one who knows which key they mean.
func (s *service) Install(ctx context.Context, db database.Tx,
	req *licenseservice.InstallReq) (*licenseservice.InstallResp, error) {
	keyText := license.NormalizeKeyText(req.Key)
	// Its refusal is ErrLicenseInvalid, whose detail says what failed and never
	// what the key holds.
	verified, err := s.verify(keyText, s.allowDev())
	if err != nil {
		return nil, err
	}
	installationID, err := s.InstallationID(ctx)
	if err != nil {
		return nil, err
	}
	now, err := s.now(ctx)
	if err != nil {
		return nil, err
	}
	state := license.Evaluate(&license.EvalInput{
		KeyText:        keyText,
		Verified:       verified,
		InstallationID: installationID,
		Now:            now,
	})
	if err = refuseInstalling(state); err != nil {
		return nil, err
	}

	setting, created, err := s.saveKey(ctx, db, keyText)
	if err != nil {
		return nil, err
	}
	// The next Current verifies what is stored now. The cache would see the key
	// has changed anyway; dropping it also leaves nothing of the old key behind.
	s.dropVerification()
	return &licenseservice.InstallResp{Setting: setting, Created: created, State: state}, nil
}

// refuseInstalling is the refusal of a key that would not be in force here.
func refuseInstalling(state *license.State) error {
	switch state.Status {
	case base.LicenseStatusValid, base.LicenseStatusTrial, base.LicenseStatusGrace:
		return nil
	case base.LicenseStatusOtherInstallation:
		return hperrors.Wrap(hperrors.ErrLicenseOtherInstallation).WithParam("InstallationID", state.InstallationID)
	case base.LicenseStatusExpired:
		return hperrors.Wrap(hperrors.ErrLicenseExpired)
	case base.LicenseStatusNone, base.LicenseStatusInvalid, base.LicenseStatusMoved:
		// A key that did not verify was refused before it was evaluated, and one
		// being installed has not been moved.
	}
	return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("a license key being installed evaluated as %q",
		state.Status)
}

// saveKey writes the key as the global license setting, the one there is, and
// reports whether it was written for the first time. A key installed was not
// moved, whatever the one it replaces was.
func (s *service) saveKey(ctx context.Context, db database.Tx, keyText string) (*entity.Setting, bool, error) {
	setting, err := s.loadSetting(ctx, db, bunex.SelectFor("UPDATE OF setting"))
	if err != nil {
		return nil, false, err
	}
	timeNow := s.timeNow().UTC()
	created := setting == nil
	if created {
		setting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.ObjectScopeGlobal,
			Type:      base.SettingTypeLicense,
			CreatedAt: timeNow,
		}
	} else {
		setting.UpdateVer++
	}
	setting.Status = base.SettingStatusActive
	setting.Name = licenseSettingName
	setting.Version = entity.CurrentLicenseSettingsVersion
	setting.UpdatedAt = timeNow
	if err = setting.SetData(&entity.LicenseSettings{Key: keyText}); err != nil {
		return nil, false, hperrors.Wrap(err)
	}

	err = s.settingRepo.Upsert(ctx, db, setting, entity.SettingUpsertingConflictCols,
		entity.SettingUpsertingUpdateCols)
	if err != nil {
		return nil, false, hperrors.Wrap(err)
	}
	// As for every singleton setting: a first install that finds a row another
	// one has just committed is refused rather than leaving two licenses.
	if err = s.settingRepo.EnsureUnique(ctx, db, entity.NewObjectScopeGlobal(), base.SettingTypeLicense); err != nil {
		return nil, false, hperrors.Wrap(err)
	}
	return setting, created, nil
}
