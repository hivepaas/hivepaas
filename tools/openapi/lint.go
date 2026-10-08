package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"maps"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

const handlerDir = "/hivepaas_app/interface/api/handler"

var (
	routerPattern    = regexp.MustCompile(`@Router\s+(\S+)\s+\[(\w+)\]`)
	paramPattern     = regexp.MustCompile(`@Param\s+(\S+)\s+(\w+)\s`)
	pathParamPattern = regexp.MustCompile(`\{([^}]+)\}`)
	// ignorePattern marks a parameter a handler decodes and deliberately
	// leaves out of its API - a field of a request type it shares with
	// endpoints where it means something: "openapi:ignore-param revealSecrets
	// - nothing listed here has a secret". Swag does not read it.
	ignorePattern = regexp.MustCompile(`openapi:ignore-param\s+(\S+)`)

	// pagingParams are what a handler passing a Paging reads beside its
	// request type. pageOffset and pageLimit are documented wherever they are
	// read; sort only where the list honors it, which some refuse.
	pagingParams = []string{"pageOffset", "pageLimit"}
	pagingSort   = "sort"
)

// queryParsers are the BaseHandler methods that decode the query and the form
// into a request type, by its mapstructure tags: (ctx, request, paging).
var queryParsers = map[string]bool{"ParseAndValidateRequest": true, "ParseRequest": true}

// parseHelpers are the BaseHandler methods behind queryParsers, whose reading
// is modeled rather than followed: inside them, the paging's own parameters
// are read by name.
var parseHelpers = map[string]bool{
	"ParseAndValidateRequest": true, "ParseRequest": true, "parseQuery": true, "parsePagination": true,
	"ParseAndValidateJSONBody": true, "ParseJSONBody": true,
}

// queryReaders are gin's methods reading one query or form value by name.
var queryReaders = map[string]bool{
	"Query": true, "DefaultQuery": true, "GetQuery": true, "QueryArray": true, "GetQueryArray": true,
	"QueryMap": true, "GetQueryMap": true, "PostForm": true, "DefaultPostForm": true, "GetPostForm": true,
	"PostFormArray": true, "GetPostFormArray": true, "FormFile": true, "FormValue": true,
}

type lintProblem struct {
	pos     token.Position
	handler string
	message string
}

func (p lintProblem) String() string {
	return fmt.Sprintf("%s:%d: %s: %s", p.pos.Filename, p.pos.Line, p.handler, p.message)
}

func runLint(_ []string, out io.Writer) error {
	src, root, err := loadModule()
	if err != nil {
		return err
	}
	problems, checked := lintHandlers(src)
	for _, problem := range problems {
		problem.pos.Filename = strings.TrimPrefix(problem.pos.Filename, root+"/")
		fmt.Fprintln(out, problem)
	}
	if len(problems) > 0 {
		return fmt.Errorf("%w: %d problem(s)", errProblems, len(problems))
	}
	fmt.Fprintf(out, "OK: %d documented handlers\n", checked)
	return nil
}

// lintHandlers checks every documented handler of the API, and answers how
// many there were.
func lintHandlers(src *goSource) ([]lintProblem, int) {
	var problems []lintProblem
	checked := 0
	for _, pkgPath := range slices.Sorted(maps.Keys(src.byPath)) {
		if !strings.HasPrefix(pkgPath, src.module+handlerDir) {
			continue
		}
		for _, file := range src.byPath[pkgPath].files {
			for _, decl := range file.ast.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Doc == nil || fn.Body == nil || !routerPattern.MatchString(fn.Doc.Text()) {
					continue
				}
				checked++
				problems = append(problems, lintHandler(src, file, fn)...)
			}
		}
	}
	sort.Slice(problems, func(i, j int) bool {
		if problems[i].pos.Filename != problems[j].pos.Filename {
			return problems[i].pos.Filename < problems[j].pos.Filename
		}
		return problems[i].pos.Line < problems[j].pos.Line
	})
	return problems, checked
}

