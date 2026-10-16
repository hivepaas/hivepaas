package licenseserviceimpl

import (
	"context"
	"time"

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

// Install refuses only what could never be in force here: a key that does not
// verify, one for another installation, and one past its grace. Any other key
// replaces the one stored, a lower plan or a trial included: the person pasting
// it is the one who knows which key they mean.
func (s *service) Install(ctx context.Context, db database.Tx,
	req *licenseservice.InstallReq) (*licenseservice.InstallResp, error) {
	keyText := license.NormalizeKeyText(req.Key)
	state, err := s.installable(ctx, keyText)
	if err != nil {
		return nil, err
	}

	setting, err := s.loadSetting(ctx, db, bunex.SelectFor("UPDATE OF setting"))
	if err != nil {
		return nil, err
	}
	settings, err := settingsOf(setting)
	if err != nil {
		return nil, err
	}
	installKey(settings, keyText)
	setting, created, err := s.saveSettings(ctx, db, setting, settings)
	if err != nil {
		return nil, err
	}
	// The next Current verifies what is stored now. The cache would see the key
	// has changed anyway; dropping it also leaves nothing of the old key behind.
	s.dropVerification()
	return &licenseservice.InstallResp{Setting: setting, Created: created, State: state}, nil
}

// installable is the license a key would put in force here, or the refusal of a
// key that would not be in force: one that does not verify, is for another
// installation, or is past its grace. Its refusal of a key that does not verify
// is ErrLicenseInvalid, whose detail says what failed and never what the key
// holds.
func (s *service) installable(ctx context.Context, keyText string) (*license.State, error) {
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
	return state, nil
}

// installKey makes keyText the stored key. A key installed was not moved,
// whatever the one it replaces was; the refresh's schedule and record are the
// installation's, and are kept.
func installKey(settings *entity.LicenseSettings, keyText string) {
	settings.Key = keyText
	settings.MovedAt = time.Time{}
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

// settingsOf is the license a setting holds; none when there is no setting.
func settingsOf(setting *entity.Setting) (*entity.LicenseSettings, error) {
	if setting == nil {
		return &entity.LicenseSettings{}, nil
	}
	settings, err := setting.AsLicenseSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return settings, nil
}

// saveSettings writes settings as the global license setting, the one there is,
// over setting as it was read for update (nil when there was none), and reports
// whether it was written for the first time.
func (s *service) saveSettings(ctx context.Context, db database.Tx, setting *entity.Setting,
	settings *entity.LicenseSettings) (*entity.Setting, bool, error) {
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
	setting.Name = entity.LicenseSettingName
	setting.Version = entity.CurrentLicenseSettingsVersion
	setting.UpdatedAt = timeNow
	if err := setting.SetData(settings); err != nil {
		return nil, false, hperrors.Wrap(err)
	}

	err := s.settingRepo.Upsert(ctx, db, setting, entity.SettingUpsertingConflictCols,
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
