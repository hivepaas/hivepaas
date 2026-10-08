package apphandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

// ListAppInEnv Lists apps of an env
// @Summary Lists apps of an env
// @Description Lists apps of an env
// @Tags    Apps
// @Produce json
// @Id      listProjectEnvApp
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   status query string false "`status=<target>`"
// @Param   category query string false "`category=function`, or `category=webapp,database`; no kind is webapp"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Param   getChildApps query bool false "`getChildApps=true` to include apps that belong to another"
// @Param   getStats query bool false "`getStats=true` to include each app's stats"
// @Param   parentId query string false "`parentId=<app ID>` lists the preview apps of that app"
// @Success 200 {object} appdto.ListAppResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps [get]
func (h *Handler) ListAppInEnv(ctx *gin.Context) {
	auth, projectID, projectEnvID, _, err := h.GetAuthInEnv(ctx, base.ActionTypeRead, false)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdto.NewListAppReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appUC.ListApp(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetApp Gets app details
// @Summary Gets app details
// @Description Gets app details
// @Tags    Apps
// @Produce json
// @Id      getApp
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   getStats query bool false "`getStats=true` to include the app's stats"
// @Success 200 {object} appdto.GetAppResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID} [get]
func (h *Handler) GetApp(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdto.NewGetAppReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil { // to make sure Validate() to be called
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appUC.GetApp(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
