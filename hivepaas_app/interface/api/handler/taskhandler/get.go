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

type GetTaskOptions struct {
	PreRequestHandler func(auth *basedto.Auth, req any) error
}

type GetTaskOption func(*GetTaskOptions)

func GetTaskPreRequestHandler(fn func(auth *basedto.Auth, req any) error) GetTaskOption {
	return func(opts *GetTaskOptions) {
		opts.PreRequestHandler = fn
	}
}

func (h *Handler) GetTask(
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

	req := taskdto.NewGetTaskReq()
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

	resp, err := h.TaskUC.GetTask(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
