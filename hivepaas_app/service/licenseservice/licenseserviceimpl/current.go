package licenseserviceimpl

import (
	"context"
	"crypto/sha256"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
)

// verification is Verify's answer for a key, kept for as long as neither the key
// nor whether development keys count changes. Verifying is the one costly step
// of reading the license, and the answer cannot change otherwise.
type verification struct {
	keySum   [sha256.Size]byte
	allowDev bool
	verified *license.Verified
	err      error
}

// Current reads the setting on every call. That is cheap, as creations are
// rare, and it is how a key installed by another process - the app or the
// worker, or phase 2's refresh - is in force here at once, with no event or
// expiry to miss.
func (s *service) Current(ctx context.Context) (*license.State, error) {
	settings, err := s.loadSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	now, err := s.now(ctx)
	if err != nil {
		return nil, err
	}
	if settings.Key == "" {
		// No key is Bee whatever the installation, so docker is not asked.
		return license.Evaluate(&license.EvalInput{Now: now}), nil
	}

	installationID, err := s.InstallationID(ctx)
	if err != nil {
		return nil, err
	}
	verified, verifyErr := s.verifyStored(settings.Key)
	return license.Evaluate(&license.EvalInput{
		KeyText:        settings.Key,
		Verified:       verified,
		VerifyErr:      verifyErr,
		InstallationID: installationID,
		Now:            now,
		MovedAt:        settings.MovedAt,
	}), nil
}

// loadSettings is the stored license, empty when there is none.
func (s *service) loadSettings(ctx context.Context, db database.IDB) (*entity.LicenseSettings, error) {
	setting, err := s.loadSetting(ctx, db)
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return &entity.LicenseSettings{}, nil
	}
	settings, err := setting.AsLicenseSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return settings, nil
}

// loadSetting is the license setting, nil when there is none.
func (s *service) loadSetting(ctx context.Context, db database.IDB,
	opts ...bunex.SelectQueryOption) (*entity.Setting, error) {
	setting, err := s.settingRepo.GetSingle(ctx, db, entity.NewObjectScopeGlobal(), base.SettingTypeLicense, false,
		opts...)
	if errors.Is(err, hperrors.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return setting, nil
}

// verifyStored is Verify's answer for the stored key, verified again only when
// the key or whether development keys count has changed.
func (s *service) verifyStored(keyText string) (*license.Verified, error) {
	keySum := sha256.Sum256([]byte(keyText))
	allowDev := s.allowDev()

	s.mu.Lock()
	defer s.mu.Unlock()
	if v := s.verified; v != nil && v.keySum == keySum && v.allowDev == allowDev {
		return v.verified, v.err
	}
	verified, err := s.verify(keyText, allowDev)
	s.verified = &verification{keySum: keySum, allowDev: allowDev, verified: verified, err: err}
	return verified, err
}

// dropVerification forgets the stored key's verification, once another key is
// stored.
func (s *service) dropVerification() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.verified = nil
}

// verify checks a key against the keys this binary trusts. A key set that does
// not parse is a build mistake, ErrLicenseSigningKeyInvalid, and is answered for
// the key like any other refusal: the license is then invalid, which leaves the
// installation on Bee rather than unable to create anything.
func (s *service) verify(keyText string, allowDev bool) (*license.Verified, error) {
	trust, err := s.trust(allowDev)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	verified, err := license.Verify(trust, keyText)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return verified, nil
}
