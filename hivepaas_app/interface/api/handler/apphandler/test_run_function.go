package apphandler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

// testRunWriteTimeout lifts the server's write timeout for a test run: its
// first run on a node may install the libraries, which takes minutes. It is the
// use case's own limit and a minute for the answer.
const testRunWriteTimeout = 16 * time.Minute

// TestRunFunction Runs a function once
// @Summary Runs a function once
// @Description Calls the function once with the code sent, not yet saved, and a request, in a throwaway
// @Description container on a build node, with the function's variables and secrets.
// @Tags    Apps
// @Produce json
// @Id      testRunFunction
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   body body appdto.TestRunFunctionReq true "request data"
// @Success 200 {object} appdto.TestRunFunctionResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/function/test-run [post]
func (h *Handler) TestRunFunction(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdto.NewTestRunFunctionReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	if err = http.NewResponseController(ctx.Writer).SetWriteDeadline(time.Now().Add(testRunWriteTimeout)); err != nil {
		_ = ctx.Error(err)
	}
	resp, err := h.appUC.TestRunFunction(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
