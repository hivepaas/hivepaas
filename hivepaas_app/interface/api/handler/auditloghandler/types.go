package auditloghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/auditloguc/auditlogdto"
)

func (h *Handler) ListAuditLogTypes(ctx *gin.Context, scopeType base.ObjectScopeType) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := auditlogdto.NewListAuditLogTypeReq()
	req.Scope = &entity.ObjectScope{ScopeType: scopeType}
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.AuditLogUC.ListAuditLogType(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
