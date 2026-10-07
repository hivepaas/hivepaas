package entity

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// marshalJSONB is v as JSON a JSONB column takes: Postgres refuses a string
// holding \u0000 - the statement fails, and the transaction it is in - and text
// from outside can hold one: what a process printed, what a snapshot or a
// docker label says, a YAML "\0". Such characters are dropped, and clean is
// false: what is stored is no longer v, so v is not to be kept as what it parses
// to. Numbers stay exact. A task's args and output and a setting's data go
// through it.
func marshalJSONB(v any) (b []byte, clean bool, err error) {
	if b, err = json.Marshal(v); err != nil || !bytes.Contains(b, []byte(`\u0000`)) {
		return b, true, hperrors.Wrap(err)
	}
	// The escape is also the text of a string holding a backslash before
	// "u0000": decoded, the strings say which.
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var decoded any
	if err = dec.Decode(&decoded); err != nil {
		return nil, false, hperrors.Wrap(err)
	}
	stripped, changed := withoutNUL(decoded)
	if !changed {
		return b, true, nil
	}
	b, err = json.Marshal(stripped)
	return b, false, hperrors.Wrap(err)
}

// withoutNUL is a decoded JSON value with the NUL characters of its strings
// and keys dropped, and whether there were any.
func withoutNUL(v any) (any, bool) {
	switch x := v.(type) {
	case string:
		if !strings.Contains(x, "\x00") {
			return x, false
		}
		return strings.ReplaceAll(x, "\x00", ""), true
	case []any:
		changed := false
		for i, e := range x {
			var c bool
			x[i], c = withoutNUL(e)
			changed = changed || c
		}
		return x, changed
	case map[string]any:
		changed := false
		out := make(map[string]any, len(x))
		for k, e := range x {
			key := strings.ReplaceAll(k, "\x00", "")
			value, cv := withoutNUL(e)
			out[key] = value
			changed = changed || cv || key != k
		}
		return out, changed
	}
	return v, false
}
