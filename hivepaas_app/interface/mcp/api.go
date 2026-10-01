package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

// A tool that reads is one GET endpoint of the API: it takes the endpoint's
// query parameters, read from the fields its request type decodes the query
// into, and answers what the endpoint answers, decoded into its response type.
// Nothing about the endpoint is declared twice; what a tool adds is its
// description, and the names a model can give instead of ids.
// See docs/superpowers/specs/2026-09-26-mcp-tools-follow-the-api-design.md.

// typeSchemas are the API's types whose JSON is not what their Go kind says:
// a size or a duration is a string, not the integer it is stored as.
var typeSchemas = map[reflect.Type]*jsonschema.Schema{
	reflect.TypeFor[time.Time]():         {Type: typeString, Description: "a time in RFC 3339: 2026-09-26T08:00:00Z"},
	reflect.TypeFor[timeutil.Date]():     {Type: typeString, Description: "a date, YYYY-MM-DD"},
	reflect.TypeFor[timeutil.Duration](): {Type: typeString, Description: "a duration such as 30s, 5m, 2h or 1d"},
	reflect.TypeFor[unit.DataSize]():     {Type: typeString, Description: "a size such as 512MB or 1GB"},
	reflect.TypeFor[json.RawMessage]():   {},
}

// schemaOf is the JSON schema of one of the API's types.
func schemaOf(t reflect.Type) (*jsonschema.Schema, error) {
	s, err := jsonschema.ForType(t, &jsonschema.ForOptions{TypeSchemas: typeSchemas})
	if err != nil {
		return nil, fmt.Errorf("mcp: the schema of %v: %w", t, err)
	}
	return s, nil
}

// Names many tools share: of arguments, query parameters and schema types.
const (
	argProject    = "project"
	argEnv        = "env"
	argApp        = "app"
	argTemplate   = "template"
	argName       = "name"
	argKind       = "kind"
	paramSearch   = "search"
	paramFromDate = "fromDate"
	paramDomain   = "domain"
	paramToDate   = "toDate"
	paramStatus   = "status"
	paramGetStats = "getStats"
	typeString    = "string"
	typeInteger   = "integer"
	// redactedValue stands in the audit log for a value that may be secret.
	redactedValue = "(given)"
)

// ---- query parameters ----

// queryParam is one query parameter of an endpoint: a field of its request
// type with a mapstructure name, or a paging parameter.
type queryParam struct {
	name   string
	schema *jsonschema.Schema
}

const (
	paramPageOffset = "pageOffset"
	paramSort       = "sort"
)

// queryParamsOf reads an endpoint's query parameters from its request type:
// the fields the handler decodes the query into, by their mapstructure names,
// and pageOffset, pageLimit and sort when the request pages.
func queryParamsOf(req any) ([]queryParam, error) {
	if req == nil {
		return nil, nil
	}
	var params []queryParam
	pages := false
	var walk func(t reflect.Type) error
	walk = func(t reflect.Type) error {
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		for i := range t.NumField() {
			f := t.Field(i)
			if f.Type == reflect.TypeFor[basedto.Paging]() {
				pages = true
				continue
			}
			tag, _, _ := strings.Cut(f.Tag.Get("mapstructure"), ",")
			if tag == "" && f.Anonymous {
				if err := walk(f.Type); err != nil {
					return err
				}
				continue
			}
			if tag == "" || tag == "-" {
				continue
			}
			s, err := querySchemaOf(f.Type)
			if err != nil {
				return fmt.Errorf("mcp: query parameter %s: %w", tag, err)
			}
			params = append(params, queryParam{name: tag, schema: s})
		}
		return nil
	}
	if err := walk(reflect.TypeOf(req)); err != nil {
		return nil, err
	}
	if pages {
		params = append(params,
			queryParam{name: paramPageOffset, schema: &jsonschema.Schema{Type: typeInteger, Minimum: new(float64)}},
			queryParam{name: paramPageLimit, schema: &jsonschema.Schema{Type: typeInteger,
				Minimum: ptr(1.0), Maximum: ptr(float64(basedto.PageLimitMax))}},
			queryParam{name: paramSort, schema: &jsonschema.Schema{Type: typeString}},
		)
	}
	return params, nil
}

