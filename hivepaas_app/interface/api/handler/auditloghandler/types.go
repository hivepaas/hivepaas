package auditloghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/auditloguc/auditlogdto"
)

// ListAuditLogTypes lists available audit log types
// @Summary Lists available audit log types
// @Description Lists available audit log types
// @Tags    audit_logs
// @Produce json
// @Id      listAuditLogTypes
// @Success 200 {object} auditlogdto.ListAuditLogTypeResp
// @Router  /system/audit-logs/types [get]
func (h *Handler) ListAuditLogTypes(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := auditlogdto.NewListAuditLogTypeReq()
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
