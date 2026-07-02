package systemhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/taskuc/taskdto"
)

// ListTask Lists task
// @Summary Lists task
// @Description Lists task
// @Tags    system_tasks
// @Produce json
// @Id      listTask
// @Param   jobName query string false "`jobName=<system job name>`"
// @Param   targetId query string false "`taregtId=<target id>`"
// @Param   status query string false "`status=<target>`"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} taskdto.ListTaskResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/tasks [get]
func (h *Handler) ListTask(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeTask,
		Action:         base.ActionTypeRead,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := taskdto.NewListTaskReq()
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.taskUC.ListTask(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetTask Gets task
// @Summary Gets task
// @Description Gets task
// @Tags    system_tasks
// @Produce json
// @Id      getTask
// @Param   taskID path string true "task ID"
// @Success 200 {object} taskdto.GetTaskResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /tasks/{taskID} [get]
func (h *Handler) GetTask(ctx *gin.Context) {
	taskID, err := h.ParseStringParam(ctx, "taskID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeTask,
		ResourceID:     taskID,
		Action:         base.ActionTypeRead,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := taskdto.NewGetTaskReq()
	req.ID = taskID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.taskUC.GetTask(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetTaskStatus Gets task status
// @Summary Gets task status
// @Description Gets task status
// @Tags    system_tasks
// @Produce json
// @Produce plain
// @Id      getTaskStatus
// @Param   taskID path string true "task ID"
// @Success 200 {object} taskdto.GetTaskStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /tasks/{taskID}/status [get]
func (h *Handler) GetTaskStatus(ctx *gin.Context) {
	taskID, err := h.ParseStringParam(ctx, "taskID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeTask,
		ResourceID:     taskID,
		Action:         base.ActionTypeRead,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := taskdto.NewGetTaskStatusReq()
	req.ID = taskID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.taskUC.GetTaskStatus(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	if ctx.ContentType() == "text/plain" {
		ctx.String(http.StatusOK, "status=%v", resp.Data.Status)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// CancelTask Cancels a task
// @Summary Cancels a task
// @Description Cancels a task
// @Tags    system_tasks
// @Produce json
// @Id      cancelTask
// @Param   taskID path string true "task ID"
// @Param   body body taskdto.CancelTaskReq true "request data"
// @Success 200 {object} taskdto.CancelTaskResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/tasks/{taskID}/cancel [post]
func (h *Handler) CancelTask(ctx *gin.Context) {
	taskID, err := h.ParseStringParam(ctx, "taskID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeTask,
		ResourceID:     taskID,
		Action:         base.ActionTypeWrite,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := taskdto.NewCancelTaskReq()
	req.ID = taskID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.taskUC.CancelTask(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetTaskLogs Gets logs of a task
// @Summary Gets logs of a task
// @Description Gets logs of a task
// @Tags    system_tasks
// @Produce json
// @Id      getTaskLogs
// @Param   taskID path string true "task ID"
// @Success 200 {object} taskdto.GetTaskLogsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/tasks/{taskID}/logs [get]
func (h *Handler) GetTaskLogs(ctx *gin.Context) {
	taskID, err := h.ParseStringParam(ctx, "taskID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeTask,
		ResourceID:     taskID,
		Action:         base.ActionTypeRead,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := taskdto.NewGetTaskLogsReq()
	req.TaskID = taskID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	isWebsocketReq := h.IsWebsocketRequest(ctx)
	if !isWebsocketReq {
		req.Follow = false // Not a websocket request, we don't support `follow` flag
	}

	resp, err := h.taskUC.GetTaskLogs(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	if !isWebsocketReq {
		// Not a websocket request, return data via body
		ctx.JSON(http.StatusOK, resp)
	} else {
		h.StreamAppLogs(ctx, resp.Data.StaticLogs, resp.Data.LogsStream, resp.Data.LogsStreamCloser)
	}
}
