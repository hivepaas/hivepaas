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

type ListTaskOptions struct {
	PreRequestHandler func(auth *basedto.Auth, req any) error
}

type ListTaskOption func(*ListTaskOptions)

func ListTaskPreRequestHandler(fn func(auth *basedto.Auth, req any) error) ListTaskOption {
	return func(opts *ListTaskOptions) {
		opts.PreRequestHandler = fn
	}
}

func (h *Handler) ListTask(
	ctx *gin.Context,
	scopeType base.ObjectScopeType,
	opts ...ListTaskOption,
) {
	var auth *basedto.Auth
	var err error

	options := &ListTaskOptions{}
	for _, o := range opts {
		o(options)
	}

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

	req := taskdto.NewListTaskReq()
	req.Scope = scope
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	if options.PreRequestHandler != nil {
		if err = options.PreRequestHandler(auth, req); err != nil {
			h.RenderError(ctx, err)
			return
		}
	}

	resp, err := h.TaskUC.ListTask(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
