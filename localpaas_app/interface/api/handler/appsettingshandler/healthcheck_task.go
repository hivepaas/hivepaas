package appsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/healthcheckuc/healthcheckdto"
)

// ListAppHealthcheckTask Lists healthcheck tasks
// @Summary Lists healthcheck tasks
// @Description Lists healthcheck tasks
// @Tags    app_settings
// @Produce json
// @Id      listAppHealthcheckTask
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} healthcheckdto.ListHealthcheckTaskResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/healthchecks/{itemID}/tasks [get]
func (h *Handler) ListAppHealthcheckTask(ctx *gin.Context) {
	auth, projectID, appID, jobID, err := h.GetAuthAppSettings(ctx, base.ActionTypeRead, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := healthcheckdto.NewListHealthcheckTaskReq()
	req.JobID = jobID
	req.Scope = base.NewObjectScopeApp(appID, projectID)
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.HealthcheckUC.ListHealthcheckTask(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
