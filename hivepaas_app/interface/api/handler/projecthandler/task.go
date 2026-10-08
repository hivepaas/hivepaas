package projecthandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/taskuc/taskdto"
)

// ListTask Lists tasks
// @Summary Lists tasks
// @Description Lists tasks
// @Tags    Project tasks
// @Produce json
// @Id      listProjectTask
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Param   appId query string false "`appId=<app ID>` narrows to one app's tasks"
// @Param   fromDate query string false "`fromDate=YYYY-MM-DD`"
// @Param   projectEnvId query string false "`projectEnvId=<project ID>:<env>` narrows to one environment's tasks"
// @Param   projectId query string false "`projectId=<project ID>` narrows to one project's tasks"
// @Param   scopeOnly query bool false "`scopeOnly=true` leaves out the tasks of what the scope holds"
// @Param   status query string false "`status=<status>`, comma separated"
// @Param   targetId query string false "`targetId=<ID>`, comma separated: the objects the tasks act on"
// @Param   toDate query string false "`toDate=YYYY-MM-DD`"
// @Param   type query string false "`type=<task type>`, comma separated"
// @Success 200 {object} taskdto.ListTaskResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/tasks [get]
func (h *Handler) ListTask(ctx *gin.Context) {
	h.TaskHandler.ListTask(ctx, base.ObjectScopeProject)
}

// GetTask Gets task
// @Summary Gets task
// @Description Gets task
// @Tags    Project tasks
// @Produce json
// @Id      getProjectTask
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "task ID"
// @Success 200 {object} taskdto.GetTaskResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/tasks/{itemID} [get]
func (h *Handler) GetTask(ctx *gin.Context) {
	h.TaskHandler.GetTask(ctx, base.ObjectScopeProject)
}

// GetTaskStatus Gets task status
// @Summary Gets task status
// @Description Gets task status
// @Tags    Project tasks
// @Produce json
// @Id      getProjectTaskStatus
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "task ID"
// @Success 200 {object} taskdto.GetTaskStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/tasks/{itemID}/status [get]
func (h *Handler) GetTaskStatus(ctx *gin.Context) {
	h.TaskHandler.GetTaskStatus(ctx, base.ObjectScopeProject)
}

// ListTaskType Lists task types
// @Summary Lists task types
// @Description Lists task types
// @Tags    Project tasks
// @Produce json
// @Id      listProjectTaskType
// @Param   projectID path string true "project ID"
// @Success 200 {object} taskdto.ListTaskTypeResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/tasks/types [get]
func (h *Handler) ListTaskType(ctx *gin.Context) {
	h.TaskHandler.ListTaskType(ctx, base.ObjectScopeProject)
}

// GetTaskLogs Gets task logs
// @Summary Gets task logs
// @Description Gets task logs
// @Tags    Project tasks
// @Produce json
// @Id      getProjectTaskLogs
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "task ID"
// @Param   duration query string false "`duration=1h` for the logs within that period"
// @Param   follow query string false "`follow=true/false`"
// @Param   since query string false "`since=YYYY-MM-DDTHH:mm:SSZ`"
// @Param   tail query int false "`tail=1000` for the last 1000 lines"
// @Success 200 {object} taskdto.GetTaskLogsReq
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/tasks/{itemID}/logs [get]
func (h *Handler) GetTaskLogs(ctx *gin.Context) {
	h.TaskHandler.GetTaskLogs(ctx, base.ObjectScopeProject)
}

// CancelTask Cancels task
// @Summary Cancels task
// @Description Cancels task
// @Tags    Project tasks
// @Produce json
// @Id      cancelProjectTask
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "task ID"
// @Success 200 {object} taskdto.CancelTaskReq
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/tasks/{itemID}/cancel [post]
func (h *Handler) CancelTask(ctx *gin.Context) {
	h.TaskHandler.CancelTask(ctx, base.ObjectScopeProject)
}

// ListTaskTargetObject Lists task target objects
// @Summary Lists task target objects
// @Description Lists task target objects
// @Tags    Project tasks
// @Produce json
// @Id      listProjectTaskTargetObject
// @Param   projectID path string true "project ID"
// @Success 200 {object} taskdto.ListTargetObjectsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/tasks/target-objects [get]
func (h *Handler) ListTaskTargetObject(ctx *gin.Context) {
	h.TaskHandler.ListTargetObject(ctx, base.ObjectScopeProject)
}
