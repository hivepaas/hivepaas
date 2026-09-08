package auditloghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/auditloguc/auditlogdto"
)

type GetAuditLogOptions struct {
	PreRequestHandler func(auth *basedto.Auth, req any) error
}

type GetAuditLogOption func(*GetAuditLogOptions)

func GetAuditLogPreRequestHandler(fn func(auth *basedto.Auth, req any) error) GetAuditLogOption {
	return func(opts *GetAuditLogOptions) {
		opts.PreRequestHandler = fn
	}
}

func (h *Handler) GetAuditLog(
	ctx *gin.Context,
	scopeType base.ObjectScopeType,
	opts ...GetAuditLogOption,
) {
	var auth *basedto.Auth
	var itemID string
	var err error

	options := &GetAuditLogOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &entity.ObjectScope{ScopeType: scopeType}
	switch scopeType {
	case base.ObjectScopeProject:
		auth, scope.ProjectID, itemID, err = h.GetAuthProjectAuditLogs(ctx, base.ActionTypeRead, "itemID")
	case base.ObjectScopeProjectEnv:
		auth, scope.ProjectID, scope.ProjectEnvID, itemID, err = h.GetAuthProjectEnvAuditLogs(ctx,
			base.ActionTypeRead, "itemID")
	case base.ObjectScopeApp:
		auth, scope.ProjectID, scope.ProjectEnvID, scope.AppID, itemID, err = h.GetAuthAppAuditLogs(ctx,
			base.ActionTypeRead, "itemID")
	case base.ObjectScopeUser:
		auth, scope.UserID, itemID, err = h.GetAuthUserAuditLogs(ctx, base.ActionTypeRead, "itemID")
	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		auth, itemID, err = h.GetAuthGlobalAuditLogs(ctx, base.ResourceTypeAuditLog, base.ActionTypeRead, "itemID")
	default:
		err = hperrors.NewUnsupported("AuditLog scope 'none'")
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := auditlogdto.NewGetAuditLogReq()
	req.Scope, req.ID = scope, itemID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	if options.PreRequestHandler != nil {
		if err = options.PreRequestHandler(auth, req); err != nil {
			h.RenderError(ctx, err)
			return
		}
	}

	resp, err := h.AuditLogUC.GetAuditLog(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
