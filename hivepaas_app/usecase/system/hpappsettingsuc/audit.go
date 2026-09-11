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

// The sections HivePaaS's own settings pages are recorded under.
//
// Named here rather than spelled at each call site because the confirmations and
// reverts have to carry the same value as the change they answer: that is what
// lets a reader follow one page through the whole story.
const (
	auditSectionRouting = "routing"
	auditSectionService = "service"

	// The two halves of the security page, which share one type: the switches,
	// and the credential that wraps every stored secret. See
	// AuditLogTypeHivePaaSSecuritySettingsUpdate.
	auditSectionSecuritySettings = "security-settings"
	auditSectionAppSecretRotate  = "app-secret-rotate" //nolint:gosec // G101: a section name
)

// auditSectionOfSetting maps a probation's setting type back to the page it was
// written on, so an answer is filed under the same section as the change.
func auditSectionOfSetting(settingType base.SettingType) string {
	switch settingType { //nolint:exhaustive // only the types that go on trial here
	case base.SettingTypeAppRouting:
		return auditSectionRouting
	case base.SettingTypeHivePaaSService:
		return auditSectionService
	default:
		return ""
	}
}

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
// Both outcomes, for the reason the security switches next door record both: a
// wrong app secret against an admin-only endpoint is somebody working from a
// session they should not have, and refusals are the only place that shows.
//
// Filed as a section of the security settings rather than under a type of its
// own - see AuditLogTypeHivePaaSSecuritySettingsUpdate for why the weight of the
// action is not by itself a reason for a type.
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
		Type:    base.AuditLogTypeHivePaaSSecuritySettingsUpdate,
		Scope:   base.ObjectScopeHivepaas,
		Source:  base.AuditLogSourceAPIUpdate,
		Section: auditSectionAppSecretRotate,
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
