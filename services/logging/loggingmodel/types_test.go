package loggingmodel

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// allowedImportPrefixes are the only hivepaas packages anything under
// services/logging may import. The rule is the point of this package: an
// implementation that reaches for entity or docker has moved application
// concerns into a layer that exists to be free of them.
var allowedImportPrefixes = []string{
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors",
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/",
	"github.com/hivepaas/hivepaas/services/logging",
}

func TestLoggingPackagesDoNotImportTheApplication(t *testing.T) {
	root := ".."

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		fset := token.NewFileSet()
		file, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if perr != nil {
			return perr
		}
		for _, imp := range file.Imports {
			p, uerr := strconv.Unquote(imp.Path.Value)
			if uerr != nil {
				return uerr
			}
			if !strings.HasPrefix(p, "github.com/hivepaas/hivepaas/") {
				continue // stdlib and third-party are fine
			}
			assert.True(t, allowed(p), "%s imports %s", path, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}

func allowed(path string) bool {
	for _, prefix := range allowedImportPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