func ptr[T any](v T) *T { return &v }

var (
	errNotQueryable = errors.New("not what a query carries")
	errNoSuchField  = errors.New("mcp: a description of no field")
)

// querySchemaOf is the schema of a query parameter of a type: what a model
// gives, which the query carries as text.
func querySchemaOf(t reflect.Type) (*jsonschema.Schema, error) {
	if s, ok := typeSchemas[t]; ok {
		return s.CloneSchemas(), nil
	}
	switch t.Kind() { //nolint:exhaustive // the kinds a query decodes into
	case reflect.String:
		return &jsonschema.Schema{Type: typeString}, nil
	case reflect.Bool:
		return &jsonschema.Schema{Type: "boolean"}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &jsonschema.Schema{Type: typeInteger}, nil
	case reflect.Float32, reflect.Float64:
		return &jsonschema.Schema{Type: "number"}, nil
	case reflect.Slice:
		items, err := querySchemaOf(t.Elem())
		if err != nil {
			return nil, err
		}
		return &jsonschema.Schema{Type: "array", Items: items}, nil
	case reflect.Pointer:
		return querySchemaOf(t.Elem())
	}
	return nil, fmt.Errorf("%w: %v", errNotQueryable, t)
}

// toQuery writes a tool's arguments as the endpoint's query: a list as its
// values joined by commas, which is how the handler reads one.
func toQuery(params []queryParam, args map[string]any) (url.Values, error) {
	query := url.Values{}
	for _, p := range params {
		v, ok := args[p.name]
		if !ok || v == nil {
			continue
		}
		text, err := queryText(v)
		if err != nil {
			return nil, &InputError{Message: fmt.Sprintf("%s: %v", p.name, err)}
		}
		query.Set(p.name, text)
	}
	return query, nil
}

func queryText(v any) (string, error) {
	switch v := v.(type) {
	case string:
		return v, nil
	case bool:
		return strconv.FormatBool(v), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case json.Number:
		return v.String(), nil
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			text, err := queryText(item)
			if err != nil {
				return "", err
			}
			parts = append(parts, text)
		}
		return strings.Join(parts, ","), nil
	}
	return "", fmt.Errorf("%w: %v", errNotQueryable, v)
}

// ---- where an endpoint is ----

// under is what an endpoint's path is under, and so which names a tool takes
// to find it.
type under int

const (
	underNothing under = iota
	underProject
	underEnv
	underApp
)

const descProjectArg = "the project's key, name or id; list_projects lists them"

// scopeArgs are the tool arguments that name what an endpoint is under.
var scopeArgs = map[under][]scopeArg{
	underProject: {
		{argProject, descProjectArg},
	},
	underEnv: {
		{argProject, descProjectArg},
		{argEnv, "the env's name, such as production; list_projects lists each project's envs"},
	},
	underApp: {
		{argProject, descProjectArg},
		{argEnv, "the env's name, such as production; list_projects lists each project's envs"},
		{argApp, "the app's key, name or id; list_apps lists them"},
	},
}

type scopeArg struct{ name, description string }

// pathItem is an id in an endpoint's path after what it is under: a
// deployment of an app, a task.
type pathItem struct {
	arg         string
	description string
}

// place is where one call of a tool goes: the base path, and the names found.
type place struct {
	path string
	ref  *appRef
	env  *envRef
}

