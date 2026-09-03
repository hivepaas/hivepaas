// Command goroutinelint reports goroutines that are started without a panic guard.
//
// A panic in a bare `go func()` cannot be recovered by whoever started it: gin's
// recovery middleware and safego.RecoverTo only protect the goroutine they
// run in, and sync.WaitGroup.Go does not recover either. Such a panic takes the
// whole process down.
//
// The check is deliberately not "ban `go func`" - most goroutines here are
// legitimate. It reports only the ones whose body does not start with a
// recovering defer.
//
// A goroutine is considered guarded when its body has a top-level defer of:
//
//   - safego.Recover / RecoverWithLogger / RecoverPipe / RecoverTo (or the
//     unqualified form inside package safego itself)
//   - a function literal that calls recover()
//
// It also rejects safego.RecoverTo(nil): with nowhere to return the panic to,
// that call is a misuse - use safego.Recover, which logs it.
//
// For `go f()` / `go x.m()` the same check is applied to the declaration of f/m
// in the same package. A callee this tool cannot see (another package) is
// reported: wrap it in safego.Go instead.
//
// Escape hatch: put `//safego:allow <reason>` on the goroutine line or the line
// above it. A reason is required.
//
// Usage:
//
//	go run ./tools/goroutinelint [dir...]   # default: the current directory
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// guardFuncs are the helpers that recover a panic for the current goroutine.
var guardFuncs = map[string]bool{
	"Recover":           true,
	"RecoverWithLogger": true,
	"RecoverPipe":       true,
	"RecoverTo":         true,
}

// skipDirs are never scanned.
var skipDirs = map[string]bool{
	"vendor":         true,
	"node_modules":   true,
	"testdata":       true,
	"tmp":            true,
	"temp":           true,
	"deployment":     true,
	"dist-dashboard": true,
	"test-results":   true,
}

const allowDirective = "//safego:allow"

type finding struct {
	pos token.Position
	msg string
}

func main() {
	flag.Parse()
	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}

	fset := token.NewFileSet()
	var findings []finding
	for _, root := range roots {
		dirs, err := collectDirs(root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "goroutinelint: %v\n", err)
			os.Exit(2)
		}
		for _, dir := range dirs {
			f, err := checkDir(fset, dir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "goroutinelint: %v\n", err)
				os.Exit(2)
			}
			findings = append(findings, f...)
		}
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].pos.Filename != findings[j].pos.Filename {
			return findings[i].pos.Filename < findings[j].pos.Filename
		}
		return findings[i].pos.Line < findings[j].pos.Line
	})
	for _, f := range findings {
		fmt.Printf("%s:%d:%d: %s\n", f.pos.Filename, f.pos.Line, f.pos.Column, f.msg)
	}
	if len(findings) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d problem(s). Start goroutines with safego.Go, or add"+
			" `defer safego.Recover(\"name\")` as the first statement of the body.\n", len(findings))
		os.Exit(1)
	}
}

func collectDirs(root string) ([]string, error) {
	seen := map[string]bool{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if path != root && (skipDirs[name] || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")) {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			seen[filepath.Dir(path)] = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	dirs := make([]string, 0, len(seen))
	for d := range seen {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	return dirs, nil
}

// pkgContext holds everything needed to resolve callees within one package.
type pkgContext struct {
	fset *token.FileSet
	// decls maps a function or method name to every declaration with that name.
	decls map[string][]*ast.FuncDecl
	// allowed holds the lines carrying a //safego:allow directive, per file.
	allowed map[string]map[int]bool
	// imports holds the package-qualifier idents used in each file.
	imports map[string]map[string]bool
}

func checkDir(fset *token.FileSet, dir string) ([]finding, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	ctx := &pkgContext{
		fset:    fset,
		decls:   map[string][]*ast.FuncDecl{},
		allowed: map[string]map[int]bool{},
		imports: map[string]map[string]bool{},
	}
	var files []*ast.File
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		files = append(files, file)

		ctx.allowed[path] = allowedLines(fset, file)
		ctx.imports[path] = importIdents(file)
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
				ctx.decls[fn.Name.Name] = append(ctx.decls[fn.Name.Name], fn)
			}
		}
	}

	var findings []finding
	for _, file := range files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.GoStmt:
				findings = append(findings, ctx.checkGoStmt(node)...)
			case *ast.CallExpr:
				findings = append(findings, ctx.checkWaitGroupGo(node)...)
				findings = append(findings, ctx.checkRecoverToNil(node)...)
			}
			return true
		})
	}
	return findings, nil
}

// checkGoStmt handles `go func(){...}()`, `go f()` and `go x.m()`.
func (c *pkgContext) checkGoStmt(stmt *ast.GoStmt) []finding {
	pos := c.fset.Position(stmt.Pos())
	if c.isAllowed(pos) {
		return nil
	}

	switch fn := ast.Unparen(stmt.Call.Fun).(type) {
	case *ast.FuncLit:
		if hasPanicGuard(fn.Body) {
			return nil
		}
		return []finding{{pos, "goroutine started without a panic guard: a panic here kills the" +
			" process. Use safego.Go, or `defer safego.Recover(\"name\")` as the first statement"}}

	case *ast.Ident:
		return c.checkCallee(pos, fn.Name, "")

	case *ast.SelectorExpr:
		// `go pkg.Func()` cannot be resolved without type info.
		if x, ok := fn.X.(*ast.Ident); ok && c.imports[pos.Filename][x.Name] {
			return []finding{{pos, fmt.Sprintf("goroutine calls %s.%s in another package, its panic"+
				" guard cannot be verified here. Use safego.Go instead", x.Name, fn.Sel.Name)}}
		}
		return c.checkCallee(pos, fn.Sel.Name, "")
	}
	return nil
}

