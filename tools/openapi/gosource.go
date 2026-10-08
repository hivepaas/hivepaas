package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

// goSource is the module's Go declarations, read from source alone: enough to
// tell what a struct is on the wire, without type-checking anything.
type goSource struct {
	module string
	fset   *token.FileSet
	byPath map[string]*goPackage
	byName map[string][]*goPackage
}

type goPackage struct {
	path  string
	name  string
	types map[string]*typeDecl
	funcs map[string]*funcDecl
	// methods are the methods declared on each type, pointer or not, by name.
	methods map[string]map[string]*funcDecl
	files   []*goFile
}

type goFile struct {
	pkg  *goPackage
	ast  *ast.File
	name string
}

type typeDecl struct {
	spec *ast.TypeSpec
	file *goFile
}

type funcDecl struct {
	decl *ast.FuncDecl
	file *goFile
}

// typeRef is a type expression and the file it is written in, which is what
// its names are resolved against.
type typeRef struct {
	expr ast.Expr
	file *goFile
}

// loadGoSource parses every non-test Go file under dirs, relative to root,
// whose module path is module.
func loadGoSource(root, module string, dirs ...string) (*goSource, error) {
	fset := token.NewFileSet()
	src := &goSource{module: module, fset: fset, byPath: map[string]*goPackage{}, byName: map[string][]*goPackage{}}
	for _, dir := range dirs {
		err := filepath.WalkDir(filepath.Join(root, dir), func(p string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				switch entry.Name() {
				case "vendor", "testdata", "node_modules", ".git":
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
				return nil
			}
			file, err := parser.ParseFile(fset, p, nil, parser.ParseComments|parser.SkipObjectResolution)
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(root, filepath.Dir(p))
			if err != nil {
				return err
			}
			src.add(path.Join(module, filepath.ToSlash(rel)), p, file)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return src, nil
}

func (s *goSource) add(importPath, fileName string, file *ast.File) {
	pkg := s.byPath[importPath]
	if pkg == nil {
		pkg = &goPackage{
			path: importPath, name: file.Name.Name,
			types: map[string]*typeDecl{}, funcs: map[string]*funcDecl{}, methods: map[string]map[string]*funcDecl{},
		}
		s.byPath[importPath] = pkg
		s.byName[pkg.name] = append(s.byName[pkg.name], pkg)
	}
	gf := &goFile{pkg: pkg, ast: file, name: fileName}
	pkg.files = append(pkg.files, gf)
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					pkg.types[ts.Name.Name] = &typeDecl{spec: ts, file: gf}
				}
			}
		case *ast.FuncDecl:
			if d.Recv == nil {
				pkg.funcs[d.Name.Name] = &funcDecl{decl: d, file: gf}
				continue
			}
			if recv := receiverType(d.Recv.List[0].Type); recv != "" {
				if pkg.methods[recv] == nil {
					pkg.methods[recv] = map[string]*funcDecl{}
				}
				pkg.methods[recv][d.Name.Name] = &funcDecl{decl: d, file: gf}
			}
		}
	}
}

func receiverType(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.StarExpr:
		return receiverType(e.X)
	case *ast.Ident:
		return e.Name
	case *ast.IndexExpr:
		return receiverType(e.X)
	}
	return ""
}

// importPath is the package a file imports under name, or "".
func (f *goFile) importPath(s *goSource, name string) string {
	for _, spec := range f.ast.Imports {
		p, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		local := ""
		switch {
		case spec.Name != nil:
			local = spec.Name.Name
		case s.byPath[p] != nil:
			local = s.byPath[p].name
		default:
			local = path.Base(p)
		}
		if local == name {
			return p
		}
	}
	return ""
}

// lookup finds the declaration a type name refers to: one of the module's, or
// none, in which case external is the package path and name, such as
// "encoding/json.RawMessage", or a builtin's name.
func (s *goSource) lookup(ref typeRef) (decl *typeDecl, external string) {
	switch e := ref.expr.(type) {
	case *ast.Ident:
		if d := ref.file.pkg.types[e.Name]; d != nil {
			return d, ""
		}
		return nil, e.Name
	case *ast.SelectorExpr:
		x, ok := e.X.(*ast.Ident)
		if !ok {
			return nil, ""
		}
		p := ref.file.importPath(s, x.Name)
		if pkg := s.byPath[p]; pkg != nil {
			if d := pkg.types[e.Sel.Name]; d != nil {
				return d, ""
			}
		}
		return nil, p + "." + e.Sel.Name
	case *ast.IndexExpr: // a generic type's instance: its fields are its type's
		return s.lookup(typeRef{e.X, ref.file})
	}
	return nil, ""
}

// wireKind is what a type can be on the wire, as far as a schema is concerned.
type wireKind int

const (
	// kindValue is never null: a string, a number, a struct.
	kindValue wireKind = iota
	// kindNullable is null when nil: a pointer, a slice, a map.
	kindNullable
	// kindAny is any JSON value: an interface, json.RawMessage.
	kindAny
	// kindCustom writes itself with MarshalJSON or MarshalText; its schema is
	// whatever swag was told it is.
	kindCustom
)

// wireShape is a type as its schema sees it: its kind, and the element of a
// slice or the value of a map, for what they hold.
type wireShape struct {
	kind wireKind
	// elem is a slice's element or a map's value, and the pointer's target.
	elem    *typeRef
	isSlice bool
	isMap   bool
}

const maxTypeDepth = 20

func (s *goSource) shape(ref typeRef) wireShape {
	return s.shapeAt(ref, 0)
}

