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

	filters := make([]string, 0, len(req.Match)+2) //nolint:mnd // streams and contains
	for _, m := range req.Match {
		if m.Field == "" {
			return "", hperrors.Wrap(loggingmodel.ErrQueryScopeRequired)
		}
		filters = append(filters, strconv.Quote(m.Field)+":="+strconv.Quote(m.Value))
	}
	if len(req.Streams) > 0 {
		filters = append(filters, "stream:in("+quoteAll(req.Streams)+")")
	}
	if req.Contains != "" {
		filters = append(filters, "~"+strconv.Quote("(?i)"+regexp.QuoteMeta(req.Contains)))
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

func quoteAll(values []string) string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, strconv.Quote(v))
	}
	return strings.Join(out, ",")
}
