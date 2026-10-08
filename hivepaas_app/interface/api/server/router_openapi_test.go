package server

import (
	"encoding/json"
	"maps"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// templatedOperations are operations of the spec that stand for several routes:
// the routes a handler factory registers once per group, which have no comment
// of their own. The parameter in braces is what differs between them.
var templatedOperations = map[string]string{
	"GET /api/settings/{kind}/{itemID}/usages": "kind",
}

var specParam = regexp.MustCompile(`\{([^}]+)\}`)

// specOperations are the operations of docs/openapi/swagger.json, as gin
// writes routes: GET /api/projects/:projectID.
func specOperations(t *testing.T) map[string]bool {
	t.Helper()
	data, err := os.ReadFile("../../../../docs/openapi/swagger.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err = json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	ops := map[string]bool{}
	for path, methods := range doc.Paths {
		for method := range methods {
			switch method {
			case "get", "post", "put", "patch", "delete", "head", "options":
				ops[strings.ToUpper(method)+" /api"+path] = true
			}
		}
	}
	return ops
}

// templateMatches is the routes a templated operation stands for.
func templateMatches(op, param string, routes map[string]bool) []string {
	segments := strings.Split(op, "/")
	for i, segment := range segments {
		m := specParam.FindStringSubmatch(segment)
		switch {
		case m == nil || m[0] != segment:
			segments[i] = regexp.QuoteMeta(segment)
		case m[1] == param:
			segments[i] = `[a-z][a-z0-9-]*`
		default:
			segments[i] = ":" + m[1]
		}
	}
	re := regexp.MustCompile("^" + strings.Join(segments, "/") + "$")
	var matched []string
	for route := range routes {
		if re.MatchString(route) {
			matched = append(matched, route)
		}
	}
	return matched
}

// Every route the API registers is an operation in the spec, and every
// operation is a route. A client generated from the spec - the CLI - calls what
// is there: a route missing from it cannot be called, and an operation with a
// path no route has fails every time.
func TestTheOpenAPISpecHasEveryRouteAndNothingElse(t *testing.T) {
	routes := map[string]bool{}
	for _, route := range allRoutes(t) {
		routes[route.Method+" "+route.Path] = true
	}
	ops := map[string]bool{}
	for op := range specOperations(t) {
		ops[specParam.ReplaceAllString(op, ":$1")] = true
	}

	covered := map[string]bool{}
	for op, param := range templatedOperations {
		matched := templateMatches(op, param, routes)
		assert.NotEmpty(t, matched, "the templated operation %s stands for no route", op)
		for _, route := range matched {
			covered[route] = true
		}
		delete(ops, specParam.ReplaceAllString(op, ":$1"))
	}

	var missing, extra []string
	for route := range routes {
		if !ops[route] && !covered[route] {
			missing = append(missing, route)
		}
	}
	for op := range ops {
		if !routes[op] {
			extra = append(extra, op)
		}
	}
	slices.Sort(missing)
	slices.Sort(extra)
	assert.Empty(t, missing, "routes with no operation in docs/openapi/swagger.json: add their swag comments, "+
		"then run make gen-swag")
	assert.Empty(t, extra, "operations no route serves: their @Router does not match the route")
	assert.NotEmpty(t, slices.Collect(maps.Keys(routes)))
}

// Every operation has an id, and no two the same: a client generated from the
// spec names each method after it, and a repeated id is a method declared
// twice, which does not compile.
func TestTheOpenAPISpecNamesEveryOperationOnce(t *testing.T) {
	data, err := os.ReadFile("../../../../docs/openapi/swagger.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Paths map[string]map[string]struct {
			OperationID string `json:"operationId"`
		} `json:"paths"`
	}
	if err = json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	byID := map[string][]string{}
	var unnamed []string
	for path, methods := range doc.Paths {
		for method, op := range methods {
			if op.OperationID == "" {
				unnamed = append(unnamed, strings.ToUpper(method)+" "+path)
				continue
			}
			byID[op.OperationID] = append(byID[op.OperationID], strings.ToUpper(method)+" "+path)
		}
	}
	var repeated []string
	for id, ops := range byID {
		if len(ops) > 1 {
			slices.Sort(ops)
			repeated = append(repeated, id+": "+strings.Join(ops, ", "))
		}
	}
	slices.Sort(unnamed)
	slices.Sort(repeated)
	assert.Empty(t, unnamed, "operations without an @Id")
	assert.Empty(t, repeated, "@Ids naming more than one operation")
}
