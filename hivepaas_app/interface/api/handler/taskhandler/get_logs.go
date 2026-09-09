package taskhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/taskuc/taskdto"
)

func (h *Handler) GetTaskLogs(
	ctx *gin.Context,
	scopeType base.ObjectScopeType,
	opts ...GetTaskOption,
) {
	var auth *basedto.Auth
	var itemID string
	var err error

	options := &GetTaskOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &entity.ObjectScope{ScopeType: scopeType}
	switch scopeType {
	case base.ObjectScopeProject:
		auth, scope.ProjectID, itemID, err = h.GetAuthProjectTasks(ctx, base.ActionTypeRead, "itemID")
	case base.ObjectScopeProjectEnv:
		auth, scope.ProjectID, scope.ProjectEnvID, itemID, err = h.GetAuthProjectEnvTasks(ctx,
			base.ActionTypeRead, "itemID")
	case base.ObjectScopeApp:
		auth, scope.ProjectID, scope.ProjectEnvID, scope.AppID, itemID, err = h.GetAuthAppTasks(ctx,
			base.ActionTypeRead, "itemID")
	case base.ObjectScopeUser:
		auth, scope.UserID, itemID, err = h.GetAuthUserTasks(ctx, base.ActionTypeRead, "itemID")
	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		auth, itemID, err = h.GetAuthGlobalTasks(ctx, base.ResourceTypeTask, base.ActionTypeRead, "itemID")
	default:
		err = hperrors.NewUnsupported("Task scope 'none'")
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := taskdto.NewGetTaskLogsReq()
	req.Scope, req.ID = scope, itemID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	if options.PreRequestHandler != nil {
		if err = options.PreRequestHandler(auth, req); err != nil {
			h.RenderError(ctx, err)
			return
		}
	}

	isWebsocketReq := h.IsWebsocketRequest(ctx)
	if !isWebsocketReq {
		req.Follow = false // Not a websocket request, we don't support `follow` flag
	}

	resp, err := h.TaskUC.GetTaskLogs(h.RequestCtx(ctx), auth, req)
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
