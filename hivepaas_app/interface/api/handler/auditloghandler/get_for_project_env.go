package auditloghandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// ListProjectEnvAuditLog Lists audit logs
// @Summary Lists audit logs
// @Description Lists audit logs
// @Tags    audit_logs
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
func (h *Handler) ListProjectEnvAuditLog(ctx *gin.Context) {
	h.ListAuditLog(ctx, base.ObjectScopeProjectEnv)
}

// GetProjectEnvAuditLog Gets audit log
// @Summary Gets audit log
// @Description Gets audit log
// @Tags    audit_logs
// @Produce json
// @Id      getProjectEnvAuditLog
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "log ID"
// @Success 200 {object} auditlogdto.GetAuditLogResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/audit-logs/{itemID} [get]
func (h *Handler) GetProjectEnvAuditLog(ctx *gin.Context) {
	h.GetAuditLog(ctx, base.ObjectScopeProjectEnv)
}
