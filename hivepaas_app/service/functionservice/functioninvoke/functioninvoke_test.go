package functioninvoke

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestAJobsRequestIsWhatInvokeReads(t *testing.T) {
	req := RequestOf(&entity.SchedJobFunctionInvoke{
		Method:  "POST",
		Path:    "/report?day=1&day=2&empty=",
		Headers: map[string][]string{"content-type": {"text/plain"}},
		Body:    "hi",
	})

	assert.Equal(t, &Request{
		Method:  "POST",
		Path:    "/report",
		Query:   map[string][]string{"day": {"1", "2"}, "empty": {""}},
		Headers: map[string][]string{"content-type": {"text/plain"}},
		Body:    []byte("hi"),
	}, req)
	raw, err := json.Marshal(req)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"method":"POST","path":"/report","query":{"day":["1","2"],"empty":[""]},`+
		`"headers":{"content-type":["text/plain"]},"body":"aGk="}`, string(raw), "the body is base64, as invoke reads it")

	assert.Equal(t, &Request{Method: "GET", Path: "/"}, RequestOf(&entity.SchedJobFunctionInvoke{Method: "GET",
		Path: "/"}))
}

func TestAResultLineIsTheResponse(t *testing.T) {
	result, err := ParseResult(`{"status":201,"headers":{"x-request-id":["r1"]},"body":"b2s=",` +
		`"requestId":"r1","durationMs":3.5,"outcome":"ok"}`)

	assert.NoError(t, err)
	assert.Equal(t, &Result{Status: 201, Headers: map[string][]string{"x-request-id": {"r1"}}, Body: []byte("ok"),
		RequestID: "r1", DurationMs: 3.5, Outcome: "ok"}, result)

	_, err = ParseResult("not json")
	assert.Error(t, err)
}

// newWriter is an output writer whose log goes to logged.
func newWriter() (w *OutputWriter, logged *strings.Builder) {
	logged = &strings.Builder{}
	return NewOutputWriter(func(line []byte) { logged.Write(line) }), logged
}

// collect writes chunks into an output writer, and returns the log it passed on.
func collect(w *OutputWriter, logged *strings.Builder, chunks ...string) string {
	for _, chunk := range chunks {
		_, _ = w.Write([]byte(chunk))
	}
	w.Flush()
	return logged.String()
}

func TestInvokesOutputIsTheCallsLogThenItsResult(t *testing.T) {
	w, sink := newWriter()
	logged := collect(w, sink, "GET /re", "port\nsecond line\n#hivepaas-res", `ult {"status":200,"outcome":"ok"}`, "\n")

	assert.Equal(t, "GET /report\nsecond line\n", logged)
	if assert.NotNil(t, w.Result()) {
		assert.Equal(t, `{"status":200,"outcome":"ok"}`, string(w.Result()))
	}
	assert.False(t, w.ResultTooLarge())
}

func TestAResultWithoutItsNewlineIsStillTheResult(t *testing.T) {
	w, sink := newWriter()
	collect(w, sink, "log\n#hivepaas-result {}")

	assert.Equal(t, "{}", string(w.Result()))
}

func TestOnlyALineThatStartsWithTheMarkerIsTheResult(t *testing.T) {
	w, sink := newWriter()
	logged := collect(w, sink, "said #hivepaas-result {}\n")

	assert.Nil(t, w.Result())
	assert.Equal(t, "said #hivepaas-result {}\n", logged)
}

func TestALongLogLineIsPassedOnInPieces(t *testing.T) {
	w, sink := newWriter()
	long := strings.Repeat("a", lineMax*2+10)
	logged := collect(w, sink, long, "\n")

	assert.Equal(t, long+"\n", logged, "nothing of it is lost")
	assert.Nil(t, w.Result())
}

func TestAResultTooLargeToReadIsSaid(t *testing.T) {
	w, sink := newWriter()
	w.resultMax = 20
	collect(w, sink, "#hivepaas-result {\"body\":\""+strings.Repeat("a", 30)+"\"}\n")

	assert.Nil(t, w.Result())
	assert.True(t, w.ResultTooLarge())
}

// A result too large arrives in pieces, as a pipe gives it: none of it is kept,
// as the result or as the log.
func TestAResultTooLargeInPiecesIsSaid(t *testing.T) {
	w, sink := newWriter()
	w.resultMax = 20
	logged := collect(w, sink, "log\n#hivepaas-result {\"body\":\"", strings.Repeat("a", 30), "aaaa", "\"}\n")

	assert.Equal(t, "log\n", logged)
	assert.Nil(t, w.Result())
	assert.True(t, w.ResultTooLarge())
}
