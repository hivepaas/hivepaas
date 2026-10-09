package appcontainerhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appcontaineruc/appcontainerdto"
)

// slowReader gives chunks of a byte each, every so often, then ends - or fails,
// with failAfter chunks given.
type slowReader struct {
	chunks, given, failAfter int
	every                    time.Duration
}

var errReadFailed = errors.New("the container's copy failed")

func (r *slowReader) Read(p []byte) (int, error) {
	if r.failAfter < 0 || (r.failAfter > 0 && r.given == r.failAfter) {
		return 0, errReadFailed
	}
	if r.given == r.chunks {
		return 0, io.EOF
	}
	time.Sleep(r.every)
	r.given++
	n := copy(p, bytes.Repeat([]byte{'x'}, 1024))
	return n, nil
}

func serveDownload(t *testing.T, writeTimeout time.Duration, r io.Reader) *httptest.Server {
	t.Helper()
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/x-tar")
		_ = copyDownload(w, r, time.Second)
	}))
	srv.Config.WriteTimeout = writeTimeout
	srv.Start()
	t.Cleanup(srv.Close)
	return srv
}

// A copy longer than the server's write timeout goes on while its bytes do.
func TestADownloadOutlastsTheWriteTimeout(t *testing.T) {
	srv := serveDownload(t, 150*time.Millisecond, &slowReader{chunks: 8, every: 50 * time.Millisecond})

	resp, err := http.Get(srv.URL) //nolint:noctx // a test's
	mustNot(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	mustNot(t, err)
	assert.Len(t, body, 8*1024, "the whole copy, though it took 400 ms of a 150 ms timeout")
}

// A copy cut half-way is read as a cut: a tar has no length to tell it by.
func TestADownloadCutIsNotAnEnd(t *testing.T) {
	srv := serveDownload(t, time.Minute, &slowReader{chunks: 8, failAfter: 2})

	resp, err := http.Get(srv.URL) //nolint:noctx // a test's
	mustNot(t, err)
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)

	assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
}

// streamServer pumps what a websocket client sends into a pipe, and keeps what
// came out of it and how the pump ended.
type streamServer struct {
	srv     *httptest.Server
	got     chan []byte
	pumped  chan error
	readErr error // the reader's failure, to stop the pump with
}

func newStreamServer(t *testing.T, idle time.Duration, readErr error) *streamServer {
	t.Helper()
	s := &streamServer{got: make(chan []byte, 1), pumped: make(chan error, 1), readErr: readErr}
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		pr, pw := io.Pipe()
		go func() {
			if s.readErr != nil {
				_ = pr.CloseWithError(s.readErr)
				s.got <- nil
				return
			}
			data, _ := io.ReadAll(pr)
			s.got <- data
		}()
		s.pumped <- pumpUpload(conn, pw, idle, false)
	}))
	t.Cleanup(s.srv.Close)
	return s
}

