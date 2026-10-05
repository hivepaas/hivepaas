package appmetricshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmetricsuc/appmetricsdto"
)

// GetAppHTTPMetrics Gets an app's HTTP metrics
// @Summary Gets an app's HTTP metrics
// @Description Counts an app's requests over a range ending now, from the proxy's access log: how many,
// @Description how many the client got a 4xx and a 5xx for, how many the proxy could not get to the app at
// @Description all (unreachable), and how long they took end to end - p50, p95 and p99, in milliseconds,
// @Description close rather than exact. In totals, by method and path (numbers and ids replaced by :n and
// @Description :id), by replica, and as series: one point per step, oldest first, as for a function's
// @Description metrics. `available` is false, with a `reason`, when they cannot be counted: the app has no
// @Description domain (not-exposed), the proxy's access log is off, not JSON or unlabelled, or the logs
// @Description cannot be read.
// @Tags    Apps
// @Produce json
// @Id      getAppHTTPMetrics
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   range query string false "1h, 6h, 24h or 7d; default 24h"
// @Success 200 {object} appmetricsdto.GetAppHTTPMetricsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/http-metrics [get]
func (h *Handler) GetAppHTTPMetrics(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appmetricsdto.NewGetAppHTTPMetricsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appMetricsUC.GetAppHTTPMetrics(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetAppResourceMetrics Gets an app's resource metrics
// @Summary Gets an app's resource metrics
// @Description Reads an app's containers' CPU and memory over a range ending now, from the rows the agent on
// @Description each node writes every 15 seconds: CPU in cores and memory in bytes - the working set, as
// @Description `docker stats` counts it - each with its limit (0 for none), OOM kills, and network and disk in
// @Description bytes a second. One point per step, oldest first, the containers summed; totals over the
// @Description range; and the containers, the last seen first. `available` is false, with a `reason`, when
// @Description they cannot be read: the agent does not mark its lines yet (agent-unlabelled, until its next
// @Description update), or the logs cannot be read.
// @Tags    Apps
// @Produce json
// @Id      getAppResourceMetrics
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   range query string false "1h, 6h, 24h or 7d; default 24h"
// @Success 200 {object} appmetricsdto.GetAppResourceMetricsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/resource-metrics [get]
func (h *Handler) GetAppResourceMetrics(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appmetricsdto.NewGetAppResourceMetricsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appMetricsUC.GetAppResourceMetrics(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetAppRouteMetrics Gets an app's routes
// @Summary Gets an app's routes
// @Description Reads what an app served over a range ending now, from the rows the agent on each node writes
// @Description from OBI (eBPF): every request its containers answered, from the proxy or from inside the
// @Description project, by kind (http or rpc), method and route - a template such as /users/{id} when the
// @Description app's framework names one. How many, how many failed (a 5xx, or an error), and p50, p95 and p99
// @Description in milliseconds, close rather than exact: in totals, by route, and as series, one point per
// @Description step, oldest first. `nodes` and `nodesCovered` count the nodes the app runs on now and those
// @Description that run OBI. `available` is false, with a `reason`, when they cannot be read: the collection
// @Description is off (performance-disabled) or off for the app (app-disabled), no node the app runs on runs
// @Description OBI (node-disabled) or can (node-unsupported, with `preflightReasons`), the agent does not mark
// @Description its lines yet (agent-unlabelled), or the logs cannot be read.
// @Tags    Apps
// @Produce json
// @Id      getAppRouteMetrics
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   range query string false "1h, 6h, 24h or 7d; default 24h"
// @Success 200 {object} appmetricsdto.GetAppRouteMetricsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/route-metrics [get]
func (h *Handler) GetAppRouteMetrics(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appmetricsdto.NewGetAppPerformanceMetricsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appMetricsUC.GetAppRouteMetrics(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetAppDependencyMetrics Gets what an app calls
// @Summary Gets what an app calls
// @Description Reads what an app called over a range ending now, from the rows the agent on each node writes
// @Description from OBI (eBPF): by kind (http, db or rpc), with series, one point per step, oldest first; and
// @Description by peer - a host and port as the app named it, or for a database its system and database, as
// @Description postgresql/shop - with the env's app behind it when one is known, and its methods or
// @Description operations. How many, how many failed, and p50, p95 and p99 in milliseconds, close rather than
// @Description exact. `available` is false, with a `reason`, as for the app's routes.
// @Tags    Apps
// @Produce json
// @Id      getAppDependencyMetrics
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   range query string false "1h, 6h, 24h or 7d; default 24h"
// @Success 200 {object} appmetricsdto.GetAppDependencyMetricsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/dependency-metrics [get]
func (h *Handler) GetAppDependencyMetrics(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appmetricsdto.NewGetAppPerformanceMetricsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appMetricsUC.GetAppDependencyMetrics(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetFunctionMetrics Gets a function's metrics
// @Summary Gets a function's metrics
// @Description Counts a function's calls over a range ending now, from the invocation line its runtime writes
// @Description for every call into its logs: how many, how many failed - an outcome other than ok - how many
// @Description the handler answered 5xx, and the handler's duration's p50, p95 and p99, in milliseconds, which
// @Description are close rather than exact. One point per step, oldest first: 1 min for 1h, 5 min for 6h,
// @Description 15 min for 24h, 1 h for 7d. When the logs cannot be read, `available` is false and `reason`
// @Description says why, as for the logs' history.
// @Tags    Apps
// @Produce json
// @Id      getFunctionMetrics
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   range query string false "1h, 6h, 24h or 7d; default 24h"
// @Success 200 {object} appmetricsdto.GetFunctionMetricsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/function-metrics [get]
func (h *Handler) GetFunctionMetrics(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appmetricsdto.NewGetFunctionMetricsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appMetricsUC.GetFunctionMetrics(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
