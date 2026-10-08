package appcontainerhandler

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// slowReader gives chunks of a byte each, every so often, then ends - or fails,
// with failAfter chunks given.
type slowReader struct {
	chunks, given, failAfter int
	every                    time.Duration
}

var errReadFailed = errors.New("the container's copy failed")

func (r *slowReader) Read(p []byte) (int, error) {
	if r.failAfter > 0 && r.given == r.failAfter {
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
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	require.NoError(t, err)
	assert.Len(t, body, 8*1024, "the whole copy, though it took 400 ms of a 150 ms timeout")
}

// A copy cut half-way is read as a cut: a tar has no length to tell it by.
func TestADownloadCutIsNotAnEnd(t *testing.T) {
	srv := serveDownload(t, time.Minute, &slowReader{chunks: 8, failAfter: 2})

	resp, err := http.Get(srv.URL) //nolint:noctx // a test's
	require.NoError(t, err)
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
		s.pumped <- pumpUpload(conn, pw, idle)
	}))
	t.Cleanup(s.srv.Close)
	return s
}

func (s *streamServer) dial(t *testing.T) *websocket.Conn {
	t.Helper()
	conn, resp, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.srv.URL, "http"), nil)
	require.NoError(t, err)
	_ = resp.Body.Close()
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func TestAStreamEndsWithItsEndMessage(t *testing.T) {
	s := newStreamServer(t, time.Second, nil)
	conn := s.dial(t)

	for _, part := range []string{"FROM ", "scratch", "\n"} {
		require.NoError(t, conn.WriteMessage(websocket.BinaryMessage, []byte(part)))
	}
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"end"}`)))

	assert.NoError(t, <-s.pumped)
	assert.Equal(t, "FROM scratch\n", string(<-s.got), "whole, in order")
}

// A client that goes silent does not hold the copy: the pump ends, and the
// reader gets the error.
func TestASilentStreamIsCut(t *testing.T) {
	s := newStreamServer(t, 100*time.Millisecond, nil)
	conn := s.dial(t)
	require.NoError(t, conn.WriteMessage(websocket.BinaryMessage, []byte("part")))

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
