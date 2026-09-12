package victorialogs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

const appField = "attrs.hivepaas.app.id"

func scoped(mod func(r *loggingmodel.QueryReq)) *loggingmodel.QueryReq {
	r := &loggingmodel.QueryReq{
		Match: []loggingmodel.FieldMatch{{Field: appField, Value: "APP1"}},
		Limit: 100,
	}
	if mod != nil {
		mod(r)
	}
	return r
}

const tail = ` | unpack_json from _msg fields (level) result_prefix "app."` +
	` | sort by (_time desc) | limit 100 | fields _time, _msg, stream, app.level`

func TestBuildQueryScopeOnly(t *testing.T) {
	q, err := BuildQuery(scoped(nil))
	assert.NoError(t, err)
	assert.Equal(t, `"attrs.hivepaas.app.id":="APP1"`+tail, q)
}

func TestBuildQueryRefusesAnUnscopedRequest(t *testing.T) {
	_, err := BuildQuery(&loggingmodel.QueryReq{Limit: 10})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryScopeRequired)

	_, err = BuildQuery(&loggingmodel.QueryReq{Limit: 10, Match: []loggingmodel.FieldMatch{{Value: "x"}}})
	assert.ErrorIs(t, err, loggingmodel.ErrQueryScopeRequired, "a match with no field scopes nothing")
}

func TestBuildQueryRefusesALimitOutOfRange(t *testing.T) {
	for _, limit := range []int{0, -1, loggingmodel.MaxQueryLimit + 1} {
		_, err := BuildQuery(scoped(func(r *loggingmodel.QueryReq) { r.Limit = limit }))
		assert.ErrorIs(t, err, loggingmodel.ErrQueryInvalid, "limit %d", limit)
	}
}

func TestBuildQueryEveryParameter(t *testing.T) {
	q, err := BuildQuery(scoped(func(r *loggingmodel.QueryReq) {
		r.Contains = "Timeout (x+y)"
		r.Streams = []string{"stderr"}
		r.Levels = []string{"error", "warn"}
	}))
	assert.NoError(t, err)
	assert.Equal(t,
		`"attrs.hivepaas.app.id":="APP1" AND stream:in("stderr") AND ~"(?i)Timeout \\(x\\+y\\)"`+
			` | unpack_json from _msg fields (level) result_prefix "app."`+
			` | filter "app.level":~"(?i)^(error|warn)$"`+
			` | sort by (_time desc) | limit 100 | fields _time, _msg, stream, app.level`,
		q)
}

// Each of these tries to break out of the literal it is placed in. Built with
// Go quoting, each stays a single literal - verified against VictoriaLogs
// v1.52.0, where the first returns no line of APP2.
func TestBuildQueryKeepsHostileInputInsideItsLiteral(t *testing.T) {
	hostile := []string{
		`x" OR "attrs.hivepaas.app.id":="APP2`,
		`") OR ("attrs.hivepaas.app.id":="APP2`,
		`x | delete attrs.hivepaas.app.id`,
		"back\\slash \" and \n newline",
		`"""`,
	}
	for _, h := range hostile {
		q, err := BuildQuery(scoped(func(r *loggingmodel.QueryReq) { r.Contains = h }))
		assert.NoError(t, err)
		assert.Contains(t, q, `"attrs.hivepaas.app.id":="APP1" AND ~`, h)
		// The scope and the pipes after it are untouched: nothing the input
		// held became syntax.
		assert.Equal(t, 4, countOutsideLiterals(q, "|"), h) // unpack_json, sort, limit, fields
		assert.NotContains(t, stripLiterals(q), "APP2", h)
	}

	for _, h := range hostile {
		q, err := BuildQuery(scoped(func(r *loggingmodel.QueryReq) { r.Match[0].Value = h }))
		assert.NoError(t, err)
		assert.NotContains(t, stripLiterals(q), "APP2", h)
	}
	for _, h := range hostile {
		q, err := BuildQuery(scoped(func(r *loggingmodel.QueryReq) { r.Levels = []string{h} }))
		assert.NoError(t, err)
		assert.NotContains(t, stripLiterals(q), "APP2", h)
	}
}

func TestBuildQueryAlwaysPrefixesUnpackedFields(t *testing.T) {
	// Spec section 7: an unprefixed unpack_json lets a line's own JSON
	// overwrite the daemon's attrs.hivepaas.app.id at query time.
	for _, r := range []*loggingmodel.QueryReq{
		scoped(nil),
		scoped(func(r *loggingmodel.QueryReq) { r.Levels = []string{"error"} }),
	} {
		q, err := BuildQuery(r)
		assert.NoError(t, err)
		assert.Contains(t, q, `unpack_json from _msg fields (level) result_prefix "app."`)
		assert.Equal(t, 1, countOutsideLiterals(q, "unpack_json"))
	}
}

// stripLiterals removes every double-quoted literal, honoring backslash
// escapes, leaving only what VictoriaLogs parses as syntax.
func stripLiterals(q string) string {
	out := make([]rune, 0, len(q))
	in, esc := false, false
	for _, c := range q {
		switch {
		case esc:
			esc = false
		case in && c == '\\':
			esc = true
		case c == '"':
			in = !in
		case !in:
			out = append(out, c)
		}
	}
	return string(out)
}

func countOutsideLiterals(q, sub string) int {
	s := stripLiterals(q)
	n := 0
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			n++
		}
	}
	return n
}
