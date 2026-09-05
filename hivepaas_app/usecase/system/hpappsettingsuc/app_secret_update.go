package hpappsettingsuc

import (
	"context"
	"sync"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/secrethelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

// secretRotationMu serializes rotations within this process.
var secretRotationMu sync.Mutex

// UpdateAppSecret changes the app secret.
//
// The app secret does not encrypt the stored values directly: it wraps the data
// encryption key, which is what encrypts them. So this rewraps one row and stops.
// The settings table is not touched, which is why there is no batching here, no
// progress to resume and no window where some rows carry one key and some another.
//
// The order matters: the database row is rewrapped first, and only then is the
// new secret written to disk. A crash between the two leaves the row wrapped with
// the new secret while the file still holds the old one, which fails loudly at
// the next start instead of quietly opening nothing.
func (uc *UC) UpdateAppSecret(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.UpdateAppSecretReq,
) (*hpappsettingsdto.UpdateAppSecretResp, error) {
	if auth.User.IsDemoUser() {
		return nil, hperrors.Wrap(hperrors.ErrUserDemoUnauthorized)
	}

	if req.NewSecret == req.CurrentSecret {
		return nil, hperrors.Wrap(hperrors.ErrBadRequest).
			WithExtraDetail("the new app secret must differ from the current one")
	}

	strengthReqs := hpappservice.SecretRequirements
	strengthReqs.PrevSecrets = []string{req.CurrentSecret}
	if err := secrethelper.ValidateStrength(req.NewSecret, &strengthReqs); err != nil {
		return nil, hperrors.Wrap(err)
	}

	if !secretRotationMu.TryLock() {
		return nil, hperrors.Wrap(hperrors.ErrConflict).
			WithMsgLog("an app secret change is already running")
	}
	defer secretRotationMu.Unlock()

	// Rewrap verifies the current secret by using it: it either opens the stored
	// key or it does not, so there is no separate comparison to get wrong.
	if err := uc.dataKeyService.Rewrap(ctx, uc.db, req.CurrentSecret, req.NewSecret); err != nil {
		return nil, hperrors.Wrap(err)
	}

	if err := config.SaveAppSecret(req.NewSecret); err != nil {
		// The row is already rewrapped, so the old secret no longer opens it. Say so
		// plainly: the operator has to put the new secret in place by hand.
		return nil, hperrors.Wrap(err).
			WithMsgLog("the data encryption key was rewrapped but the new app secret could not be saved")
	}

	// Other processes sharing this app path - the worker above all - hold the old
	// secret. They only need it to unwrap the key at startup, so a reload is
	// enough; nothing they already decrypted becomes invalid.
	uc.reloadHivepaasConfig(ctx)

	return &hpappsettingsdto.UpdateAppSecretResp{}, nil
}

// reloadHivepaasConfig asks the other HivePaaS processes to re-read their config.
// Best effort: a process that misses the signal keeps working with the key it
// already unwrapped, and picks the new secret up when it next restarts.
func (uc *UC) reloadHivepaasConfig(ctx context.Context) {
	if err := uc.hpAppService.ReloadHpAppConfig(ctx); err != nil {
		logging.Warnf("app secret change: failed to reload the app config: %v", err)
	}
}