// resolvePlace finds what an endpoint is under from a tool's arguments.
func resolvePlace(ctx context.Context, call *Call, u under, args map[string]any) (*place, error) {
	str := func(name string) string { s, _ := args[name].(string); return strings.TrimSpace(s) }
	switch u {
	case underProject:
		ref, err := resolveProject(ctx, call, str(argProject))
		if err != nil {
			return nil, err
		}
		return &place{path: ref.path("")}, nil
	case underEnv:
		ref, err := resolveEnv(ctx, call, str(argProject), str(argEnv))
		if err != nil {
			return nil, err
		}
		return &place{path: ref.path(""), env: ref}, nil
	case underApp:
		ref, err := resolveApp(ctx, call, str(argProject), str(argEnv), str(argApp))
		if err != nil {
			return nil, err
		}
		return &place{path: ref.path(""), ref: ref, env: &ref.envRef}, nil
	case underNothing:
	}
	return &place{}, nil
}

// ---- GET endpoint tools ----

// getEndpoint is a read tool that is one GET endpoint.
type getEndpoint struct {
	name, title, description string
	// paths are the endpoint's path for each place it may be under; a tool
	// with several takes the most specific one its arguments name.
	paths map[under]string
	item  *pathItem
	// query is the endpoint's request type, whose mapstructure fields are its
	// query parameters; nil for an endpoint that takes none.
	query any
	// params describe each query parameter, for the model.
	params map[string]string
	// omit are query parameters the endpoint takes but does nothing with, or
	// that are for the dashboard only: not offered.
	omit []string
	// answer is a new value of the endpoint's response type.
	answer func() any
}

func (e getEndpoint) tool() Tool {
	return Tool{Name: e.name, Title: e.title, Description: e.description, Kind: KindRead, needs: NeedRead,
		add: func(s *mcpsdk.Server, deps *Deps) {
			params, err := e.queryParams()
			if err != nil {
				panic(err) // a registry error: TestEveryToolBuilds finds it
			}
			sdkTool := &mcpsdk.Tool{Name: e.name, Title: e.title, Description: e.description,
				InputSchema: e.inputSchema(params),
				Annotations: &mcpsdk.ToolAnnotations{ReadOnlyHint: true, Title: e.title}}
			addTool(s, deps, sdkTool, true, func(ctx context.Context, call *Call, _ *caller, args map[string]any) (
				any, error) {
				return e.call(ctx, call, params, args)
			})
		}}
}

// queryParams are the endpoint's query parameters the tool offers.
func (e getEndpoint) queryParams() ([]queryParam, error) {
	params, err := queryParamsOf(e.query)
	if err != nil {
		return nil, err
	}
	return slices.DeleteFunc(params, func(p queryParam) bool { return slices.Contains(e.omit, p.name) }), nil
}

func (e getEndpoint) unders() []under {
	out := make([]under, 0, len(e.paths))
	for u := range e.paths {
		out = append(out, u)
	}
	slices.Sort(out)
	return out
}

func (e getEndpoint) inputSchema(params []queryParam) *jsonschema.Schema {
	s := objectSchema()
	unders := e.unders()
	optionalScope := len(unders) > 1
	for _, arg := range scopeArgs[unders[len(unders)-1]] {
		s.Properties[arg.name] = &jsonschema.Schema{Type: typeString, Description: arg.description}
		if !optionalScope {
			s.Required = append(s.Required, arg.name)
		}
	}
	if e.item != nil {
		s.Properties[e.item.arg] = &jsonschema.Schema{Type: typeString, Description: e.item.description}
		s.Required = append(s.Required, e.item.arg)
	}
	for _, p := range params {
		ps := p.schema.CloneSchemas()
		if d := e.params[p.name]; d != "" {
			ps.Description = d
		}
		s.Properties[p.name] = ps
	}
	return s
}

func objectSchema() *jsonschema.Schema {
	return &jsonschema.Schema{Type: "object", Properties: map[string]*jsonschema.Schema{},
		AdditionalProperties: &jsonschema.Schema{Not: &jsonschema.Schema{}}}
}

