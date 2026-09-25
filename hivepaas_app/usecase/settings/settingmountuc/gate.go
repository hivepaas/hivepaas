package settingmountuc

import (
	"context"
	"encoding/json"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

// activeGrants are what an entry hands its app now: nothing while disabled.
func activeGrants(setting *entity.Setting) []settingmountservice.Grant {
	if setting == nil || setting.Status != base.SettingStatusActive {
		return nil
	}
	mount, err := setting.AsAppSettingMount()
	if err != nil {
		return nil
	}
	return settingmountservice.Grants(mount)
}

// authorizeGrants passes §7's gate for what after hands out that before did not.
// Mounting a private key or a password is revealing it: whoever controls the
// container reads the file.
//
// It is asked on the database, not on the save's transaction: a denial rolls
// the save back, and the record of the attempt has to outlive it.
func (uc *UC) authorizeGrants(
	ctx context.Context, auth *basedto.Auth, scope *entity.ObjectScope,
	before, after *entity.Setting, source base.AuditLogSource,
) error {
	added := settingmountservice.Widens(activeGrants(before), activeGrants(after))
	if len(added) == 0 {
		return nil
	}
	detail, err := json.Marshal(map[string]any{"grants": added})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(uc.PermissionManager.AuthorizeSecretReveal(ctx, uc.DB, auth, &permission.RevealSubject{
		Scope:    base.ObjectScopeApp,
		ObjectID: scope.AppID,
		Source:   source,
		ResType:  base.ResourceTypeSettingMount,
		ResID:    after.ID,
		ResName:  after.Name,
		Detail:   string(detail),
	}))
}