func lintHandler(src *goSource, file *goFile, fn *ast.FuncDecl) []lintProblem {
	doc := fn.Doc.Text()
	pos := src.fset.Position(fn.Pos())
	var problems []lintProblem
	report := func(format string, args ...any) {
		problems = append(problems, lintProblem{pos: pos, handler: fn.Name.Name, message: fmt.Sprintf(format, args...)})
	}

	documented := map[string]map[string]bool{}
	for _, m := range paramPattern.FindAllStringSubmatch(doc, -1) {
		if documented[m[2]] == nil {
			documented[m[2]] = map[string]bool{}
		}
		documented[m[2]][m[1]] = true
	}

	// The path's parameters and the documented ones are the same.
	for _, route := range routerPattern.FindAllStringSubmatch(doc, -1) {
		inPath := map[string]bool{}
		for _, m := range pathParamPattern.FindAllStringSubmatch(route[1], -1) {
			inPath[m[1]] = true
			if !documented["path"][m[1]] {
				report("the route %s has {%s}, which no @Param ... path documents", route[1], m[1])
			}
		}
		for name := range documented["path"] {
			if !inPath[name] {
				report("@Param %s path is not in the route %s", name, route[1])
			}
		}
	}

	reads, paging, known := handlerReads(src, file, fn)
	given := map[string]bool{}
	maps.Copy(given, documented["query"])
	maps.Copy(given, documented["formData"])
	ignored := map[string]bool{}
	for _, m := range ignorePattern.FindAllStringSubmatch(doc, -1) {
		ignored[m[1]] = true
		if !reads[m[1]] {
			report("ignores %q, which it does not read from the query or the form", m[1])
		}
	}
	for _, name := range slices.Sorted(maps.Keys(reads)) {
		if !given[name] && !ignored[name] {
			report("reads %q from the query or the form, and no @Param ... query (or formData) documents it", name)
		}
	}
	if paging {
		for _, name := range pagingParams {
			if !given[name] {
				report("pages its list, and no @Param ... query documents %q", name)
			}
		}
	}
	if !known {
		// What the request type reads could not be told: what is documented
		// cannot be checked against it.
		return problems
	}
	for _, name := range slices.Sorted(maps.Keys(given)) {
		if reads[name] || (paging && (slices.Contains(pagingParams, name) || name == pagingSort)) {
			continue
		}
		report("documents %q, which it does not read from the query or the form", name)
	}
	return problems
}

// maxDelegation is how many methods deep a handler is followed to where it
// reads its request.
const maxDelegation = 2

// handlerReads are the query and form parameters a handler reads: the
// mapstructure fields of the request types it decodes, and the names it reads
// one by one. known is false when a decoded request's type could not be told.
func handlerReads(src *goSource, file *goFile, fn *ast.FuncDecl) (reads map[string]bool, paging, known bool) {
	reads = map[string]bool{}
	paging, known, _ = src.readsIn(readScope{fn: &funcDecl{decl: fn, file: file}}, reads, 0)
	return reads, paging, known
}

// readScope is a function read for what it decodes, with the constants its
// caller passed for its parameters, by parameter name.
type readScope struct {
	fn     *funcDecl
	consts map[string]string
}

// readsIn adds what a function reads of the query to reads, and says whether
// it reads any at all. What a handler hands to a method of its own is read
// through that method - the settings handlers pass their resource type to one
// ListSetting for all, which reads the branch of its switch for that type.
func (s *goSource) readsIn(scope readScope, reads map[string]bool, depth int) (paging, known, found bool) {
	file, body := scope.fn.file, scope.fn.decl.Body
	known = true
	ast.Inspect(body, func(node ast.Node) bool {
		if name := formIndex(node); name != "" {
			found = true
			reads[name] = true
			return true
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		switch {
		case queryParsers[sel.Sel.Name] && len(call.Args) == 3: //nolint:mnd
			found = true
			decl := s.requestType(file, scope.fn.decl, call.Args[1])
			if decl == nil {
				decl = s.switchedRequest(scope)
			}
			if decl == nil {
				known = false
				return true
			}
			s.queryFields(decl, 0, reads)
			paging = paging || s.pages(call.Args[2], decl)
		case queryReaders[sel.Sel.Name] && len(call.Args) > 0:
			if lit, isLit := call.Args[0].(*ast.BasicLit); isLit && lit.Kind == token.STRING {
				if name, err := strconv.Unquote(lit.Value); err == nil {
					found = true
					reads[name] = true
				}
			}
		}
		return true
	})
	if depth >= maxDelegation {
		return paging, known, found
	}

	recvDecl, recvName := s.receiver(scope.fn)
	if recvDecl == nil {
		return paging, known, found
	}
	ast.Inspect(body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		owner := s.calledOn(recvDecl, recvName, sel.X)
		if owner == nil || parseHelpers[sel.Sel.Name] {
			return true
		}
		callee := s.method(owner, sel.Sel.Name, 0)
		if callee == nil || callee.decl.Body == nil {
			return true
		}
		p, k, f := s.readsIn(readScope{fn: callee, consts: boundConsts(callee.decl, call.Args)}, reads, depth+1)
		if f {
			found, paging, known = true, paging || p, known && k
		}
		return true
	})
	return paging, known, found
}