func (e getEndpoint) call(ctx context.Context, call *Call, params []queryParam, args map[string]any) (any, error) {
	u := e.placeFor(args)
	at, err := resolvePlace(ctx, call, u, args)
	if err != nil {
		return nil, err
	}
	path := at.path + e.paths[u]
	if e.item != nil {
		id, _ := args[e.item.arg].(string)
		if id = strings.TrimSpace(id); id == "" {
			return nil, &InputError{Message: e.item.arg + " is required"}
		}
		path = at.path + strings.Replace(e.paths[u], "{"+e.item.arg+"}", url.PathEscape(id), 1)
	}
	query, err := toQuery(params, args)
	if err != nil {
		return nil, err
	}
	answer := e.answer()
	if err = call.Get(ctx, path, query, answer); err != nil {
		return nil, err
	}
	return answer, nil
}

// placeFor is the most specific place the arguments name, of those the
// endpoint is under.
func (e getEndpoint) placeFor(args map[string]any) under {
	given := func(name string) bool { s, _ := args[name].(string); return strings.TrimSpace(s) != "" }
	unders := e.unders()
	for i := len(unders) - 1; i >= 0; i-- {
		u := unders[i]
		named := true
		for _, arg := range scopeArgs[u] {
			named = named && given(arg.name)
		}
		if named || i == 0 {
			return u
		}
	}
	return unders[0]
}

// ---- request bodies ----

// bodyInput is the input schema of a tool that sends a request body: the names
// that find where it goes, then the body's fields as its type's JSON has them,
// less those in omit. descs describe the fields, a nested one by its path, such
// as schedule.cronExpr. The fields the endpoint needs are required, as the tool
// says; the schema generator would require every field without omitempty.
func bodyInput(u under, body any, descs map[string]string, required []string, omit ...string) func() (
	*jsonschema.Schema, error) {
	return func() (*jsonschema.Schema, error) {
		s := objectSchema()
		for _, arg := range scopeArgs[u] {
			s.Properties[arg.name] = &jsonschema.Schema{Type: typeString, Description: arg.description}
			s.Required = append(s.Required, arg.name)
		}
		bs, err := schemaOf(reflect.TypeOf(body))
		if err != nil {
			return nil, err
		}
		for name, prop := range bs.Properties {
			if !slices.Contains(omit, name) {
				s.Properties[name] = withoutRequired(prop)
			}
		}
		for path, desc := range descs {
			prop := propertyAt(s, path)
			if prop == nil {
				return nil, fmt.Errorf("%w: %T has no field %s", errNoSuchField, body, path)
			}
			prop.Description = desc
		}
		s.Required = append(s.Required, required...)
		return s, nil
	}
}

// propertyAt is the schema of a field by its path, or nil.
func propertyAt(s *jsonschema.Schema, path string) *jsonschema.Schema {
	for name := range strings.SplitSeq(path, ".") {
		if s == nil {
			return nil
		}
		s = s.Properties[name]
	}
	return s
}

// withoutRequired is a schema with no field of any object in it required: a
// request's type says what it may carry, not what must be given.
func withoutRequired(s *jsonschema.Schema) *jsonschema.Schema {
	if s == nil {
		return nil
	}
	out := s.CloneSchemas()
	var walk func(s *jsonschema.Schema)
	walk = func(s *jsonschema.Schema) {
		if s == nil {
			return
		}
		s.Required = nil
		for _, p := range s.Properties {
			walk(p)
		}
		walk(s.Items)
		walk(s.AdditionalProperties)
	}
	walk(out)
	return out
}

// decodeBody reads a tool's arguments into a request body of the API's type:
// the arguments that name places are not its fields, and fall away.
func decodeBody(args map[string]any, body any) error {
	raw, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("mcp: encoding the arguments: %w", err)
	}
	if err = json.Unmarshal(raw, body); err != nil {
		return &InputError{Message: "the arguments do not fit the request: " + err.Error()}
	}
	return nil
}

// statusValues says a status's values, for a description.
func statusValues[T ~string](values []T) string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, string(v))
	}
	return strings.Join(out, ", ")
}
