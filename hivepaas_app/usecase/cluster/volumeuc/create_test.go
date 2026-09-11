package volumeuc

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// CreateVolume must never assign to req.Name. The settings framework checks
// VerifyingName - a copy of req.Name taken before PrepareCreation runs - and
// then persists whatever req.Name holds by the time the PrepareCreation
// closure sets pData.Setting.Name. If the two diverge, the duplicate-name
// check compares one string while a different one lands in the database,
// silently defeating it.
//
// This file's CreateVolume once did exactly that: it prefixed req.Name with
// the project key inside PrepareCreation, after VerifyingName had already
// been evaluated against the unprefixed name. Creating "pgdata" twice in the
// same project stored "X_pgdata" the first time and then found nothing named
// "pgdata" the second time, so the collision went uncaught. The fix removed
// the prefix outright - the docker name is now the setting's own id
// (pData.Setting.RefID = pData.Setting.ID), so nothing needs the project key
// to keep two projects' volumes apart in a shared docker namespace any more.
//
// A behavioral test of this would have to drive uc.CreateSetting through a
// real database transaction (settings.BaseUC.DB is a concrete *database.DB,
// and CreateSetting also touches ScopeService, SettingRepo,
// SettingEventService and AuditService along the way) - there is no fake or
// in-memory DB harness for that anywhere in this codebase, and building one
// for this one assignment would be exactly the elaborate harness not worth
// building. Parsing the function's own source and asserting it contains no
// assignment to req.Name is the narrower thing actually reachable, and it
// catches the regression this test is named for.
func TestCreateVolumeNeverMutatesReqName(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "create.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	var fn *ast.FuncDecl
	for _, decl := range file.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok && fd.Name.Name == "CreateVolume" {
			fn = fd
			break
		}
	}
	if fn == nil {
		t.Fatal("CreateVolume not found in create.go")
	}

	ast.Inspect(fn, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, lhs := range assign.Lhs {
			sel, ok := lhs.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "req" && sel.Sel.Name == "Name" {
				t.Errorf("CreateVolume must not assign to req.Name (found at %s): doing so after "+
					"VerifyingName is computed lets the stored name diverge from the name the "+
					"duplicate-name check actually verified",
					fset.Position(assign.Pos()))
			}
		}
		return true
	})
}
