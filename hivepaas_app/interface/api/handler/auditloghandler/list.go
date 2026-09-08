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

type ListAuditLogOptions struct {
	PreRequestHandler func(auth *basedto.Auth, req any) error
}

type ListAuditLogOption func(*ListAuditLogOptions)

func ListAuditLogPreRequestHandler(fn func(auth *basedto.Auth, req any) error) ListAuditLogOption {
	return func(opts *ListAuditLogOptions) {
		opts.PreRequestHandler = fn
	}
}

func (h *Handler) ListAuditLog(
	ctx *gin.Context,
	scopeType base.ObjectScopeType,
	opts ...ListAuditLogOption,
) {
	var auth *basedto.Auth
	var err error

	options := &ListAuditLogOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &entity.ObjectScope{ScopeType: scopeType}
	switch scopeType {
	case base.ObjectScopeProject:
		auth, scope.ProjectID, _, err = h.GetAuthProjectAuditLogs(ctx, base.ActionTypeRead, "")
	case base.ObjectScopeProjectEnv:
		auth, scope.ProjectID, scope.ProjectEnvID, _, err = h.GetAuthProjectEnvAuditLogs(ctx, base.ActionTypeRead, "")
	case base.ObjectScopeApp:
		auth, scope.ProjectID, scope.ProjectEnvID, scope.AppID, _, err = h.GetAuthAppAuditLogs(ctx, base.ActionTypeRead, "")
	case base.ObjectScopeUser:
		auth, scope.UserID, _, err = h.GetAuthUserAuditLogs(ctx, base.ActionTypeRead, "")
	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		auth, _, err = h.GetAuthGlobalAuditLogs(ctx, base.ResourceTypeAuditLog, base.ActionTypeRead, "")
	default:
		err = hperrors.NewUnsupported("AuditLog scope 'none'")
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := auditlogdto.NewListAuditLogReq()
	req.Scope = scope
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	if options.PreRequestHandler != nil {
		if err = options.PreRequestHandler(auth, req); err != nil {
			h.RenderError(ctx, err)
			return
		}
	}

	resp, err := h.AuditLogUC.ListAuditLog(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
