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
// authenticated yet.
func (uc *UC) IsEnabled(ctx context.Context) (bool, error) {
	current, err := uc.Current(ctx)
	return current.Enabled, err
}

// Current is the MCP setting as the endpoint acts on it: whether it serves, and
// whether its tools may change things. It asks the database at most once per
// enabledCacheTTL; a setting that cannot be read is off, and the error says why.
func (uc *UC) Current(ctx context.Context) (entity.MCPSettings, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if !uc.read.IsZero() && timeNow().Sub(uc.read) < enabledCacheTTL {
		return uc.current, nil
	}

	setting, err := uc.SettingRepo.GetSingle(ctx, uc.DB, entity.NewObjectScopeGlobal(), currentSettingType, false)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return entity.MCPSettings{}, hperrors.Wrap(err)
	}
	var current entity.MCPSettings
	if setting != nil && setting.IsActive() {
		data, err := setting.AsMCPSettings()
		if err != nil {
			return entity.MCPSettings{}, hperrors.Wrap(err)
		}
		current = *data
	}
	uc.current, uc.read = current, timeNow()
	return current, nil
}
