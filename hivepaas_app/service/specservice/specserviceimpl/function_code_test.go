package specserviceimpl

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const helloDir = "projects/project_a/envs/dev/functions/hello"

// inlineFilesOf is the inline code a function's source holds, as a bundle reads
// it back: its files' paths and contents.
func inlineFilesOf(t *testing.T, doc *specmodel.AppDoc) map[string]string {
	t.Helper()
	inline := inlineCodeOf(doc)
	if !assert.NotNil(t, inline, "the source holds inline code") {
		return nil
	}
	assert.NotContains(t, inline, filesFromKey, "the directory is replaced by the files")
	files, _ := inline["files"].([]any)
	out := map[string]string{}
	for _, file := range files {
		entry, _ := file.(map[string]any)
		path, _ := entry["path"].(string)
		content, _ := entry["content"].(string)
		out[path] = content
	}
	return out
}

// A bundle read back gives the function's source with its code inline again,
// as export built it.
func TestAFunctionsCodeIsReadBackFromItsFiles(t *testing.T) {
	bundle, err := readBundle(exportedBytes(t, specmodel.SecretsModeOmit, ""), "")
	if !assert.NoError(t, err) {
		return
	}

	hello := bundle.Envs["project_a"]["dev"].Apps["hello"]
	assert.Equal(t, map[string]string{
		"index.js":    "export default () => ({})",
		"lib/util.js": "export const two = 2;\n",
	}, inlineFilesOf(t, hello))
}

// exportedFiles are the payload files of the fixture's export, and the
// manifest, to change before reading them back.
func exportedFiles(t *testing.T) map[string][]byte {
	t.Helper()
	svc, _ := exportFixture(t).(*service)
	built, err := svc.buildBundle(context.Background(), nil, exportReqOmit())
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	manifest, err := marshalDoc(built.Manifest)
	assert.NoError(t, err)
	files := map[string][]byte{manifestFilename: manifest}
	for name, content := range built.Files {
		files[name] = content
	}
	return files
}

func assertBundleRefused(t *testing.T, files map[string][]byte, why string) {
	t.Helper()
	_, err := parseBundleFiles(files)
	assert.True(t, errors.Is(err, hperrors.ErrSpecBundleInvalid), "%s: got %v", why, err)
}

func TestABundleWhoseFunctionNamesAnotherDirectoryIsRefused(t *testing.T) {
	files := exportedFiles(t)
	env := string(files["projects/project_a/envs/dev.yaml"])
	files["projects/project_a/envs/dev.yaml"] = []byte(replaceOnce(t, env,
		"filesFrom: "+helloDir, "filesFrom: projects/project_a/envs/dev/functions/other"))

	assertBundleRefused(t, files, "a function reads its own directory only")
}

func TestABundleWhoseFunctionHasNoFileIsRefused(t *testing.T) {
	files := exportedFiles(t)
	delete(files, helloDir+"/index.js")
	delete(files, helloDir+"/lib/util.js")

	assertBundleRefused(t, files, "a function without code")
}

func TestABundleWhoseFunctionHasAFileNoFunctionMayHaveIsRefused(t *testing.T) {
	for _, path := range []string{".hivepaas/Dockerfile", "a b.js", "lib/../x.js", "lib//x.js"} {
		files := exportedFiles(t)
		files[helloDir+"/"+path] = []byte("x")

		assertBundleRefused(t, files, path)
	}
}

func TestABundleWhoseFunctionHasAFileThatIsNotTextIsRefused(t *testing.T) {
	files := exportedFiles(t)
	files[helloDir+"/blob.bin"] = []byte{0xff, 0xfe, 0x00}

	assertBundleRefused(t, files, "code is UTF-8 text")
}

// A file under functions/ that no source names is left alone, as any file the
// reader does not know.
func TestAFunctionsDirectoryNobodyNamesIsIgnored(t *testing.T) {
	files := exportedFiles(t)
	files["projects/project_a/envs/dev/functions/gone/index.js"] = []byte("x")

	_, err := parseBundleFiles(files)
	assert.NoError(t, err)
}

func exportReqOmit() *specservice.ExportReq {
	return &specservice.ExportReq{Scope: entity.NewObjectScopeGlobal(), SecretsMode: specmodel.SecretsModeOmit}
}

func replaceOnce(t *testing.T, s, old, replacement string) string {
	t.Helper()
	assert.Equal(t, 1, strings.Count(s, old), "%q is in the document once", old)
	return strings.Replace(s, old, replacement, 1)
}

// Exported and imported back where it came from, a function plans no change:
// both sides read its code the same way.
func TestAnExportedFunctionImportedWhereItCameFromChangesNothing(t *testing.T) {
	svc, bundle := planFixture(t)

	out := plan(t, svc, bundle, specmodel.ImportOptions{})

	hello := node(t, out, "projects/project_a/envs/dev/apps/hello")
	assert.Equal(t, specmodel.ActionUnchanged, hello.Action, "changes: %v", hello.Changes)
}
