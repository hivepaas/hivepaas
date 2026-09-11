package appuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

// recordAppWrite records one write against an app.
//
// Called inside the caller's transaction and before it returns, so a record that
// cannot be written rolls the change back with it. A change that happened without
// a record is the gap the record exists to close.
//
// logType separates the three things that can happen to an app. The edits all
// share app-update, with section saying which endpoint made them, because an app
// is edited from a dozen endpoints and a type per endpoint would name the routing
// of the day it was written while filling the type filter with values nobody can
// use. Creation and removal get their own types: they are what somebody scanning
// that column is looking for, and they read as nothing at all folded into edits.
//
// detail may be nil, for a section with nothing to add beyond its own name.
func (uc *UC) recordAppWrite(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	app *entity.App,
	logType base.AuditLogType,
	source base.AuditLogSource,
	section string,
	detail *auditdetail.Builder,
) error {
	if app == nil {
		return hperrors.NewArgumentInvalid("audited app")
	}
	if detail == nil {
		detail = auditdetail.New()
	}

	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:     logType,
		Scope:    base.ObjectScopeApp,
		ObjectID: app.ID,
		Source:   source,
		Auth:     auth,
		ResType:  base.ResourceTypeApp,
		ResID:    app.ID,
		ResName:  app.Name,
		Detail:   detail.Set("section", section).String(),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// photoAction names what the request did to the photo, so the entry says which
// of the three it was rather than only that the photo moved.
func photoAction(req *appdto.AppPhotoReq) string {
	switch {
	case req.Delete:
		return "remove"
	case req.IsPresetIcon:
		return "preset-icon"
	default:
		return "upload"
	}
}
