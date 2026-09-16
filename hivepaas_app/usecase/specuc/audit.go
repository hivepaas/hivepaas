package specuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// recordSpecExport records a configuration spec being handed over.
//
// It runs after the bundle is built and before it is returned, and a failure to
// record aborts the export: handing the archive over without its record would
// leave exactly the gap the record exists to close.
//
// The resource is the exported scope itself - there is no stored object for an
// export to point at. The detail says what was taken and in what form, and
// never carries the passphrase.
func (uc *UC) recordSpecExport(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	scope *entity.ObjectScope,
	mode specmodel.SecretsMode,
	resp *specservice.ExportResp,
) error {
	detail := auditdetail.New().
		Set("secretsMode", string(mode)).
		Set("filename", resp.Filename).
		Set("sizeBytes", resp.Size)
	if resp.Summary != nil {
		detail.Set("files", resp.Summary.Files).Set("issues", resp.Summary.Issues)
	}

	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:     base.AuditLogTypeSpecExport,
		Scope:    scope.ScopeType,
		ObjectID: scope.ScopeObjectID(),
		Source:   base.AuditLogSourceAPIAction,
		Auth:     auth,
		ResType:  resourceTypeOfScope(scope.ScopeType),
		ResID:    scope.ScopeObjectID(),
		Detail:   detail.String(),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// resourceTypeOfScope names the object an export was taken from. The global
// scope has no object, and is left empty rather than given a made-up one.
func resourceTypeOfScope(scope base.ObjectScopeType) base.ResourceType {
	switch scope {
	case base.ObjectScopeProject:
		return base.ResourceTypeProject
	case base.ObjectScopeProjectEnv:
		return base.ResourceTypeProjectEnv
	case base.ObjectScopeApp:
		return base.ResourceTypeApp
	case base.ObjectScopeUser, base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		fallthrough
	default:
		return ""
	}
}
