package auditloghandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/auditloguc/auditlogdto"
)

// ListGlobalAuditLog Lists audit logs
// @Summary Lists audit logs
// @Description Lists audit logs
// @Tags    audit_logs
// @Produce json
// @Id      listAuditLog
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} auditlogdto.ListAuditLogResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/audit-logs [get]
func (h *Handler) ListGlobalAuditLog(ctx *gin.Context) {
	h.ListAuditLog(ctx, base.ObjectScopeGlobal)
}

// GetGlobalAuditLog Gets audit log
// @Summary Gets audit log
// @Description Gets audit log
// @Tags    audit_logs
// @Produce json
// @Id      getAuditLog
// @Param   itemID path string true "log ID"
// @Success 200 {object} auditlogdto.GetAuditLogResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/audit-logs/{itemID} [get]
func (h *Handler) GetGlobalAuditLog(ctx *gin.Context) {
	h.GetAuditLog(ctx, base.ObjectScopeGlobal)
}
