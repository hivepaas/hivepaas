package appsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/periodicjobuc/periodicjobdto"
)

// ListAppPeriodicJobTask Lists periodic job's tasks
// @Summary Lists periodic job's tasks
// @Description Lists periodic job's tasks
// @Tags    app_settings
// @Produce json
// @Id      listAppPeriodicJobTask
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} periodicjobdto.ListPeriodicJobTaskResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/periodic-jobs/{itemID}/tasks [get]
func (h *Handler) ListAppPeriodicJobTask(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, jobID, err := h.GetAuthAppSettings(ctx, base.ActionTypeRead, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := periodicjobdto.NewListPeriodicJobTaskReq()
	req.JobID = jobID
	req.Scope = entity.NewObjectScopeApp(appID, "", projectID, projectEnvID)
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.PeriodicJobUC.ListPeriodicJobTask(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
