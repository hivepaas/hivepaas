package projectenvhandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/auditloguc/auditlogdto"
)

// ListAuditLog Lists audit logs
// @Summary Lists audit logs
// @Description Lists audit logs
// @Tags    project_env_audit_logs
// @Produce json
// @Id      listProjectEnvAuditLog
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} auditlogdto.ListAuditLogResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/audit-logs [get]
func (h *Handler) ListAuditLog(ctx *gin.Context) {
	h.AuditLogHandler.ListAuditLog(ctx, base.ObjectScopeProjectEnv)
}

// GetAuditLog Gets audit log
// @Summary Gets audit log
// @Description Gets audit log
// @Tags    project_env_audit_logs
// @Produce json
// @Id      getProjectEnvAuditLog
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "log ID"
// @Success 200 {object} auditlogdto.GetAuditLogResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/audit-logs/{itemID} [get]
func (h *Handler) GetAuditLog(ctx *gin.Context) {
	h.AuditLogHandler.GetAuditLog(ctx, base.ObjectScopeProjectEnv)
}

// ListAuditLogTypes Lists audit log types
// @Summary Lists audit log types
// @Description Lists audit log types
// @Tags    project_env_audit_logs
// @Produce json
// @Id      listProjectEnvAuditLogType
// @Param   projectID path string true "project ID"
// @Success 200 {object} auditlogdto.ListAuditLogTypeResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/audit-logs/types [get]
func (h *Handler) ListAuditLogTypes(ctx *gin.Context) {
	h.AuditLogHandler.ListAuditLogTypes(ctx, base.ObjectScopeProjectEnv)
}
