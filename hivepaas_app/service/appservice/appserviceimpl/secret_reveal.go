package appserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
)

// secretDecrypter is implemented by every setting type that stores secrets.
//
// Reaching the decrypt through this interface rather than through each type's own
// MustAsX() is what lets one place cover all of them - and what makes a setting
// type added later covered without anybody remembering to come back here.
type secretDecrypter interface {
	Decrypt() error
}

// RevealSecrets hands out the setting's secrets in the clear, if the caller asked
// for them and may have them.
//
// This lives here, once, rather than in each setting's own GetX. Repeated per
// type it is a check that fails open: forgetting it on one of fifteen usecases
// publishes that setting's secrets, silently, and nothing about the code looks
// wrong. Here, forgetting is not an option a future setting type has.
func (s *service) RevealSecrets(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	app *entity.App,
	setting *entity.Setting,
) (revealed bool, err error) {
	if setting == nil {
		return false, nil
	}
	// An inherited setting is read through, not owned, by this scope; its secrets
	// belong to whoever defined it.
	if setting.ObjectID != app.ID {
		return false, nil
	}

	settingData, err := setting.Parse()
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	decrypter, ok := settingData.(secretDecrypter)
	if !ok {
		return false, nil // the type holds no secrets, so there is nothing to reveal
	}

	err = s.permissionManager.AuthorizeSecretReveal(ctx, db, auth, &permission.RevealSubject{
		Scope:    base.ObjectScopeApp,
		ObjectID: app.ID,
		Source:   base.AuditLogSourceAPIGet,
		ResType:  base.ResourceTypeSetting,
		ResID:    setting.ID,
		ResName:  setting.Name,
	})
	if err != nil {
		return false, hperrors.Wrap(err)
	}

	if err = decrypter.Decrypt(); err != nil {
		return false, hperrors.Wrap(err)
	}
	return true, nil
}