func (s *streamServer) dial(t *testing.T) *websocket.Conn {
	t.Helper()
	conn, resp, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.srv.URL, "http"), nil)
	mustNot(t, err)
	_ = resp.Body.Close()
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func TestAStreamEndsWithItsEndMessage(t *testing.T) {
	s := newStreamServer(t, time.Second, nil)
	conn := s.dial(t)

	for _, part := range []string{"FROM ", "scratch", "\n"} {
		mustNot(t, conn.WriteMessage(websocket.BinaryMessage, []byte(part)))
	}
	mustNot(t, conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"end"}`)))

	assert.NoError(t, <-s.pumped)
	assert.Equal(t, "FROM scratch\n", string(<-s.got), "whole, in order")
}

// A client that goes silent does not hold the copy: the pump ends, and the
// reader gets the error.
func TestASilentStreamIsCut(t *testing.T) {
	s := newStreamServer(t, 100*time.Millisecond, nil)
	conn := s.dial(t)
	mustNot(t, conn.WriteMessage(websocket.BinaryMessage, []byte("part")))

	select {
	case err := <-s.pumped:
		assert.Error(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("the pump waited on a silent client")
	}
}

// A copy that fails while the client still sends stops the pump with its error.
func TestAFailedCopyStopsTheStream(t *testing.T) {
	s := newStreamServer(t, time.Second, errReadFailed)
	conn := s.dial(t)
	<-s.got

	_ = conn.WriteMessage(websocket.BinaryMessage, []byte("part"))

	select {
	case err := <-s.pumped:
		assert.ErrorIs(t, err, errReadFailed)
	case <-time.After(5 * time.Second):
		t.Fatal("the pump went on after the copy failed")
	}
}

// Through gin, as the server serves it: gin refuses to hand over a connection
// it has written to, and the cut has to reach the client anyway.
func TestADownloadCutThroughGinIsNotAnEnd(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/x", func(ctx *gin.Context) {
		ctx.Header("Content-Type", "application/x-tar")
		_ = copyDownload(ctx.Writer, &slowReader{chunks: 8, failAfter: 2}, time.Second)
	})
	srv := httptest.NewServer(engine)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/x") //nolint:noctx // a test's
	mustNot(t, err)
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)

	assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
}

// A copy that fails before its first byte has sent nothing: its error can still
// be answered as one.
func TestADownloadThatFailsAtOnceIsAnswered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/x", func(ctx *gin.Context) {
		if err := copyDownload(ctx.Writer, &slowReader{chunks: 8, failAfter: -1}, time.Second); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		}
	})
	srv := httptest.NewServer(engine)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/x") //nolint:noctx // a test's
	mustNot(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// stream serves an upload stream with upload as the use case, and gives the
// client's connection and the errors the server answered.
func stream(t *testing.T, req *appcontainerdto.UploadFileToContainerReq, upload uploadFunc,
) (*websocket.Conn, chan *hperrors.ErrorInfo) {
	t.Helper()
	conn, answered, _ := streamWith(t, req, upload, false)
	return conn, answered
}

// streamWith is stream, the client asking for progress or not; served is
// closed once the server is done with the connection.
func streamWith(t *testing.T, req *appcontainerdto.UploadFileToContainerReq, upload uploadFunc, progress bool,
) (conn *websocket.Conn, answered chan *hperrors.ErrorInfo, served chan struct{}) {
	t.Helper()
	answered = make(chan *hperrors.ErrorInfo, 1)
	served = make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer close(served)
		defer conn.Close()
		serveUploadStream(r.Context(), conn, req, upload, progress, func(err error) *hperrors.ErrorInfo {
			info, _ := hperrors.ParseError(err, "")
			answered <- info
			return info
		})
	}))
	t.Cleanup(srv.Close)
	conn, resp, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	mustNot(t, err)
	_ = resp.Body.Close()
	t.Cleanup(func() { _ = conn.Close() })
	return conn, answered, served
}

// answerOf reads the server's last message.
func answerOf(t *testing.T, conn *websocket.Conn) streamControl {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		kind, message, err := conn.ReadMessage()
		mustNot(t, err, "the server answered nothing")
		if kind == websocket.TextMessage {
			var answer streamControl
			mustNot(t, json.Unmarshal(message, &answer))
			return answer
		}
	}
}

func readAll(_ context.Context, req *appcontainerdto.UploadFileToContainerReq) (
	*appcontainerdto.UploadFileToContainerResp, error,
) {
	if _, err := io.ReadAll(req.FileContent); err != nil {
		return nil, err
	}
	return &appcontainerdto.UploadFileToContainerResp{Data: &appcontainerdto.UploadFileToContainerDataResp{
		Path: "/app/x", Message: "ok"}}, nil
}

func send(t *testing.T, conn *websocket.Conn, parts ...string) {
	t.Helper()
	for _, part := range parts {
		mustNot(t, conn.WriteMessage(websocket.BinaryMessage, []byte(part)))
	}
	mustNot(t, conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"end"}`)))
}

// A single file is held to the size it declared, and a client that says one
// and sends another is told so - not left to a buffer of all it sends.
func TestAStreamedFileIsTheSizeItSaid(t *testing.T) {
	for name, parts := range map[string][]string{"more": {"12345", "678"}, "fewer": {"123"}} {
		conn, _ := stream(t, &appcontainerdto.UploadFileToContainerReq{FileSize: 5}, readAll)
		send(t, conn, parts...)

		answer := answerOf(t, conn)

		assert.Equal(t, streamError, answer.Type, name)
		if assert.NotNil(t, answer.Error, name) {
			assert.Equal(t, http.StatusBadRequest, answer.Error.Status, name)
		}
	}
	conn, _ := stream(t, &appcontainerdto.UploadFileToContainerReq{FileSize: 5}, readAll)
	send(t, conn, "12", "345")
	assert.Equal(t, streamDone, answerOf(t, conn).Type)
}