// checkWaitGroupGo handles `wg.Go(func(){...})`. sync.WaitGroup.Go and
// errgroup.Group.Go do not recover panics either.
func (c *pkgContext) checkWaitGroupGo(call *ast.CallExpr) []finding {
	sel, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Go" || len(call.Args) == 0 {
		return nil
	}
	// safego.Go already guards its callback.
	if x, ok := sel.X.(*ast.Ident); ok && x.Name == "safego" {
		return nil
	}
	lit, ok := ast.Unparen(call.Args[len(call.Args)-1]).(*ast.FuncLit)
	if !ok || hasPanicGuard(lit.Body) {
		return nil
	}
	pos := c.fset.Position(call.Pos())
	if c.isAllowed(pos) {
		return nil
	}
	return []finding{{pos, "goroutine started without a panic guard: WaitGroup.Go does not recover," +
		" a panic here kills the process. Add `defer safego.Recover(\"name\")` as the first statement"}}
}

// checkRecoverToNil rejects `RecoverTo(nil)`: there is nowhere to return the
// panic to, so the call would only hide it.
func (c *pkgContext) checkRecoverToNil(call *ast.CallExpr) []finding {
	var name string
	switch fn := ast.Unparen(call.Fun).(type) {
	case *ast.SelectorExpr:
		name = fn.Sel.Name
	case *ast.Ident:
		name = fn.Name
	}
	if name != "RecoverTo" || len(call.Args) != 1 {
		return nil
	}
	if id, ok := ast.Unparen(call.Args[0]).(*ast.Ident); !ok || id.Name != "nil" {
		return nil
	}
	pos := c.fset.Position(call.Pos())
	if c.isAllowed(pos) {
		return nil
	}
	return []finding{{pos, "RecoverTo(nil) has nowhere to return the panic to and only hides it." +
		" Pass a *error, or use safego.Recover(\"name\") which logs the panic"}}
}

// checkCallee verifies the same-package declaration(s) of a goroutine entry point.
func (c *pkgContext) checkCallee(pos token.Position, name, _ string) []finding {
	decls := c.decls[name]
	if len(decls) == 0 {
		return []finding{{pos, fmt.Sprintf("goroutine calls %s, which is not declared in this package,"+
			" its panic guard cannot be verified. Use safego.Go instead", name)}}
	}
	for _, decl := range decls {
		if !hasPanicGuard(decl.Body) {
			return []finding{{pos, fmt.Sprintf("goroutine entry point %s has no panic guard (declared at"+
				" %s): a panic there kills the process. Add `defer safego.Recover(\"name\")` to it",
				name, c.fset.Position(decl.Pos()))}}
		}
	}
	return nil
}

// hasPanicGuard reports whether the body defers something that recovers.
func hasPanicGuard(body *ast.BlockStmt) bool {
	if body == nil {
		return false
	}
	for _, stmt := range body.List {
		deferStmt, ok := stmt.(*ast.DeferStmt)
		if !ok {
			continue
		}
		switch fn := ast.Unparen(deferStmt.Call.Fun).(type) {
		case *ast.FuncLit:
			if callsRecover(fn.Body) {
				return true
			}
		case *ast.Ident: // `defer Recover(name)` inside package safego
			if guardFuncs[fn.Name] {
				return true
			}
		case *ast.SelectorExpr: // `defer safego.Recover(name)`
			if guardFuncs[fn.Sel.Name] {
				return true
			}
		}
	}
	return false
}

func callsRecover(n ast.Node) bool {
	found := false
	ast.Inspect(n, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if id, ok := ast.Unparen(call.Fun).(*ast.Ident); ok && id.Name == "recover" {
			found = true
			return false
		}
		return true
	})
	return found
}

// importIdents collects the qualifiers a file uses to reference other packages.
func importIdents(file *ast.File) map[string]bool {
	idents := map[string]bool{}
	for _, imp := range file.Imports {
		if imp.Name != nil {
			if imp.Name.Name != "_" && imp.Name.Name != "." {
				idents[imp.Name.Name] = true
			}
			continue
		}
		path := strings.Trim(imp.Path.Value, `"`)
		idents[path[strings.LastIndex(path, "/")+1:]] = true
	}
	return idents
}

func (c *pkgContext) isAllowed(pos token.Position) bool {
	lines := c.allowed[pos.Filename]
	return lines[pos.Line] || lines[pos.Line-1]
}

// allowedLines indexes the lines carrying a `//safego:allow <reason>` directive.
// The directive is ignored when no reason is given.
func allowedLines(fset *token.FileSet, file *ast.File) map[int]bool {
	lines := map[int]bool{}
	for _, group := range file.Comments {
		for _, comment := range group.List {
			text := strings.TrimSpace(comment.Text)
			if !strings.HasPrefix(text, allowDirective) {
				continue
			}
			if strings.TrimSpace(strings.TrimPrefix(text, allowDirective)) == "" {
				continue // a reason is required
			}
			lines[fset.Position(comment.Pos()).Line] = true
		}
	}
	return lines
}
