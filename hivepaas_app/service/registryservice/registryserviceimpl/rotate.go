package registryserviceimpl

import (
	"context"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
)

// credentialGracePeriod is how long the previous password keeps working after a
// rotation. A swarm service presents the credential it was deployed with, so this
// is the window an operator has to redeploy their apps.
const credentialGracePeriod = 14 * 24 * time.Hour

// htpasswdFor returns the account file for the password in use, reusing the line
// that is already in it.
//
// bcrypt is salted, so hashing the same password twice gives two different lines.
// Generating one on every save would rewrite the secret, replace the swarm object
// and restart the registry for no reason at all - so the line that verifies
// against the password in use is kept, and a new one is written only when the
// file has none.
//
// keepPrevious carries the lines that verify against nothing current, which is
// what a rotation left behind. They go when the grace period ends.
func htpasswdFor(existing, password string, keepPrevious bool) (string, error) {
	current, previous := splitHtpasswd(existing, password)
	if current == "" {
		line, err := htpasswdLine(password)
		if err != nil {
			return "", hperrors.Wrap(err)
		}
		current = line
	}
	if !keepPrevious {
		previous = nil
	}
	return htpasswdContent(append([]string{current}, previous...)...), nil
}

// splitHtpasswd separates the line that verifies against password from whatever
// else the file holds.
func splitHtpasswd(existing, password string) (current string, previous []string) {
	for _, line := range strings.Split(existing, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		_, hash, found := strings.Cut(line, ":")
		if found && bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil {
			current = line
			continue
		}
		previous = append(previous, line)
	}
	return current, previous
}

// RotateCredential issues a new password and leaves the old one working.
func (s *service) RotateCredential(
	ctx context.Context,
	db database.IDB,
	req *registryservice.RotateCredentialReq,
) (*registryservice.RotateCredentialResp, error) {
	setting := req.Setting
	if setting == nil {
		return nil, hperrors.Wrap(hperrors.ErrRegistryNotConfigured)
	}
	cfg, err := setting.AsRegistrySettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !cfg.Enabled || cfg.RegistryAuthID == "" {
		return nil, hperrors.Wrap(hperrors.ErrRegistryNotConfigured)
	}

	app, err := s.loadApp(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if app == nil {
		return nil, hperrors.Wrap(hperrors.ErrRegistryNotConfigured).
			WithExtraDetail("The registry has not been provisioned yet.")
	}

	credential, err := s.settingRepo.GetByID(ctx, db, entity.NewObjectScopeGlobal(),
		base.SettingTypeRegistryAuth, cfg.RegistryAuthID, false)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	password, err := generatePassword()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	line, err := htpasswdLine(password)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Everything in the file today was issued for the password being replaced, so
	// all of it is the previous credential.
	existing, err := s.currentHtpasswd(app)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	content := htpasswdContent(append([]string{line}, nonEmptyLines(existing)...)...)
	if err = s.applyAccount(ctx, db, app, content); err != nil {
		return nil, hperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	if err = updateRegistryAuthSetting(credential, cfg.Domain, password, timeNow); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = s.settingRepo.Upsert(ctx, db, credential,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols); err != nil {
		return nil, hperrors.Wrap(err)
	}

	graceEnds := timeNow.Add(credentialGracePeriod)
	cfg.CredentialRotatedAt, cfg.CredentialGraceEnds = timeNow, graceEnds
	if err = setting.SetData(cfg); err != nil {
		return nil, hperrors.Wrap(err)
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeNow
	if err = s.settingRepo.Upsert(ctx, db, setting,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols); err != nil {
		return nil, hperrors.Wrap(err)
	}

	s.logger.Info("rotated the system registry credential", "graceEndsAt", graceEnds)
	return &registryservice.RotateCredentialResp{GraceEndsAt: graceEnds}, nil
}

func nonEmptyLines(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// currentHtpasswd reads the account file the app is running with.
func (s *service) currentHtpasswd(app *entity.App) (string, error) {
	setting := settingNamed(app.GetSettingsByType(base.SettingTypeSecret), registrySecretName)
	if setting == nil {
		return "", nil
	}
	secret, err := setting.AsSecret()
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	plain, err := secret.Value.GetPlain()
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return plain, nil
}
