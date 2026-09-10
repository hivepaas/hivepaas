package settings

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

// recordSettingAudit records one write against a stored setting.
//
// It lives here, called from the seven generic flows, rather than in each of the
// twenty-odd setting usecases. Repeated per type it would be a record that fails
// open: forget it on one usecase and that setting's changes stop being recorded,
// silently, with nothing about the code looking wrong. The same argument as
// revealSecrets, for the same reason.
//
// Called inside the flow's transaction and before its event fires, so a record
// that cannot be written rolls the change back with it, and no subscriber acts on
// a change that was never recorded.
//
// Only writes that happened are recorded. A refusal never reaches here: write
// permission is settled in the handler, before the usecase runs. That is a real
// gap - a denied attempt is the only sign of somebody trying doors - and closing
// it means recording at the auth helpers, which is its own change.
func (uc *BaseUC) recordSettingAudit(
	ctx context.Context,
	db database.IDB,
	req *BaseSettingReq,
	logType base.AuditLogType,
	source base.AuditLogSource,
	old *entity.Setting,
	current *entity.Setting,
) error {
	setting := gofn.Coalesce(current, old)
	if setting == nil {
		return hperrors.NewArgumentInvalidNT("audited setting")
	}
	if req.Auth == nil {
		// Not a caller mistake to paper over: an unattributed entry answers none
		// of the questions the entry exists for, so refuse the write instead.
		return hperrors.NewArgumentInvalidNT("audit auth")
	}

	err := uc.AuditService.Record(ctx, db, &auditservice.Entry{
		Type:     logType,
		Scope:    req.Scope.ScopeType,
		ObjectID: req.Scope.ScopeObjectID(),
		Source:   source,
		Result:   base.AuditLogResultAllowed,
		Auth:     req.Auth,
		ResType:  base.ResourceType(setting.Type),
		ResID:    setting.ID,
		ResName:  setting.Name,
		Detail:   settingAuditDetail(setting, old, current),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// settingAuditDetail describes the change: the setting envelope before and after,
// and the names of the payload fields that moved.
//
// No value out of Setting.Data ever appears here, only field names - see
// auditdetail.FieldChanges for why inferring which of them are safe to print is
// not possible across the twenty-odd setting types, and why an audit row, which
// outlives the setting it describes, is the wrong place to be wrong about it.
//
// The envelope list is written out rather than reflected over entity.Setting, so
// a field added to Setting later cannot start being recorded without somebody
// deciding it should.
func settingAuditDetail(setting, old, current *entity.Setting) string {
	detail := auditdetail.New().
		Set("settingType", setting.Type).
		Set("kind", setting.Kind).
		Set("status", setting.Status)

	if old == nil || current == nil {
		return detail.String()
	}

	detail.
		Compare("name", old.Name, current.Name).
		Compare("status", old.Status, current.Status).
		Compare("inheritable", old.Inheritable, current.Inheritable).
		Compare("default", old.Default, current.Default).
		Compare("expireAt", formatAuditTime(old.ExpireAt), formatAuditTime(current.ExpireAt)).
		WithChangedFields(parseSettingData(old), parseSettingData(current))

	return detail.String()
}

// parseSettingData is the payload in its typed form, or nil if it cannot be read.
// Unreadable data costs the field list, not the entry.
func parseSettingData(setting *entity.Setting) entity.SettingData {
	data, err := setting.Parse()
	if err != nil {
		return nil
	}
	return data
}

func formatAuditTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
