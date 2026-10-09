package appcontainerhandler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"runtime/debug"
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
	// a byte - out - or a message - in - before it ends: one that goes on,
	// however slowly, is not cut.
	transferIdle = 60 * time.Second
	// streamMessageMax is the largest message a client may send in a stream.
	streamMessageMax = 4 << 20
)

// errStreamEnded is the pipe's end when the client went before its end message.
var errStreamEnded = errors.New("the client ended the stream before its end message")

// nothingCopiedError is copyDownload's failure before a byte was sent: the
// response is still the handler's to write.
type nothingCopiedError struct{ err error }

func (e *nothingCopiedError) Error() string { return e.err.Error() }
func (e *nothingCopiedError) Unwrap() error { return e.err }

// copyDownload copies r to the response, pushing the connection's write
// deadline idle past each write: the server's write timeout would cut a long
// copy however well it went. A copy that fails once bytes are sent closes the
// connection, so that the client reads a cut and not an end; one that fails
// before is an nothingCopiedError, the response left to the caller.
func copyDownload(w http.ResponseWriter, r io.Reader, idle time.Duration) error {
	conn := http.NewResponseController(underlying(w))
	written, err := io.Copy(writerFunc(func(p []byte) (int, error) {
		_ = conn.SetWriteDeadline(time.Now().Add(idle))
		return w.Write(p) //nolint:wrapcheck // the copy's error
	}), r)
	if err == nil {
		return nil
	}
	if written == 0 {
		return &nothingCopiedError{err: hperrors.Wrap(err)}
	}
	// gin will not hand over a connection it has written to: the server's will.
	if raw, _, hijackErr := conn.Hijack(); hijackErr == nil {
		_ = raw.Close()
	}
	return hperrors.Wrap(err)
}

// underlying is the server's own writer under the wrappers - gin's - that keep
// from it what this needs: its connection, after a write.
func underlying(w http.ResponseWriter) http.ResponseWriter {
	for {
		inner, ok := w.(interface{ Unwrap() http.ResponseWriter })
		if !ok {
			return w
		}
		w = inner.Unwrap()
	}
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
	streamEnd      = "end"
	streamDone     = "done"
	streamError    = "error"
	streamProgress = "progress"
)

// streamProgressMessage tells a client that asked how much of its upload the
// copy has taken.
type streamProgressMessage struct {
	Type     string `json:"type"`
	Received int64  `json:"received"`
}

// answerDrain is how long the server reads what a client still sends after the
// answer, for the client's close: a connection closed with bytes unread is
// reset, and a reset can cost the client the answer before it.
const answerDrain = 2 * time.Second

// pumpUpload writes what a client sends over conn into w until its end message:
// binary messages are the content, {"type":"end"} its end. A client silent for
// idle ends the copy, with an error; so does a reader that stopped. It closes w
// either way, with the error when there is one. With progress, the client is
// told after each message how much the reader has taken: a write to the pipe
// returns once it has.
func pumpUpload(conn *websocket.Conn, w *io.PipeWriter, idle time.Duration, progress bool) (err error) {
	defer func() { _ = w.CloseWithError(err) }()
	conn.SetReadLimit(streamMessageMax)
	var received int64
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
			received += int64(len(message))
			if progress {
				if err = conn.SetWriteDeadline(time.Now().Add(idle)); err != nil {
					return hperrors.Wrap(err)
				}
				if err = conn.WriteJSON(streamProgressMessage{Type: streamProgress, Received: received}); err != nil {
					return hperrors.Wrap(err)
				}
			}
		case websocket.TextMessage:
			var control streamControl
			if json.Unmarshal(message, &control) == nil && control.Type == streamEnd {
				return nil
			}
		}
	}
}

// clientError is a stream's end the client caused - it left, it went silent -
// as the 400 it is, not a 500 of the server's; nil for any other.
func clientError(err error) error {
	var netErr net.Error
	if errors.Is(err, errStreamEnded) || (errors.As(err, &netErr) && netErr.Timeout()) ||
		websocket.IsUnexpectedCloseError(err) {
		return hperrors.Wrap(hperrors.ErrBadRequest).WithCause(err).
			WithMsgLog("the client ended the upload: %v", err)
	}
	return nil
}

// exactReader is a single file's content, held to the size the client said: the
// tar it goes in starts with that size, and a size of 0 must not let a client
// fill the server's memory with what it sends after.
type exactReader struct {
	r    io.Reader
	left int64
}

func (e *exactReader) Read(p []byte) (int, error) {
	if e.left == 0 {
		// One byte more than said is an error, not an end.
		var extra [1]byte
		if n, _ := io.ReadFull(e.r, extra[:]); n > 0 {
			return 0, hperrors.NewArgumentInvalid("fileSize").
				WithMsgLog("the client sent more than the fileSize it gave")
		}
		return 0, io.EOF
	}
	if int64(len(p)) > e.left {
		p = p[:e.left]
	}
	n, err := e.r.Read(p)
	e.left -= int64(n)
	if errors.Is(err, io.EOF) && e.left > 0 {
		return n, hperrors.NewArgumentInvalid("fileSize").
			WithMsgLog("the client sent %d bytes fewer than the fileSize it gave", e.left)
	}
	return n, err //nolint:wrapcheck // the pipe's own
}

