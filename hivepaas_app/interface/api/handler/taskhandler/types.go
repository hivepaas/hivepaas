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

func (h *Handler) ListTaskType(
	ctx *gin.Context,
	scopeType base.ObjectScopeType,
) {
	var auth *basedto.Auth
	var err error

	scope := &entity.ObjectScope{ScopeType: scopeType}
	switch scopeType {
	case base.ObjectScopeProject:
		auth, scope.ProjectID, _, err = h.GetAuthProjectTasks(ctx, base.ActionTypeRead, "")
	case base.ObjectScopeProjectEnv:
		auth, scope.ProjectID, scope.ProjectEnvID, _, err = h.GetAuthProjectEnvTasks(ctx, base.ActionTypeRead, "")
	case base.ObjectScopeApp:
		auth, scope.ProjectID, scope.ProjectEnvID, scope.AppID, _, err = h.GetAuthAppTasks(ctx, base.ActionTypeRead, "")
	case base.ObjectScopeUser:
		auth, scope.UserID, _, err = h.GetAuthUserTasks(ctx, base.ActionTypeRead, "")
	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		auth, _, err = h.GetAuthGlobalTasks(ctx, base.ResourceTypeTask, base.ActionTypeRead, "")
	default:
		err = hperrors.NewUnsupported("Task scope 'none'")
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := taskdto.NewListTaskTypeReq()
	req.Scope = scope
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.TaskUC.ListTaskType(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
