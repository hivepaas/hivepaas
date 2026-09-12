package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// codePattern is what an error code identifier looks like. Anything else in a
// string literal is not our business.
// codePattern is what an error code identifier looks like: ERR_ followed by
// underscore-separated segments. The segments matter - they reject a bare
// prefix such as "ERR_VLD_", which is a constant naming a family of codes
// rather than a code, and which a looser pattern reports as a missing
// translation.
var codePattern = regexp.MustCompile(`^ERR_[A-Z0-9]+(_[A-Z0-9]+)*$`)

// declaringFuncs are the calls whose string argument declares a code rather
// than merely mentioning one.
var declaringFuncs = map[string]bool{
	"NewErr": true,
	"New":    true, // errors.New, used for the base errors in hperrors
}

// skipDirs are never scanned. Mirrors tools/goroutinelint.
var skipDirs = map[string]bool{
	"vendor":         true,
	"node_modules":   true,
	"tmp":            true,
	"temp":           true,
	"deployment":     true,
	"dist-dashboard": true,
	"test-results":   true,

	// testdata holds fixtures, including this tool's own. A code in a fixture
	// is not a code the product raises. It is still scannable by naming it as
	// a root, which is how the tests here reach their sample package.
	"testdata": true,
}

// codeUse is one appearance of an error code in Go source.
type codeUse struct {
	Code     string
	Pos      token.Position
	Declared bool

	// Ident is the variable a declaration binds the code to, empty otherwise.
	// Code is what the wire carries; Ident is what Go code refers to, and the
	// two have to be tracked separately: a declaration mentions the string once
	// and every use afterwards mentions only the name.
	Ident string
}

func skipDir(name string) bool {
	return skipDirs[name] || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

// scanGoDirs collects every error code appearing in Go sources under roots.
//
// Fixtures are excluded: testdata directories and _test.go files hold codes
// that exist to exercise something, not codes the product can raise.
func scanGoDirs(fset *token.FileSet, roots []string) ([]codeUse, error) {
	var uses []codeUse
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if path != root && skipDir(d.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			// _test.go is skipped: a code appearing only in a test is a
			// fixture, and requiring a translation for one would mean shipping
			// message entries that exist to satisfy a test.
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return err
			}
			uses = append(uses, codesInFile(fset, file)...)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return uses, nil
}

// codesInFile walks one parsed file.
//
// Declarations are collected first so that the literal inside a declaring call
// is not also counted as a bare reference.
func codesInFile(fset *token.FileSet, file *ast.File) []codeUse {
	declared := map[*ast.BasicLit]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !declaringFuncs[calleeName(call.Fun)] {
			return true
		}
		for _, arg := range call.Args {
			if lit, ok := arg.(*ast.BasicLit); ok && codePattern.MatchString(litValue(lit)) {
				declared[lit] = true
			}
		}
		return true
	})

	// The name each declaring literal is bound to. A ValueSpec pairs its names
	// with its values by position, which is how "ErrX = NewErr(_, \"ERR_X\")"
	// connects the two.
	boundTo := map[*ast.BasicLit]string{}
	ast.Inspect(file, func(n ast.Node) bool {
		spec, ok := n.(*ast.ValueSpec)
		if !ok {
			return true
		}
		for i, value := range spec.Values {
			if i >= len(spec.Names) {
				break
			}
			call, ok := value.(*ast.CallExpr)
			if !ok || !declaringFuncs[calleeName(call.Fun)] {
				continue
			}
			for _, arg := range call.Args {
				if lit, ok := arg.(*ast.BasicLit); ok && declared[lit] {
					boundTo[lit] = spec.Names[i].Name
				}
			}
		}
		return true
	})

	var uses []codeUse
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok {
			return true
		}
		code := litValue(lit)
		if !codePattern.MatchString(code) {
			return true
		}
		uses = append(uses, codeUse{
			Code:     code,
			Pos:      fset.Position(lit.Pos()),
			Declared: declared[lit],
			Ident:    boundTo[lit],
		})
		return true
	})
	return uses
}

// calleeName is the function name of a call, ignoring any package qualifier, so
// that NewErr and hperrors.NewErr both answer "NewErr".
func calleeName(fun ast.Expr) string {
	switch f := fun.(type) {
	case *ast.Ident:
		return f.Name
	case *ast.SelectorExpr:
		return f.Sel.Name
	}
	return ""
}

// litValue is the unquoted contents of a string literal, or "" for anything else.
func litValue(lit *ast.BasicLit) string {
	if lit.Kind != token.STRING {
		return ""
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil {
		return ""
	}
	return s
}

// scanIdentUses counts every appearance of the given identifiers, the
// declaration included, so that a count of one means declared and referenced
// nowhere.
//
// Unlike scanGoDirs this reads _test.go too. A code a test still names is not
// one to report as abandoned, and counting conservatively costs only a finding
// nobody needed.
func scanIdentUses(fset *token.FileSet, roots []string, names map[string]bool) (map[string]int, error) {
	counts := map[string]int{}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if path != root && skipDir(d.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return err
			}
			ast.Inspect(file, func(n ast.Node) bool {
				if ident, ok := n.(*ast.Ident); ok && names[ident.Name] {
					counts[ident.Name]++
				}
				return true
			})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return counts, nil
}