// uploadFunc copies an upload into the container: the use case, but for tests.
type uploadFunc func(context.Context, *appcontainerdto.UploadFileToContainerReq) (
	*appcontainerdto.UploadFileToContainerResp, error)

// serveUploadStream feeds what the client sends over conn to upload, through a
// pipe, and answers one message: done, or the error. A copy that went well is
// done, whatever the client sent after its end - a tar's padding.
func serveUploadStream(ctx context.Context, conn *websocket.Conn, req *appcontainerdto.UploadFileToContainerReq,
	upload uploadFunc, progress bool, errInfo func(error) *hperrors.ErrorInfo,
) {
	pr, pw := io.Pipe()
	req.FileContent = pr
	if !req.Extract {
		req.FileContent = struct {
			io.Reader
			io.Closer
		}{&exactReader{r: pr, left: req.FileSize}, pr}
	}
	type outcome struct {
		resp *appcontainerdto.UploadFileToContainerResp
		err  error
	}
	copied := make(chan outcome, 1)
	safego.Go("appcontainer.uploadStream", func() {
		var out outcome
		defer func() {
			if r := recover(); r != nil {
				out.err = hperrors.Wrap(hperrors.ErrInternal).
					WithMsgLog("panic in the upload's copy: %v\n%s", r, debug.Stack())
			}
			// What the copy did not read stops the pump.
			_ = pr.CloseWithError(errors.Join(out.err, io.ErrClosedPipe))
			copied <- out
		}()
		out.resp, out.err = upload(ctx, req)
	})

	pumpErr := pumpUpload(conn, pw, transferIdle, progress)
	result := <-copied
	answer := streamControl{Type: streamDone}
	if result.resp != nil {
		answer.Data = result.resp.Data
	}
	if result.err != nil {
		err := result.err
		if byClient := clientError(pumpErr); byClient != nil {
			err = byClient
		}
		answer = streamControl{Type: streamError, Error: errInfo(err)}
	}
	if message, err := json.Marshal(answer); err == nil {
		_ = conn.WriteMessage(websocket.TextMessage, message)
	}
	_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		time.Now().Add(time.Second))
	drainUntilClosed(conn, answerDrain)
}

// drainUntilClosed reads, and drops, what the client sends until its close -
// for wait at most. One that left already ends it at once.
func drainUntilClosed(conn *websocket.Conn, wait time.Duration) {
	_ = conn.SetReadDeadline(time.Now().Add(wait))
	for {
		if _, _, err := conn.NextReader(); err != nil {
			return
		}
	}
}

// StreamFileToContainer Uploads a file or archive into container over a websocket
// @Summary Uploads a file or archive into container over a websocket
// @Description Takes what file-upload takes, as query parameters, then the content as binary messages
// @Description and {"type":"end"} as a text message; answers {"type":"done","data":{...}} or
// @Description {"type":"error","error":{...}}. Neither the server's nor Traefik's timeouts cut it; a
// @Description client sends a message at least every 60 seconds. With progress=true, each binary
// @Description message is followed by {"type":"progress","received":<bytes the copy has taken>}.
// @Description Without the websocket upgrade, the request is only checked: 204 when the stream would
// @Description be taken, else the error it would get - which a browser cannot read from a refused upgrade.
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
// @Param   overwrite query boolean false "extract only: an entry may replace a directory, or a file (default: true)"
// @Param   fileName query string false "the file's name"
// @Param   fileSize query integer false "the file's size in bytes: required unless extract"
// @Param   progress query boolean false "be told after each message how much the copy has taken"
// openapi:ignore-param file - the content comes over the websocket, not as a form's file
// @Success 101
// @Success 204
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
	progress, _ := strconv.ParseBool(ctx.Query("progress"))
	if !req.Extract {
		// The tar a single file goes in starts with its size.
		if req.FileSize, err = strconv.ParseInt(ctx.Query("fileSize"), 10, 64); err != nil || req.FileSize < 0 {
			h.RenderError(ctx, hperrors.NewArgumentInvalid("fileSize").
				WithMsgLog("fileSize is required for a file that is not extracted"))
			return
		}
	}
	if !h.IsWebsocketRequest(ctx) {
		ctx.Status(http.StatusNoContent)
		return
	}

	conn, err := h.UpgradeWebsocket(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	defer conn.Close()

	// A client that leaves reaches the copy through the pipe, closed with the
	// pump's error: a hijacked connection does not end the request's context.
	serveUploadStream(h.RequestCtx(ctx), conn, req,
		func(ctx context.Context, req *appcontainerdto.UploadFileToContainerReq) (
			*appcontainerdto.UploadFileToContainerResp, error,
		) {
			return h.appContainerUC.UploadFileToContainer(ctx, auth, req)
		}, progress,
		func(err error) *hperrors.ErrorInfo { return h.ErrorInfoOf(ctx, err) })
}
