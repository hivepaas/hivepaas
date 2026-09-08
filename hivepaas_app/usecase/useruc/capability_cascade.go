package useruc

import (
	"context"
	"encoding/json"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

// cascadeRevokedCapabilities follows through on the capabilities this update
// takes away.
//
// Taking away cap::api-key::create only stops the user minting new keys; the
// ones they already have keep working for up to a year, which is not what
// anybody means by revoking it. So the keys go with the capability.
//
// It runs inside the update's transaction: the capability and the keys are one
// change, and half of it landing is worse than none of it.
func (uc *UC) cascadeRevokedCapabilities(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	target *entity.User,
	persistingData *userservice.PersistingUserData,
) error {
	for _, capability := range revokedCapabilities(persistingData) {
		if capability != base.ResourceCapAPIKeyCreate {
			continue
		}
		if err := uc.revokeUserAPIKeys(ctx, db, auth, target, persistingData); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

// revokedCapabilities returns the capabilities the update actually takes away.
//
// A row being deleted is not enough on its own: a wholesale replace deletes every
// row it is about to write back, so a capability the request still asks for shows
// up in both lists. Only the ones that leave and do not return count.
func revokedCapabilities(persistingData *userservice.PersistingUserData) []base.ResourceCapability {
	kept := make(map[string]struct{}, len(persistingData.UpsertingAccesses))
	for _, access := range persistingData.UpsertingAccesses {
		if access.ResourceType == base.ResourceTypeCapability {
			kept[access.ResourceID] = struct{}{}
		}
	}

	var revoked []base.ResourceCapability
	for _, access := range persistingData.DeletingAccesses {
		if access.ResourceType != base.ResourceTypeCapability {
			continue
		}
		if _, ok := kept[access.ResourceID]; ok {
			continue
		}
		revoked = append(revoked, base.ResourceCapability(access.ResourceID))
	}
	return revoked
}

// revokeUserAPIKeys soft-deletes every live API key the user holds.
//
// Deleted rather than disabled: the status endpoint is not gated by the
// capability, so a disabled key is one request away from working again - by the
// very user it was taken from. The row survives the soft delete, so the record of
// what existed is not lost with it.
func (uc *UC) revokeUserAPIKeys(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	target *entity.User,
	persistingData *userservice.PersistingUserData,
) error {
	apiKeys, _, err := uc.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAPIKey),
		bunex.SelectWhere("setting.object_id = ?", target.ID),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	for _, apiKey := range apiKeys {
		apiKey.DeletedAt = timeNow
		apiKey.UpdatedAt = timeNow
		apiKey.UpdateVer++
		persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, apiKey)

		// The actor is whoever ran the update, not the key's owner, so the owner
		// is recorded alongside it - otherwise the entry says a key was revoked
		// without saying whose.
		detail, err := json.Marshal(map[string]string{
			"ownerId":   target.ID,
			"ownerName": target.Username,
			"reason":    string(base.ResourceCapAPIKeyCreate) + " revoked",
		})
		if err != nil {
			return hperrors.Wrap(err)
		}

		err = uc.auditService.Record(ctx, db, &auditservice.Entry{
			Type:     base.AuditLogTypeAPIKeyRevoke,
			Scope:    base.ObjectScopeUser,
			ObjectID: target.ID,
			Source:   base.AuditLogSourceCapabilityRevoked,
			Result:   base.AuditLogResultAllowed,
			Auth:     auth,
			ResType:  base.ResourceTypeAPIKey,
			ResID:    apiKey.ID,
			ResName:  apiKey.Name,
			Detail:   string(detail),
		})
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	return nil
}
