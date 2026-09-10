package systemhandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/taskuc/taskdto"
)

// ListTask Lists tasks
// @Summary Lists tasks
// @Description Lists tasks
// @Tags    system_tasks
// @Produce json
// @Id      listTask
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} taskdto.ListTaskResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/tasks [get]
func (h *Handler) ListTask(ctx *gin.Context) {
	h.taskHandler.ListTask(ctx, base.ObjectScopeGlobal)
}

// GetTask Gets task
// @Summary Gets task
// @Description Gets task
// @Tags    system_tasks
// @Produce json
// @Id      getTask
// @Param   itemID path string true "task ID"
// @Success 200 {object} taskdto.GetTaskResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/tasks/{itemID} [get]
func (h *Handler) GetTask(ctx *gin.Context) {
	h.taskHandler.GetTask(ctx, base.ObjectScopeGlobal)
}

// GetTaskStatus Gets task status
// @Summary Gets task status
// @Description Gets task status
// @Tags    system_tasks
// @Produce json
// @Id      getTaskStatus
// @Param   itemID path string true "task ID"
// @Success 200 {object} taskdto.GetTaskStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/tasks/{itemID}/status [get]
func (h *Handler) GetTaskStatus(ctx *gin.Context) {
	h.taskHandler.GetTaskStatus(ctx, base.ObjectScopeGlobal)
}

// ListTaskType Lists task types
// @Summary Lists task types
// @Description Lists task types
// @Tags    system_tasks
// @Produce json
// @Id      listTaskType
// @Success 200 {object} taskdto.ListTaskTypeResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/tasks/types [get]
func (h *Handler) ListTaskType(ctx *gin.Context) {
	h.taskHandler.ListTaskType(ctx, base.ObjectScopeGlobal)
}

// GetTaskLogs Gets task logs
// @Summary Gets task logs
// @Description Gets task logs
// @Tags    system_tasks
// @Produce json
// @Id      getTaskLogs
// @Param   itemID path string true "task ID"
// @Success 200 {object} taskdto.GetTaskLogsReq
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/tasks/{itemID}/logs [get]
func (h *Handler) GetTaskLogs(ctx *gin.Context) {
	h.taskHandler.GetTaskLogs(ctx, base.ObjectScopeGlobal)
}

// CancelTask Cancels task
// @Summary Cancels task
// @Description Cancels task
// @Tags    system_tasks
// @Produce json
// @Id      cancelTask
// @Param   itemID path string true "task ID"
// @Success 200 {object} taskdto.CancelTaskReq
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/tasks/{itemID}/cancel [post]
func (h *Handler) CancelTask(ctx *gin.Context) {
	h.taskHandler.CancelTask(ctx, base.ObjectScopeGlobal)
}

// ListTaskTargetObject Lists task target objects
// @Summary Lists task target objects
// @Description Lists task target objects
// @Tags    system_tasks
// @Produce json
// @Id      listTaskTargetObject
// @Success 200 {object} taskdto.ListTargetObjectsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/tasks/target-objects [get]
func (h *Handler) ListTaskTargetObject(ctx *gin.Context) {
	h.taskHandler.ListTargetObject(ctx, base.ObjectScopeGlobal)
}
