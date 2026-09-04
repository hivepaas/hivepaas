package datakeyserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
)

// Rewrap re-encrypts the stored data key with a new app secret.
//
// This is the entire cost of changing the app secret: one row. The values in the
// settings table are encrypted with the data key, which does not change, so none
// of them is touched - no bulk pass, nothing to resume, no window in which some
// rows carry one key and some another.
func (s *service) Rewrap(
	ctx context.Context,
	db database.IDB,
	currentAppSecret, newAppSecret string,
) error {
	return transaction.Execute(ctx, db, func(db database.Tx) error { //nolint:wrapcheck
		stored, err := s.encryptionKeyRepo.GetActive(ctx, db, bunex.SelectFor("UPDATE"))
		if err != nil {
			return hperrors.Wrap(err)
		}
		if stored == nil {
			return hperrors.NewNotFound("Data encryption key")
		}

		// Unwrapping with the current secret is what proves the caller gave the
		// right one: it either opens the key or it does not.
		key, err := datakey.Unwrap(stored.WrappedKey, currentAppSecret)
		if err != nil {
			return hperrors.Wrap(hperrors.ErrForbidden).WithCause(err).
				WithMsgLog("the current app secret does not open the stored data encryption key")
		}

		wrapped, err := key.Wrap(newAppSecret)
		if err != nil {
			return hperrors.Wrap(err)
		}

		stored.WrappedKey = wrapped
		stored.UpdatedAt = timeutil.NowUTC()
		if err := s.encryptionKeyRepo.Update(ctx, db, stored,
			bunex.UpdateColumns("wrapped_key", "updated_at")); err != nil {
			return hperrors.Wrap(err)
		}

		// The key itself is unchanged, so nothing in memory needs replacing.
		return nil
	})
}
