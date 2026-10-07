package handler

import (
	"bytes"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const jsonNULEscape = `\u0000`

// jsonHasNUL reports whether a JSON text holds a NUL character. A string carries
// one only as the escape \u0000 - a raw NUL is no JSON - and that is an escape
// only when the backslash before it is not itself escaped.
func jsonHasNUL(body []byte) bool {
	for i := 0; i < len(body); {
		j := bytes.Index(body[i:], []byte(jsonNULEscape))
		if j < 0 {
			return false
		}
		at, backslashes := i+j, 0
		for at-1-backslashes >= 0 && body[at-1-backslashes] == '\\' {
			backslashes++
		}
		if backslashes%2 == 0 {
			return true
		}
		i = at + len(jsonNULEscape)
	}
	return false
}

// queryHasNUL reports whether a query's names or values hold a NUL, %00 once
// decoded.
func queryHasNUL(query map[string]string) bool {
	for k, v := range query {
		if strings.IndexByte(k, 0) >= 0 || strings.IndexByte(v, 0) >= 0 {
			return true
		}
	}
	return false
}

// refuseNUL refuses a request that holds a NUL. A text column cannot store
// one, nor JSONB the \u0000 JSON writes for it: the statement fails, and the
// transaction it is in - a 500. A 400 that says so instead. A request that is
// basedto.NULAllowed goes on.
func refuseNUL(reqStruct any, hasNUL bool) error {
	if !hasNUL {
		return nil
	}
	if _, ok := reqStruct.(basedto.NULAllowed); ok {
		return nil
	}
	return hperrors.Wrap(hperrors.ErrRequestHasNUL)
}
