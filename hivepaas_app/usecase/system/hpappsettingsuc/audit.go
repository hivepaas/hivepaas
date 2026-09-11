package hpappsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

// recordHivePaaSSettingsUpdate records a change to HivePaaS's own settings.
//
// Filed under the hivepaas scope, the same one the security settings use: these
// are the install's own configuration, and there is no project or app to put them
// under.
//
// section says which page was written - routing, service. detail may be nil.
func (uc *UC) recordHivePaaSSettingsUpdate(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	section string,
	detail *auditdetail.Builder,
) error {
	if detail == nil {
		detail = auditdetail.New()
	}

	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:    base.AuditLogTypeHivePaaSSettingsUpdate,
		Scope:   base.ObjectScopeHivepaas,
		Source:  base.AuditLogSourceAPIUpdate,
		Section: section,
		Auth:    auth,
		ResType: base.ResourceTypeSetting,
		ResName: section + " settings",
		Detail:  detail.String(),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// recordAppSecretRotation records an attempt at the app secret, allowed or not.
//
// Both outcomes, for the reason the security settings next door record both: a
// wrong app secret against an admin-only endpoint is somebody working from a
// session they should not have, and refusals are the only place that shows.
//
// Written after the rewrap and before the new secret reaches disk. A record that
// cannot be written stops there, which leaves the documented half-state - the row
// rewrapped, the file still holding the old secret - that the next start refuses
// loudly rather than opening quietly. That is the same direction UpdateAppSecret
// already fails in when saving the file goes wrong.
func (uc *UC) recordAppSecretRotation(
	ctx context.Context,
	auth *basedto.Auth,
	allowed bool,
) error {
	result := base.AuditLogResultAllowed
	if !allowed {
		result = base.AuditLogResultDenied
	}

	err := uc.auditService.Record(ctx, uc.db, &auditservice.Entry{
		Type:    base.AuditLogTypeAppSecretRotate,
		Scope:   base.ObjectScopeHivepaas,
		Source:  base.AuditLogSourceAPIUpdate,
		Result:  result,
		Auth:    auth,
		ResType: base.ResourceTypeSecuritySettings,
		ResName: "app secret",
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// settingsSnapshotData parses the pre-change settings a probation snapshot holds,
// so the entry can name the fields that moved. Unreadable data costs the field
// list, not the entry.
func settingsSnapshotData(settingType base.SettingType, snapshot entity.SettingSnapshot) entity.SettingData {
	if snapshot.Data == "" {
		return nil
	}
	setting := &entity.Setting{Type: settingType, Data: snapshot.Data}
	data, err := setting.Parse()
	if err != nil {
		return nil
	}
	return data
}
