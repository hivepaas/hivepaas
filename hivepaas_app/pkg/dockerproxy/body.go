package dockerproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// maxBody bounds what the proxy reads to judge a request. A container create is
// a few kilobytes; a body near this size is not one.
const maxBody = 1 << 20

// readBody decodes a request's JSON object. Numbers stay json.Number, so that
// what is passed on is what was sent.
func readBody(r *http.Request) (map[string]any, error) {
	raw, err := io.ReadAll(io.LimitReader(r.Body, maxBody+1))
	if err != nil {
		return nil, fmt.Errorf("reading the request: %w", err)
	}
	if len(raw) > maxBody {
		return nil, refusef("the request body is larger than %d bytes", maxBody)
	}
	var body map[string]any
	if len(bytes.TrimSpace(raw)) > 0 {
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		if err = decoder.Decode(&body); err != nil {
			return nil, refusef("the request body is not a JSON object")
		}
	}
	if body == nil {
		body = map[string]any{}
	}
	return body, nil
}

// formValues are a request's parameters as the daemon reads them: the query,
// and for a form-encoded body the body too, whose values come first. Judging the
// query alone would let a client put one image there and pull another from the
// body - utopia-php's Docker client, which Appwrite's executor uses, sends every
// parameter of a pull in the body. The body is left for the request to forward.
func formValues(r *http.Request) (url.Values, error) {
	raw, err := io.ReadAll(io.LimitReader(r.Body, maxBody+1))
	if err != nil {
		return nil, fmt.Errorf("reading the request: %w", err)
	}
	if len(raw) > maxBody {
		return nil, refusef("the request body is larger than %d bytes", maxBody)
	}
	r.Body = io.NopCloser(bytes.NewReader(raw))
	parsed := r.Clone(r.Context())
	parsed.Body = io.NopCloser(bytes.NewReader(raw))
	if err = parsed.ParseForm(); err != nil {
		return nil, refusef("the request's parameters are not readable")
	}
	return parsed.Form, nil
}

// setBody replaces the request's body with body, as JSON.
func setBody(r *http.Request, body map[string]any) {
	// A map of what readBody decoded, and of the strings, numbers and maps the
	// proxy put in it, always marshals.
	raw, _ := json.Marshal(body)
	r.Body = io.NopCloser(bytes.NewReader(raw))
	r.ContentLength = int64(len(raw))
	r.TransferEncoding = nil
	r.Header.Set("Content-Length", strconv.Itoa(len(raw)))
	r.Header.Set("Content-Type", contentTypeJSON)
}

// object is v as a JSON object, or nil when it is not one.
func object(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

// text is v as a string, or "" when it is not one.
func text(v any) string {
	s, _ := v.(string)
	return s
}

// list is v as a JSON array, or nil when it is not one.
func list(v any) []any {
	l, _ := v.([]any)
	return l
}

// number reads a whole number the client sent, or one the proxy set. Absent and
// null are zero.
func number(v any) (int64, error) {
	switch n := v.(type) {
	case nil:
		return 0, nil
	case int64:
		return n, nil
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return 0, refusef("%s is not a whole number", n)
		}
		return i, nil
	default:
		return 0, refusef("%v is not a number", v)
	}
}
