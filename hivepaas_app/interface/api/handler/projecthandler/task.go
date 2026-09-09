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
// @Tags    project_tasks
// @Produce json
// @Id      listProjectTask
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
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
// @Tags    project_tasks
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
// @Tags    project_tasks
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
// @Tags    project_tasks
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
// @Tags    project_tasks
// @Produce json
// @Id      getProjectTaskLogs
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "task ID"
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
// @Tags    project_tasks
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
