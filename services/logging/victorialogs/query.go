package victorialogs

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

const (
	// UnpackPrefix is put in front of every field read out of a message.
	//
	// Without it, a line printing {"attrs.hivepaas.app.id":"OTHER"} would
	// overwrite the daemon's own field of that name for the rest of the query
	// - measured, not assumed. Nothing may unpack without it.
	UnpackPrefix = "app."

	// LevelField is the message's own level once unpacked.
	LevelField = UnpackPrefix + "level"
)

// BuildQuery turns a structured request into LogsQL.
//
// It is the only place HivePaaS writes LogsQL. Every value from the request is
// placed inside a Go-quoted literal, which LogsQL parses the same way, so no
// value can become syntax; the structure around them is fixed here.
func BuildQuery(req *loggingmodel.QueryReq) (string, error) {
	if len(req.Match) == 0 {
		return "", hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
	}
	if req.Limit <= 0 || req.Limit > loggingmodel.MaxQueryLimit {
		return "", hperrors.Wrap(loggingmodel.ErrQueryInvalid).
			WithExtraDetail("limit must be between 1 and %d", loggingmodel.MaxQueryLimit)
	}

	filters := make([]string, 0, len(req.Match)+3) //nolint:mnd // streams, search, level prefilter
	for _, m := range req.Match {
		if m.Field == "" {
			return "", hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
		}
		filters = append(filters, strconv.Quote(m.Field)+":="+strconv.Quote(m.Value))
	}
	if len(req.Streams) > 0 {
		filters = append(filters, "stream:in("+quoteAll(req.Streams)+")")
	}
	if !req.Search.IsEmpty() {
		filters = append(filters, textSearchFilter(req.Search))
	}
	if len(req.Levels) > 0 {
		filters = append(filters, levelPrefilter(req.Levels))
	}

	var b strings.Builder
	b.WriteString(strings.Join(filters, " AND "))
	b.WriteString(" | unpack_json from _msg fields (level) result_prefix ")
	b.WriteString(strconv.Quote(UnpackPrefix))
	if len(req.Levels) > 0 {
		quoted := make([]string, 0, len(req.Levels))
		for _, l := range req.Levels {
			quoted = append(quoted, regexp.QuoteMeta(l))
		}
		b.WriteString(" | filter ")
		b.WriteString(strconv.Quote(LevelField))
		b.WriteString(":~")
		b.WriteString(strconv.Quote("(?i)^(" + strings.Join(quoted, "|") + ")$"))
	}
	fmt.Fprintf(&b, " | sort by (_time desc) | limit %d | fields _time, _msg, stream, %s", req.Limit, LevelField)
	return b.String(), nil
}

// textSearchFilter renders the free-text filter in the mode the caller asked
// for.
//
// Every mode puts the value inside a Go-quoted literal, which LogsQL parses the
// same way, so no input becomes syntax in any of them. What the mode changes is
// only how the backend reads the string it is handed.
//
// Plain text gets a trailing `*` so that it matches from the start of a token:
// measured at the same cost as an exact token match, and it finds "timeouts"
// for someone who typed "timeout". What it does not find is a match inside a
// token - "connection_timeout" - because `_` does not end a token. That is what
// the regular expression mode is for.
func textSearchFilter(s *loggingmodel.TextSearch) string {
	if s == nil {
		return ""
	}
	if s.IsRegex {
		if s.CaseSensitive {
			return "~" + strconv.Quote(s.Value)
		}
		return "~" + strconv.Quote("(?i)"+s.Value)
	}
	if s.CaseSensitive {
		return strconv.Quote(s.Value) + "*"
	}
	return "i(" + strconv.Quote(s.Value) + "*)"
}

// levelPrefilter narrows to the lines that could carry one of these levels,
// before anything is parsed.
//
// It cannot drop a line the level filter downstream would have kept: a message
// whose JSON level is "error" contains "error" as text. It does let through
// lines that merely mention the word, and that filter removes them - so the
// result is unchanged and only the work differs. Measured over 400k lines, the
// case where nothing matches goes from 133ms to 54ms.
func levelPrefilter(levels []string) string {
	terms := make([]string, 0, len(levels))
	for _, l := range levels {
		terms = append(terms, "i("+strconv.Quote(l)+")")
	}
	return "(" + strings.Join(terms, " OR ") + ")"
}

func quoteAll(values []string) string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, strconv.Quote(v))
	}
	return strings.Join(out, ",")
}