// An archive's copy ends at the archive's end: what the client sends after it -
// a tar's padding - is no failure of a copy that went well.
func TestAStreamAfterTheCopyEndedIsDone(t *testing.T) {
	conn, _ := stream(t, &appcontainerdto.UploadFileToContainerReq{Extract: true},
		func(_ context.Context, req *appcontainerdto.UploadFileToContainerReq) (
			*appcontainerdto.UploadFileToContainerResp, error,
		) {
			_, err := io.ReadFull(req.FileContent, make([]byte, 4))
			return &appcontainerdto.UploadFileToContainerResp{}, err
		})
	send(t, conn, "data", strings.Repeat("\x00", 1024))

	assert.Equal(t, streamDone, answerOf(t, conn).Type)
}

// A use case that panics is answered, not waited on for good.
func TestAStreamWhoseCopyPanicsIsAnswered(t *testing.T) {
	conn, _ := stream(t, &appcontainerdto.UploadFileToContainerReq{Extract: true},
		func(context.Context, *appcontainerdto.UploadFileToContainerReq) (
			*appcontainerdto.UploadFileToContainerResp, error,
		) {
			panic("a bug")
		})
	send(t, conn, "data")

	answer := answerOf(t, conn)
	assert.Equal(t, streamError, answer.Type)
}

// A client that goes before its end message is the client's doing: a 400 to
// keep, not a 500 of the server's.
func TestAStreamTheClientLeftIsItsError(t *testing.T) {
	conn, answered := stream(t, &appcontainerdto.UploadFileToContainerReq{Extract: true}, readAll)
	mustNot(t, conn.WriteMessage(websocket.BinaryMessage, []byte("part")))
	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	select {
	case info := <-answered:
		assert.Equal(t, http.StatusBadRequest, info.Status)
	case <-time.After(5 * time.Second):
		t.Fatal("no answer")
	}
}

// A client that asks for progress is told, after each message, how much of the
// upload the copy has taken: what it may send next is bounded by that, with no
// timer of its own - a browser's, in a tab in the background, wakes once a
// minute.
func TestAStreamAskedForProgressTellsWhatTheCopyTook(t *testing.T) {
	conn, _, _ := streamWith(t, &appcontainerdto.UploadFileToContainerReq{FileSize: 5}, readAll, true)

	var told []string
	for _, part := range []string{"12", "345"} {
		mustNot(t, conn.WriteMessage(websocket.BinaryMessage, []byte(part)))
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, message, err := conn.ReadMessage()
		mustNot(t, err, "no progress after a message")
		told = append(told, strings.TrimSpace(string(message)))
	}
	mustNot(t, conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"end"}`)))

	assert.Equal(t, []string{`{"type":"progress","received":2}`, `{"type":"progress","received":5}`}, told)
	assert.Equal(t, streamDone, answerOf(t, conn).Type)
}

// A client not asking for progress is told nothing until the answer: one that
// reads only then would fill its buffer with what it does not read.
func TestAStreamNotAskedForProgressTellsOnlyTheAnswer(t *testing.T) {
	conn, _ := stream(t, &appcontainerdto.UploadFileToContainerReq{FileSize: 5}, readAll)
	send(t, conn, "12", "345")

	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, message, err := conn.ReadMessage()
	mustNot(t, err)
	assert.Contains(t, string(message), `"type":"done"`)
}

// After its answer the server waits for the client's close, reading what the
// client still sends: a connection closed with bytes unread is reset, and a
// reset can cost the client the answer that came before it.
func TestAStreamAnsweredWaitsForTheClientsClose(t *testing.T) {
	conn, _, served := streamWith(t, &appcontainerdto.UploadFileToContainerReq{Extract: true},
		func(context.Context, *appcontainerdto.UploadFileToContainerReq) (
			*appcontainerdto.UploadFileToContainerResp, error,
		) {
			return nil, hperrors.NewArgumentInvalid("path")
		}, false)
	mustNot(t, conn.WriteMessage(websocket.BinaryMessage, []byte("part")))
	assert.Equal(t, streamError, answerOf(t, conn).Type)

	// The client was still sending.
	_ = conn.WriteMessage(websocket.BinaryMessage, []byte("more"))
	select {
	case <-served:
		t.Fatal("the server closed the connection before the client did")
	case <-time.After(300 * time.Millisecond):
	}

	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	select {
	case <-served:
	case <-time.After(5 * time.Second):
		t.Fatal("the server waited on after the client closed")
	}
}

// mustNot stops the test on err: testify's require is not vendored here.
func mustNot(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	if err != nil {
		t.Fatal(append([]any{err}, msgAndArgs...)...)
	}
}
