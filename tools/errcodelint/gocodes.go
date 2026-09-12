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
var codePattern = regexp.MustCompile(`^ERR_[A-Z0-9_]+$`)

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
}

// codeUse is one appearance of an error code in Go source.
type codeUse struct {
	Code     string
	Pos      token.Position
	Declared bool
}

func skipDir(name string) bool {
	return skipDirs[name] || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

// scanGoDirs collects every error code appearing in Go sources under roots.
//
// testdata is scanned here, unlike in goroutinelint: this tool's own fixtures
// live there, and a stray code in someone else's testdata is worth reporting
// rather than hiding.
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
			if !strings.HasSuffix(path, ".go") {
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