// calledOn is the type a method call is made on when it is the receiver - h -
// or one of its fields - h.TaskHandler - and nil otherwise.
func (s *goSource) calledOn(recvDecl *typeDecl, recvName string, expr ast.Expr) *typeDecl {
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Name == recvName {
			return recvDecl
		}
	case *ast.SelectorExpr:
		x, ok := e.X.(*ast.Ident)
		if !ok || x.Name != recvName {
			return nil
		}
		return s.fieldType(recvDecl, e.Sel.Name, 0)
	}
	return nil
}

// fieldType is the type of a struct's field, declared on it or promoted from
// a struct it embeds, or nil.
func (s *goSource) fieldType(decl *typeDecl, name string, depth int) *typeDecl {
	st, file := s.structOf(decl)
	if st == nil || depth > maxTypeDepth {
		return nil
	}
	var embedded []*typeDecl
	for _, field := range st.Fields.List {
		ft := field.Type
		if star, isStar := ft.(*ast.StarExpr); isStar {
			ft = star.X
		}
		named := len(field.Names) == 0 && typeName(ft) == name
		for _, ident := range field.Names {
			named = named || ident.Name == name
		}
		inner, _ := s.lookup(typeRef{ft, file})
		if named {
			return inner
		}
		if len(field.Names) == 0 && inner != nil {
			embedded = append(embedded, inner)
		}
	}
	for _, inner := range embedded {
		if found := s.fieldType(inner, name, depth+1); found != nil {
			return found
		}
	}
	return nil
}

// formIndex is the name a multipart form is read by - form.File["file"],
// form.Value["path"] - or "".
func formIndex(node ast.Node) string {
	index, ok := node.(*ast.IndexExpr)
	if !ok {
		return ""
	}
	sel, ok := index.X.(*ast.SelectorExpr)
	if !ok || (sel.Sel.Name != "File" && sel.Sel.Name != "Value") {
		return ""
	}
	lit, ok := index.Index.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return ""
	}
	name, err := strconv.Unquote(lit.Value)
	if err != nil {
		return ""
	}
	return name
}

// pages says whether a parse call is given a Paging: &req.Paging, a
// PagingReq() call, or a variable set from the request when its type has one.
func (s *goSource) pages(arg ast.Expr, decl *typeDecl) bool {
	switch a := arg.(type) {
	case *ast.Ident:
		if a.Name == "nil" {
			return false
		}
		return s.method(decl, "PagingReq", 0) != nil
	default:
		return true
	}
}

// receiver is the type a method is declared on, and the name it calls its
// receiver.
func (s *goSource) receiver(fn *funcDecl) (*typeDecl, string) {
	if fn.decl.Recv == nil || len(fn.decl.Recv.List) == 0 || len(fn.decl.Recv.List[0].Names) == 0 {
		return nil, ""
	}
	recv := fn.decl.Recv.List[0]
	return fn.file.pkg.types[receiverType(recv.Type)], recv.Names[0].Name
}

// boundConsts are the constants a call passes, by the callee's parameter
// names: base.ResourceTypeIMService for resType.
func boundConsts(callee *ast.FuncDecl, args []ast.Expr) map[string]string {
	consts := map[string]string{}
	i := 0
	for _, field := range callee.Type.Params.List {
		for _, name := range field.Names {
			if i < len(args) {
				if c := constName(args[i]); c != "" {
					consts[name.Name] = c
				}
			}
			i++
		}
	}
	return consts
}

func constName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		return e.Sel.Name
	case *ast.Ident:
		return e.Name
	}
	return ""
}

