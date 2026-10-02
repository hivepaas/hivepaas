package specserviceimpl

import (
	"maps"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const (
	// functionsSegment follows an env's path in the path of its functions' code.
	functionsSegment = "functions"
	// filesFromKey names, in a function's inline code, the bundle directory its
	// files are in, in place of the files.
	filesFromKey = "filesFrom"
)

// functionDir is where the code of the function keyed appKey, in the env at
// envPath, is in a bundle: projects/<project>/envs/<env>/functions/<app>.
func functionDir(envPath, appKey string) string {
	return envPath + "/" + functionsSegment + "/" + appKey
}

// inlineCodeOf is the inline code of an app's function source, as the
// document holds it; nil for an app that is no function, or whose code is in a
// repository.
func inlineCodeOf(doc *specmodel.AppDoc) map[string]any {
	if doc == nil || doc.Deployment == nil {
		return nil
	}
	source, _ := doc.Deployment.Source["functionSource"].(map[string]any)
	code, _ := source["code"].(map[string]any)
	inline, _ := code["inline"].(map[string]any)
	return inline
}

// writeFunctionCode takes the inline code of each function of an env out of
// its document: each file becomes a file of the bundle under the function's
// directory, which its source names instead. Everything before it - the
// report, the references, an import's comparison - saw the code inline.
func writeFunctionCode(bundle *specmodel.Bundle, envPath string, apps map[string]*specmodel.AppDoc) {
	for _, key := range slices.Sorted(maps.Keys(apps)) {
		inline := inlineCodeOf(apps[key])
		if inline == nil {
			continue
		}
		dir := functionDir(envPath, key)
		files, _ := inline["files"].([]any)
		for _, file := range files {
			entry, _ := file.(map[string]any)
			path, _ := entry["path"].(string)
			content, _ := entry["content"].(string)
			bundle.Files[dir+"/"+path] = []byte(content)
		}
		delete(inline, "files")
		inline[filesFromKey] = dir
	}
}

// readFunctionCode gives each function of a bundle the files its source names
// back, inline, so that what reads the bundle next sees the code as export
// built it. A source naming anything but its own directory, a directory with
// no file, a file no function may have and a file that is not text make the
// bundle unreadable. A file under functions/ that no source names is left
// alone, as any file the reader does not know.
func readFunctionCode(bundle *specmodel.ImportBundle, files map[string][]byte) error {
	for _, project := range slices.Sorted(maps.Keys(bundle.Envs)) {
		envs := bundle.Envs[project]
		for _, env := range slices.Sorted(maps.Keys(envs)) {
			envPath := projectsSegment + "/" + project + "/" + envsSegment + "/" + env
			for _, key := range slices.Sorted(maps.Keys(envs[env].Apps)) {
				if err := readOneFunctionCode(envPath, key, envs[env].Apps[key], files); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func readOneFunctionCode(envPath, key string, doc *specmodel.AppDoc, files map[string][]byte) error {
	inline := inlineCodeOf(doc)
	from, named := inline[filesFromKey]
	if !named {
		return nil
	}
	dir := functionDir(envPath, key)
	if from != dir {
		return invalidBundle("%s/apps/%s: its code is read from %v, which is not its own directory %s",
			envPath, key, from, dir)
	}
	if _, listed := inline["files"]; listed {
		return invalidBundle("%s/apps/%s: its code is both listed and read from %s", envPath, key, dir)
	}

	prefix := dir + "/"
	var codeFiles []any
	for _, name := range slices.Sorted(maps.Keys(files)) {
		path, inDir := strings.CutPrefix(name, prefix)
		if !inDir {
			continue
		}
		if !base.FunctionPathOK(path, false) || base.FunctionPathReserved(path) {
			return invalidBundle("%s: no function may have a file at %q", dir, path)
		}
		content := files[name]
		if !utf8.Valid(content) {
			return invalidBundle("%s: %s is not UTF-8 text", dir, path)
		}
		codeFiles = append(codeFiles, map[string]any{"path": path, "content": string(content)})
	}
	if len(codeFiles) == 0 {
		return invalidBundle("%s: the function's code holds no file", dir)
	}
	delete(inline, filesFromKey)
	inline["files"] = codeFiles
	return nil
}
