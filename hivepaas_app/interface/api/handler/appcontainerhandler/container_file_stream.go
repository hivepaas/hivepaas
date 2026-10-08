package appcontainerhandler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appcontaineruc/appcontainerdto"
)

const (
	// transferIdle is how long a copy into or out of a container may go without
	// a byte before it ends: one that goes on, however slowly, is not cut.
	transferIdle = 60 * time.Second
	// streamMessageMax is the largest message a client may send in a stream.
	streamMessageMax = 4 << 20
)

// errStreamEnded is the pipe's end when the client went before its end message.
var errStreamEnded = errors.New("the client ended the stream before its end message")

// copyDownload copies r to the response, pushing the connection's write
// deadline idle past each write: the server's write timeout would cut a long
// copy however well it went. On an error, with the headers sent, it closes the
// connection, so that the client reads a cut and not an end.
func copyDownload(w http.ResponseWriter, r io.Reader, idle time.Duration) error {
	rc := http.NewResponseController(w)
	_, err := io.Copy(writerFunc(func(p []byte) (int, error) {
		_ = rc.SetWriteDeadline(time.Now().Add(idle))
		return w.Write(p) //nolint:wrapcheck // the copy's error
	}), r)
	if err != nil {
		if conn, _, hijackErr := rc.Hijack(); hijackErr == nil {
			_ = conn.Close()
		}
		return hperrors.Wrap(err)
	}
	return nil
}

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

// streamControl is a text message of a stream, from the client or to it.
type streamControl struct {
	Type  string                                         `json:"type"`
	Data  *appcontainerdto.UploadFileToContainerDataResp `json:"data,omitempty"`
	Error *hperrors.ErrorInfo                            `json:"error,omitempty"`
}

const (
	streamEnd   = "end"
	streamDone  = "done"
	streamError = "error"
)

// pumpUpload writes what a client sends over conn into w until its end message:
// binary messages are the content, {"type":"end"} its end. A client silent for
// idle ends the copy, with an error; so does a reader that stopped. It closes w
// either way, with the error when there is one.
func pumpUpload(conn *websocket.Conn, w *io.PipeWriter, idle time.Duration) (err error) {
	defer func() { _ = w.CloseWithError(err) }()
	conn.SetReadLimit(streamMessageMax)
	for {
		if err = conn.SetReadDeadline(time.Now().Add(idle)); err != nil {
			return hperrors.Wrap(err)
		}
		kind, message, readErr := conn.ReadMessage()
		if readErr != nil {
			if websocket.IsCloseError(readErr, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return errStreamEnded
			}
			return hperrors.Wrap(readErr)
		}
		switch kind {
		case websocket.BinaryMessage:
			if _, err = w.Write(message); err != nil {
				return err //nolint:wrapcheck // the reader's own error, as it stopped
			}
		case websocket.TextMessage:
			var control streamControl
			if json.Unmarshal(message, &control) == nil && control.Type == streamEnd {
				return nil
			}
		}
	}
}

// StreamFileToContainer Uploads a file or archive into container over a websocket
// @Summary Uploads a file or archive into container over a websocket
// @Description Takes what file-upload takes, as query parameters, then the content as binary messages
// @Description and {"type":"end"} as a text message; answers {"type":"done","data":{...}} or
// @Description {"type":"error","error":{...}}. Neither the server's nor Traefik's timeouts cut it.
// @Tags    Apps
// @Produce json
// @Id      streamFileToAppContainer
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   nodeId query string false "node ID"
// @Param   containerId query string false "container ID (optional, auto-picks active container if empty)"
// @Param   path query string true "file/dir path in container"
// @Param   extract query boolean false "extract archive into path"
// @Param   compressionFormat query string false "compression format (gzip, zstd, zip, tar)"
// @Param   overwrite query boolean false "allow overwrite (default: true)"
// @Param   fileName query string false "the file's name"
// @Param   fileSize query integer false "the file's size in bytes: required unless extract"
// @Success 101
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/container/file-upload/stream [get]
func (h *Handler) StreamFileToContainer(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := appcontainerdto.NewUploadFileToContainerReq()
	req.ProjectID, req.ProjectEnvID, req.AppID = projectID, projectEnvID, appID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}
	req.FileName = ctx.Query("fileName")
	if !req.Extract {
		// The tar a single file goes in starts with its size.
		if req.FileSize, err = strconv.ParseInt(ctx.Query("fileSize"), 10, 64); err != nil || req.FileSize < 0 {
			h.RenderError(ctx, hperrors.NewArgumentInvalid("fileSize").
				WithMsgLog("fileSize is required for a file that is not extracted"))
			return
		}
	}
	if !h.IsWebsocketRequest(ctx) {
		h.RenderError(ctx, hperrors.Wrap(hperrors.ErrBadRequest).WithMsgLog("the stream is a websocket"))
		return
	}

	conn, err := h.UpgradeWebsocket(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	defer conn.Close()

	pr, pw := io.Pipe()
	req.FileContent = pr
	copyCtx, cancel := context.WithCancel(h.RequestCtx(ctx))
	defer cancel()
	type outcome struct {
		resp *appcontainerdto.UploadFileToContainerResp
		err  error
	}
	copied := make(chan outcome, 1)
	safego.Go("appcontainer.streamFileToContainer", func() {
		resp, err := h.appContainerUC.UploadFileToContainer(copyCtx, auth, req)
		// What the copy did not read stops the pump.
		_ = pr.CloseWithError(errors.Join(err, io.ErrClosedPipe))
		copied <- outcome{resp, err}
	})

	pumpErr := pumpUpload(conn, pw, transferIdle)
	result := <-copied
	answer := streamControl{Type: streamDone}
	switch {
	case result.err != nil:
		answer = streamControl{Type: streamError, Error: h.ErrorInfoOf(ctx, result.err)}
	case pumpErr != nil:
		answer = streamControl{Type: streamError, Error: h.ErrorInfoOf(ctx, pumpErr)}
	case result.resp != nil:
		answer.Data = result.resp.Data
	}
	if message, err := json.Marshal(answer); err == nil {
		_ = conn.WriteMessage(websocket.TextMessage, message)
	}
	_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		time.Now().Add(time.Second))
}
