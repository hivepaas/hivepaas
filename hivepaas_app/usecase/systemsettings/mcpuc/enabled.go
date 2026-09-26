package mcpuc

import (
	"context"
	"errors"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// timeNow is swapped in tests.
var timeNow = time.Now

// IsEnabled reports whether the MCP endpoint serves, for a request no one has
// authenticated yet. It asks the database at most once per enabledCacheTTL; a
// setting that cannot be read is off, and the error says why.
func (uc *UC) IsEnabled(ctx context.Context) (bool, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if !uc.enabledRead.IsZero() && timeNow().Sub(uc.enabledRead) < enabledCacheTTL {
		return uc.enabled, nil
	}

	setting, err := uc.SettingRepo.GetSingle(ctx, uc.DB, entity.NewObjectScopeGlobal(), currentSettingType, false)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return false, hperrors.Wrap(err)
	}
	enabled := false
	if setting != nil && setting.IsActive() {
		data, err := setting.AsMCPSettings()
		if err != nil {
			return false, hperrors.Wrap(err)
		}
		enabled = data.Enabled
	}
	uc.enabled, uc.enabledRead = enabled, timeNow()
	return enabled, nil
}
