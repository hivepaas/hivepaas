package apphandler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

// GetAppTerminalInfo Gets terminal info
// @Summary Gets terminal info
// @Description Gets terminal info
// @Tags    apps
// @Produce json
// @Id      getAppTerminalInfo
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Success 200 {object} appdto.GetTerminalInfoResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/terminal/info [get]
func (h *Handler) GetAppTerminalInfo(ctx *gin.Context) {
	auth, projectID, appID, err := h.GetAuth(ctx, base.ActionTypeRead, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdto.NewGetTerminalInfoReq()
	req.ProjectID = projectID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appUC.GetTerminalInfo(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// OpenAppTerminal Opens app terminal via websocket
// @Summary Opens app terminal via websocket
// @Description Opens app terminal via websocket
// @Tags    apps
// @Produce json
// @Id      getAppTerminal
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Success 200 {object} appdto.OpenTerminalResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/terminal [get]
func (h *Handler) OpenAppTerminal(ctx *gin.Context) {
	auth, projectID, appID, err := h.GetAuth(ctx, base.ActionTypeExecute, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdto.NewOpenTerminalReq()
	req.ProjectID = projectID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appUC.OpenTerminal(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	defer resp.CloseFunc()

	if !h.IsWebsocketRequest(ctx) || resp.ExecAttachResult == nil {
		ctx.JSON(http.StatusOK, resp)
		return
	}

	wsConn, err := h.UpgradeWebsocket(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	type ResizeMessage struct {
		Type   string `json:"type"`
		Width  uint   `json:"width"`
		Height uint   `json:"height"`
	}

	// Pipe container stdout/stderr (resp.ExecAttachResult.Reader) to websocket
	go func() {
		defer wsConn.Close()
		buf := make([]byte, 4096) //nolint:mnd
		for {
			n, err := resp.ExecAttachResult.Reader.Read(buf)
			if n > 0 {
				if err := wsConn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()

	// Pipe websocket messages to container stdin (resp.ExecAttachResult.Conn)
	for {
		mt, message, err := wsConn.ReadMessage()
		if err != nil {
			break
		}
		if mt == websocket.TextMessage { // control message
			var resizeMsg ResizeMessage
			if err := json.Unmarshal(message, &resizeMsg); err == nil && resizeMsg.Type == "resize" {
				_ = resp.ExecResizeFunc(h.RequestCtx(ctx), resizeMsg.Width, resizeMsg.Height)
				continue
			}
		}
		if _, err := resp.ExecAttachResult.Conn.Write(message); err != nil {
			break
		}
	}
}
