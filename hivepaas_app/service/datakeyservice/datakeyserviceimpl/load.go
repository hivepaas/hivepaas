package datakeyserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

// Load installs the data encryption key, creating one the first time the app runs.
//
// Unwrapping needs the app secret, so a wrong or changed one fails here rather
// than surfacing later as unreadable settings.
func (s *service) Load(ctx context.Context, db database.IDB) error {
	appSecret := config.Current.Secret
	if appSecret == "" {
		return hperrors.NewMissing("App secret")
	}

	stored, err := s.encryptionKeyRepo.GetActive(ctx, db)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if stored != nil {
		key, err := datakey.Unwrap(stored.WrappedKey, appSecret)
		if err != nil {
			return hperrors.Wrap(hperrors.ErrPreconditionFailed).WithCause(err).
				WithMsgLog("the app secret does not open the stored data encryption key")
		}
		datakey.SetActive(key)
		return nil
	}

	key, err := s.createKey(ctx, db, appSecret)
	if err != nil {
		return hperrors.Wrap(err)
	}
	datakey.SetActive(key)
	return nil
}

func (s *service) createKey(
	ctx context.Context,
	db database.IDB,
	appSecret string,
) (*datakey.Key, error) {
	key, err := datakey.Generate()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	wrapped, err := key.Wrap(appSecret)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	err = transaction.Execute(ctx, db, func(db database.Tx) error {
		// Another replica may have created one between the read above and here.
		// The unique index on is_active would reject the insert; take theirs.
		existing, err := s.encryptionKeyRepo.GetActive(ctx, db)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if existing != nil {
			key, err = datakey.Unwrap(existing.WrappedKey, appSecret)
			return hperrors.Wrap(err)
		}

		return s.encryptionKeyRepo.Insert(ctx, db, &entity.EncryptionKey{
			ID:         gofn.Must(ulid.NewStringULID()),
			WrappedKey: wrapped,
			IsActive:   true,
			CreatedAt:  timeNow,
			UpdatedAt:  timeNow,
		})
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return key, nil
}
