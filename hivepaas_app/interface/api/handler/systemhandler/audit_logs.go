package systemhandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/auditloguc/auditlogdto"
)

// ListAuditLog Lists audit logs
// @Summary Lists audit logs
// @Description Lists audit logs
// @Tags    System audit logs
// @Produce json
// @Id      listAuditLog
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Param   actorId query string false "`actorId=<user ID>`, comma separated: who acted"
// @Param   appId query string false "`appId=<app ID>` narrows to one app's entries"
// @Param   fromDate query string false "`fromDate=YYYY-MM-DD`"
// @Param   projectEnvId query string false "`projectEnvId=<project ID>:<env>` narrows to one environment's entries"
// @Param   projectId query string false "`projectId=<project ID>` narrows to one project's entries"
// @Param   resourceId query string false "`resourceId=<ID>`, comma separated: what was acted on"
// @Param   result query string false "`result=<result>`, comma separated"
// @Param   scopeOnly query bool false "`scopeOnly=true` leaves out the entries of what the scope holds"
// @Param   section query string false "`section=<section>`, comma separated"
// @Param   source query string false "`source=<source>`, comma separated"
// @Param   toDate query string false "`toDate=YYYY-MM-DD`"
// @Param   type query string false "`type=<audit log type>`, comma separated"
// @Success 200 {object} auditlogdto.ListAuditLogResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/audit-logs [get]
func (h *Handler) ListAuditLog(ctx *gin.Context) {
	h.auditLogHandler.ListAuditLog(ctx, base.ObjectScopeGlobal)
}

// GetAuditLog Gets audit log
// @Summary Gets audit log
// @Description Gets audit log
// @Tags    System audit logs
// @Produce json
// @Id      getAuditLog
// @Param   itemID path string true "log ID"
// @Success 200 {object} auditlogdto.GetAuditLogResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/audit-logs/{itemID} [get]
func (h *Handler) GetAuditLog(ctx *gin.Context) {
	h.auditLogHandler.GetAuditLog(ctx, base.ObjectScopeGlobal)
}

// ListAuditLogTypes Lists audit log types
// @Summary Lists audit log types
// @Description Lists audit log types
// @Tags    System audit logs
// @Produce json
// @Id      listAuditLogType
// @Success 200 {object} auditlogdto.ListAuditLogTypeResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/audit-logs/types [get]
func (h *Handler) ListAuditLogTypes(ctx *gin.Context) {
	h.auditLogHandler.ListAuditLogTypes(ctx, base.ObjectScopeGlobal)
}