// switchedRequest is the request type a function builds in the branch of a
// switch on one of its parameters that its caller's constant selects.
func (s *goSource) switchedRequest(scope readScope) *typeDecl {
	var clause *ast.CaseClause
	ast.Inspect(scope.fn.decl.Body, func(node ast.Node) bool {
		sw, ok := node.(*ast.SwitchStmt)
		if !ok || clause != nil {
			return clause == nil
		}
		tag, ok := sw.Tag.(*ast.Ident)
		if !ok || scope.consts[tag.Name] == "" {
			return true
		}
		for _, stmt := range sw.Body.List {
			cc, isClause := stmt.(*ast.CaseClause)
			if !isClause {
				continue
			}
			for _, expr := range cc.List {
				if constName(expr) == scope.consts[tag.Name] {
					clause = cc
				}
			}
		}
		return true
	})
	if clause == nil {
		return nil
	}
	var decl *typeDecl
	for _, stmt := range clause.Body {
		ast.Inspect(stmt, func(node ast.Node) bool {
			assign, ok := node.(*ast.AssignStmt)
			if !ok || decl != nil {
				return decl == nil
			}
			for _, rhs := range assign.Rhs {
				if d := s.typeOfExpr(typeRef{rhs, scope.fn.file}); d != nil {
					decl = d
					return false
				}
			}
			return true
		})
	}
	return decl
}

// requestType is the struct a handler decodes into: the type of the variable
// it passes, as declared or assigned in the handler.
func (s *goSource) requestType(file *goFile, fn *ast.FuncDecl, arg ast.Expr) *typeDecl {
	if unary, ok := arg.(*ast.UnaryExpr); ok && unary.Op == token.AND {
		arg = unary.X
	}
	ident, ok := arg.(*ast.Ident)
	if !ok {
		return nil
	}
	var found ast.Expr
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if found != nil {
			return false
		}
		switch n := node.(type) {
		case *ast.AssignStmt:
			for i, lhs := range n.Lhs {
				if id, isIdent := lhs.(*ast.Ident); isIdent && id.Name == ident.Name && i < len(n.Rhs) {
					found = n.Rhs[i]
				}
			}
		case *ast.ValueSpec:
			for i, name := range n.Names {
				if name.Name != ident.Name {
					continue
				}
				if n.Type != nil {
					found = n.Type
				} else if i < len(n.Values) {
					found = n.Values[i]
				}
			}
		}
		return true
	})
	if found == nil {
		return nil
	}
	return s.typeOfExpr(typeRef{found, file})
}

// typeOfExpr is the struct an expression yields: a constructor's result, a
// composite literal, new(T), or a type itself.
func (s *goSource) typeOfExpr(ref typeRef) *typeDecl {
	expr := ref.expr
	if unary, ok := expr.(*ast.UnaryExpr); ok && unary.Op == token.AND {
		expr = unary.X
	}
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	switch e := expr.(type) {
	case *ast.CompositeLit:
		return s.typeOfExpr(typeRef{e.Type, ref.file})
	case *ast.CallExpr:
		if id, ok := e.Fun.(*ast.Ident); ok && id.Name == "new" && len(e.Args) == 1 {
			return s.typeOfExpr(typeRef{e.Args[0], ref.file})
		}
		fn := s.lookupFunc(typeRef{e.Fun, ref.file})
		if fn == nil || fn.decl.Type.Results == nil || len(fn.decl.Type.Results.List) == 0 {
			return nil
		}
		return s.typeOfExpr(typeRef{fn.decl.Type.Results.List[0].Type, fn.file})
	case *ast.Ident, *ast.SelectorExpr:
		decl, _ := s.lookup(typeRef{e, ref.file})
		return decl
	}
	return nil
}

func (s *goSource) lookupFunc(ref typeRef) *funcDecl {
	switch e := ref.expr.(type) {
	case *ast.Ident:
		return ref.file.pkg.funcs[e.Name]
	case *ast.SelectorExpr:
		x, ok := e.X.(*ast.Ident)
		if !ok {
			return nil
		}
		if pkg := s.byPath[ref.file.importPath(s, x.Name)]; pkg != nil {
			return pkg.funcs[e.Sel.Name]
		}
	}
	return nil
}

// queryFields are the names a request type decodes from the query: its
// mapstructure tags, those of the structs it embeds squashed into it, as the
// decoder is told to.
func (s *goSource) queryFields(decl *typeDecl, depth int, out map[string]bool) {
	st, file := s.structOf(decl)
	if st == nil || depth > maxTypeDepth {
		return
	}
	for _, field := range st.Fields.List {
		name, _, _ := strings.Cut(fieldTag(field).Get("mapstructure"), ",")
		if name == "-" {
			continue
		}
		if len(field.Names) == 0 && name == "" {
			embedded := field.Type
			if star, ok := embedded.(*ast.StarExpr); ok {
				embedded = star.X
			}
			if inner, _ := s.lookup(typeRef{embedded, file}); inner != nil {
				s.queryFields(inner, depth+1, out)
			}
			continue
		}
		if name != "" {
			out[name] = true
		}
	}
}
