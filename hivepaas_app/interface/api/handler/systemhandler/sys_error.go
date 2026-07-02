package systemhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/syserroruc/syserrordto"
)

// ListSysError Lists sys errors
// @Summary Lists sys errors
// @Description Lists sys errors
// @Tags    system_errors
// @Produce json
// @Id      listSysError
// @Param   status query int false "`status=<target>`"
// @Param   code query string false "`code=<target>`"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} syserrordto.ListSysErrorResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/errors [get]
func (h *Handler) ListSysError(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeSysError,
		Action:         base.ActionTypeRead,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := syserrordto.NewListSysErrorReq()
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.sysErrorUC.ListSysError(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetSysError Gets sys error
// @Summary Gets sys error
// @Description Gets sys error
// @Tags    system_errors
// @Produce json
// @Id      getSysError
// @Param   errorID path string true "error ID"
// @Success 200 {object} syserrordto.GetSysErrorResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/errors/{errorID} [get]
func (h *Handler) GetSysError(ctx *gin.Context) {
	errorID, err := h.ParseStringParam(ctx, "errorID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeSysError,
		ResourceID:     errorID,
		Action:         base.ActionTypeRead,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := syserrordto.NewGetSysErrorReq()
	req.ID = errorID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil { // to make sure Validate() to be called
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.sysErrorUC.GetSysError(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// DeleteSysError Deletes sys error
// @Summary Deletes sys error
// @Description Deletes sys error
// @Tags    system_errors
// @Produce json
// @Id      deleteSysError
// @Param   errorID path string true "error ID"
// @Success 200 {object} syserrordto.DeleteSysErrorResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/errors/{errorID} [delete]
func (h *Handler) DeleteSysError(ctx *gin.Context) {
	errorID, err := h.ParseStringParam(ctx, "errorID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeSysError,
		ResourceID:     errorID,
		Action:         base.ActionTypeDelete,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := syserrordto.NewGetSysErrorReq()
	req.ID = errorID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil { // to make sure Validate() to be called
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.sysErrorUC.GetSysError(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