func (s *goSource) shapeAt(ref typeRef, depth int) wireShape {
	if depth > maxTypeDepth {
		return wireShape{kind: kindValue}
	}
	switch e := ref.expr.(type) {
	case *ast.StarExpr:
		// What a pointer points at decides its schema; being a pointer adds null.
		return wireShape{kind: kindNullable, elem: &typeRef{e.X, ref.file}}
	case *ast.ArrayType:
		if e.Len != nil {
			return wireShape{kind: kindValue}
		}
		return wireShape{kind: kindNullable, elem: &typeRef{e.Elt, ref.file}, isSlice: true}
	case *ast.MapType:
		return wireShape{kind: kindNullable, elem: &typeRef{e.Value, ref.file}, isMap: true}
	case *ast.InterfaceType:
		return wireShape{kind: kindAny}
	case *ast.Ident, *ast.SelectorExpr, *ast.IndexExpr:
		decl, external := s.lookup(ref)
		if decl == nil {
			switch external {
			case "any", "encoding/json.RawMessage":
				return wireShape{kind: kindAny}
			}
			return wireShape{kind: kindValue}
		}
		methods := decl.file.pkg.methods[decl.spec.Name.Name]
		if methods["MarshalJSON"] != nil || methods["MarshalText"] != nil {
			return wireShape{kind: kindCustom}
		}
		if _, isStruct := decl.spec.Type.(*ast.StructType); isStruct {
			return wireShape{kind: kindValue}
		}
		return s.shapeAt(typeRef{decl.spec.Type, decl.file}, depth+1)
	}
	return wireShape{kind: kindValue}
}

// structOf is the struct a declaration names, following `type A B` and
// aliases, or nil.
func (s *goSource) structOf(decl *typeDecl) (*ast.StructType, *goFile) {
	for range maxTypeDepth {
		if st, ok := decl.spec.Type.(*ast.StructType); ok {
			return st, decl.file
		}
		next, _ := s.lookup(typeRef{decl.spec.Type, decl.file})
		if next == nil {
			return nil, nil
		}
		decl = next
	}
	return nil, nil
}

// jsonField is one field of a struct as encoding/json writes it.
type jsonField struct {
	name      string
	omitEmpty bool
	ref       typeRef
	// swaggerType says the field's schema was given by hand, with a
	// swaggertype tag: it is left as it is.
	swaggerType bool
	depth       int
}

// jsonFields are a struct's fields as encoding/json writes them, those of the
// structs it embeds without a name promoted into it.
func (s *goSource) jsonFields(decl *typeDecl) []jsonField {
	byName := map[string]jsonField{}
	var order []string
	s.collectJSONFields(decl, 0, byName, &order)
	out := make([]jsonField, 0, len(order))
	for _, name := range order {
		out = append(out, byName[name])
	}
	return out
}

func (s *goSource) collectJSONFields(decl *typeDecl, depth int, byName map[string]jsonField, order *[]string) {
	st, file := s.structOf(decl)
	if st == nil || depth > maxTypeDepth {
		return
	}
	for _, field := range st.Fields.List {
		tag := fieldTag(field)
		name, opts, _ := strings.Cut(tag.Get("json"), ",")
		if name == "-" && opts == "" {
			continue
		}
		add := func(fieldName string) {
			if existing, found := byName[fieldName]; found && existing.depth <= depth {
				return
			}
			if _, found := byName[fieldName]; !found {
				*order = append(*order, fieldName)
			}
			byName[fieldName] = jsonField{
				name: fieldName, omitEmpty: hasOption(opts, "omitempty"), ref: typeRef{field.Type, file},
				swaggerType: tag.Get("swaggertype") != "", depth: depth,
			}
		}
		if len(field.Names) == 0 {
			embedded := field.Type
			if star, ok := embedded.(*ast.StarExpr); ok {
				embedded = star.X
			}
			if name == "" {
				if inner, _ := s.lookup(typeRef{embedded, file}); inner != nil {
					s.collectJSONFields(inner, depth+1, byName, order)
					continue
				}
				if !ast.IsExported(typeName(embedded)) {
					continue
				}
				add(typeName(embedded))
				continue
			}
			add(name)
			continue
		}
		for _, ident := range field.Names {
			if !ident.IsExported() {
				continue
			}
			if name != "" {
				add(name)
			} else {
				add(ident.Name)
			}
		}
	}
}

func fieldTag(field *ast.Field) reflect.StructTag {
	if field.Tag == nil {
		return ""
	}
	value, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		return ""
	}
	return reflect.StructTag(value)
}

func hasOption(opts, option string) bool {
	for _, opt := range strings.Split(opts, ",") {
		if opt == option {
			return true
		}
	}
	return false
}

func typeName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return e.Sel.Name
	}
	return ""
}

// moduleRoot finds the directory of go.mod from dir upwards, and the module's
// path.
func moduleRoot(dir string) (string, string, error) {
	for {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if rest, found := strings.CutPrefix(strings.TrimSpace(line), "module "); found {
					return dir, strings.TrimSpace(rest), nil
				}
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", os.ErrNotExist
		}
		dir = parent
	}
}

// method finds a method of a type, declared on it or promoted from a struct it
// embeds.
func (s *goSource) method(decl *typeDecl, name string, depth int) *funcDecl {
	if decl == nil || depth > maxTypeDepth {
		return nil
	}
	if fn := decl.file.pkg.methods[decl.spec.Name.Name][name]; fn != nil {
		return fn
	}
	st, file := s.structOf(decl)
	if st == nil {
		return nil
	}
	for _, field := range st.Fields.List {
		if len(field.Names) > 0 {
			continue
		}
		embedded := field.Type
		if star, ok := embedded.(*ast.StarExpr); ok {
			embedded = star.X
		}
		inner, _ := s.lookup(typeRef{embedded, file})
		if fn := s.method(inner, name, depth+1); fn != nil {
			return fn
		}
	}
	return nil
}
