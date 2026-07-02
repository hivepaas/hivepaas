package userhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
)

// GetUserInviteInfo Gets user invite info
// @Summary Gets user invite info
// @Description Gets user invite info
// @Tags    users
// @Produce json
// @Id      getUserInviteInfo
// @Success 200 {object} userdto.GetUserInviteInfoResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /users/invite/info [get]
func (h *Handler) GetUserInviteInfo(ctx *gin.Context) {
	auth, _, err := h.getAuth(ctx, "", base.ActionTypeWrite, false)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := userdto.NewGetUserInviteInfoReq()
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.userUC.GetUserInviteInfo(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// InviteUser Invites a user to system
// @Summary Invites a user to system
// @Description Invites a user to system
// @Tags    users
// @Produce json
// @Id      inviteUser
// @Param   body body userdto.InviteUserReq true "request data"
// @Success 200 {object} userdto.InviteUserResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /users/invite [post]
func (h *Handler) InviteUser(ctx *gin.Context) {
	auth, _, err := h.getAuth(ctx, "", base.ActionTypeWrite, false)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := userdto.NewInviteUserReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.userUC.InviteUser(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// BeginUserSignup Begins user signup process
// @Summary Begins user signup process
// @Description Begins user signup process
// @Tags    users
// @Produce json
// @Id      beginUserSignup
// @Param   body body userdto.BeginUserSignupReq true "request data"
// @Success 200 {object} userdto.BeginUserSignupResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /users/signup-begin [post]
func (h *Handler) BeginUserSignup(ctx *gin.Context) {
	req := userdto.NewBeginUserSignupReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.userUC.BeginUserSignup(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// CompleteUserSignup Completes user signup process
// @Summary Completes user signup process
// @Description Completes user signup process
// @Tags    users
// @Produce json
// @Id      completeUserSignup
// @Param   body body userdto.CompleteUserSignupReq true "request data"
// @Success 200 {object} userdto.CompleteUserSignupResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /users/signup-complete [post]
func (h *Handler) CompleteUserSignup(ctx *gin.Context) {
	req := userdto.NewCompleteUserSignupReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.userUC.CompleteUserSignup(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
