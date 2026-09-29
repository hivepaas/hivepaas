package domainserviceimpl

import (
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// preferredSetting is the one of several settings obtaining a certificate goes
// through: the nearest to the scope, then the one marked default, then the
// oldest. A list comes back from the database in no order, and the first of it
// is a different DNS provider from one request to the next - one that may not
// manage the domain, and a failure an authority then holds against it.
func preferredSetting(settings []*entity.Setting, scope *entity.ObjectScope) *entity.Setting {
	if len(settings) == 0 {
		return nil
	}
	return slices.MinFunc(settings, func(a, b *entity.Setting) int {
		if d := scopeDistance(a, scope) - scopeDistance(b, scope); d != 0 {
			return d
		}
		if a.Default != b.Default {
			if a.Default {
				return -1
			}
			return 1
		}
		if c := a.CreatedAt.Compare(b.CreatedAt); c != 0 {
			return c
		}
		return strings.Compare(a.ID, b.ID)
	})
}

// scopeDistance is how far a setting's owner is from the scope: 0 for the scope's
// own, rising through its parents to the installation's, which is the farthest.
func scopeDistance(setting *entity.Setting, scope *entity.ObjectScope) int {
	if scope == nil {
		return 0
	}
	owners := []string{scope.AppID, scope.ParentAppID, scope.ProjectEnvID, scope.ProjectID}
	for i, owner := range owners {
		if owner != "" && setting.ObjectID == owner {
			return i
		}
	}
	return len(owners)
}
