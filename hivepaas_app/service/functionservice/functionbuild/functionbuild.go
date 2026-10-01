// Package functionbuild writes what the build of a function needs: its inline
// code as files, and the Dockerfile HivePaaS writes for it on the runtime image
// of its release. See hivepaas/function-runtimes for the images, and the
// `hivepaas-runtime deps` and `build` commands the Dockerfile runs.
package functionbuild

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// DockerfilePath is where HivePaaS writes a function's Dockerfile, in the
// function's source: the one directory the function's own files may not use.
const DockerfilePath = base.FunctionReservedDir + "/Dockerfile"

// DockerfileReq is what a function's Dockerfile is written from.
type DockerfileReq struct {
	Source *entity.DeploymentFunctionSource
	// Images are the release's function runtime images, by runtime, as
	// base.ReleaseInfo.FunctionRuntimes has them.
	Images map[string]string
	// SourceDir is the function's source: its manifest, lock file and package
	// manager settings decide what the install step copies.
	SourceDir string
	// BuildArgs are the names of the build's variables passed as build
	// arguments, and BuildSecrets those passed as BuildKit secrets. The install
	// and the compilation see both.
	BuildArgs    []string
	BuildSecrets []string
}

// DockerfileResp is a function's Dockerfile, and what its build's log should
// say about it.
type DockerfileResp struct {
	Content string
	Notes   []string
}

// manifest is what a runtime installs a function's libraries from.
type manifest struct {
	// file is the manifest itself, without which there is nothing to install.
	file string
	// lock is the lock file the install makes when there is none.
	lock string
	// settings are files the package manager reads, such as a private registry
	// and its token.
	settings []string
	// localDeps says whether the manifest names libraries from the function's
	// own files, which the install step cannot see unless the code is copied
	// first.
	localDeps func(content string) bool
}

var manifests = map[base.FunctionRuntime]manifest{
	base.FunctionRuntimeNode24: {
		file: "package.json", lock: "package-lock.json", settings: []string{".npmrc"}, localDeps: npmLocalDeps,
	},
	base.FunctionRuntimePython313: {
		file: "requirements.txt", lock: "requirements.lock", settings: []string{"uv.toml"}, localDeps: pipLocalDeps,
	},
	base.FunctionRuntimeGo127: {file: "go.mod", lock: "go.sum", localDeps: goLocalDeps},
}

// Dockerfile writes the Dockerfile of a function. The runtime's image, and for
// a compiled runtime its build image, come from the release.
func Dockerfile(req *DockerfileReq) (*DockerfileResp, error) {
	src := req.Source
	m, ok := manifests[src.Runtime]
	if !ok {
		return nil, hperrors.Wrap(hperrors.ErrFunctionRuntimeUnavailable).WithParam("Runtime", src.Runtime)
	}
	runImage, err := runtimeImage(req.Images, string(src.Runtime))
	if err != nil {
		return nil, err
	}
	baseImage := runImage
	if src.Runtime.Compiled() {
		if baseImage, err = runtimeImage(req.Images, base.FunctionBuildImageKey(src.Runtime)); err != nil {
			return nil, err
		}
	}
	for _, pkg := range src.SystemPackages {
		if !base.FunctionSystemPackagePattern.MatchString(pkg) {
			return nil, hperrors.Wrap(hperrors.ErrValueInvalid).WithParam("Name", pkg)
		}
	}

	resp := &DockerfileResp{}
	w := &strings.Builder{}
	fmt.Fprintf(w, "# The Dockerfile HivePaaS writes for a function: runtime %s, contract %s.\n", src.Runtime,
		src.Contract)
	fmt.Fprintf(w, "FROM %s AS base\n", baseImage)
	writeSystemPackages(w, src.SystemPackages)

	// The libraries are installed in a stage of their own, from the manifest
	// alone, so that a change to the code alone installs nothing again. A
	// manifest that names libraries from the function's own files needs them
	// there: the code goes in first, whole.
	stage, codeCopied := "base", false
	manifestContent, hasManifest := readFile(req.SourceDir, m.file)
	if hasManifest {
		w.WriteString("\nFROM base AS deps\n")
		if m.localDeps(manifestContent) {
			w.WriteString("COPY --chown=hivepaas:hivepaas . /app/\n")
			codeCopied = true
			resp.Notes = append(resp.Notes, fmt.Sprintf("%s names libraries from the function's own files: "+
				"they are installed with its code, on every build.", m.file))
		} else {
			files := []string{m.file}
			for _, name := range append([]string{m.lock}, m.settings...) {
				if _, ok := readFile(req.SourceDir, name); ok {
					files = append(files, name)
				}
			}
			fmt.Fprintf(w, "COPY --chown=hivepaas:hivepaas %s /app/\n", strings.Join(files, " "))
		}
		if _, ok := readFile(req.SourceDir, m.lock); !ok {
			resp.Notes = append(resp.Notes, fmt.Sprintf("The function has no %s: this build resolves the "+
				"libraries' versions, and a later build may resolve others.", m.lock))
		}
		writeArgs(w, req.BuildArgs)
		writeRun(w, req.BuildSecrets, "hivepaas-runtime deps")
		stage = "deps"
	}

	if src.Runtime.Compiled() {
		fmt.Fprintf(w, "\nFROM %s AS build\n", stage)
		if !codeCopied {
			w.WriteString("COPY --chown=hivepaas:hivepaas . /app/\n")
		}
		writeEnv(w, entrypointEnv(src))
		writeArgs(w, req.BuildArgs)
		writeRun(w, req.BuildSecrets, "hivepaas-runtime build")

		fmt.Fprintf(w, "\nFROM %s\n", runImage)
		writeSystemPackages(w, src.SystemPackages)
		w.WriteString("COPY --from=build --chown=hivepaas:hivepaas /app /app\n")
	} else {
		fmt.Fprintf(w, "\nFROM %s\n", stage)
		if !codeCopied {
			w.WriteString("COPY --chown=hivepaas:hivepaas . /app/\n")
		}
	}
	writeEnv(w, append(entrypointEnv(src), limitsEnv(src)...))

	resp.Content = w.String()
	return resp, nil
}

func runtimeImage(images map[string]string, key string) (string, error) {
	image := images[key]
	if image == "" {
		return "", hperrors.Wrap(hperrors.ErrFunctionRuntimeUnavailable).WithParam("Runtime", key)
	}
	return image, nil
}

// writeSystemPackages installs the Debian packages, as root: the runtime's
// image runs as its own user.
func writeSystemPackages(w *strings.Builder, packages []string) {
	if len(packages) == 0 {
		return
	}
	w.WriteString("USER root\n")
	w.WriteString("RUN apt-get update \\\n")
	fmt.Fprintf(w, "    && apt-get install -y --no-install-recommends %s \\\n", strings.Join(packages, " "))
	w.WriteString("    && rm -rf /var/lib/apt/lists/*\n")
	w.WriteString("USER hivepaas\n")
}

// writeArgs declares the build arguments, which a RUN step sees only declared.
func writeArgs(w *strings.Builder, names []string) {
	for _, name := range slices.Sorted(slices.Values(names)) {
		fmt.Fprintf(w, "ARG %s\n", name)
	}
}

// writeRun runs command with every secret of the build mounted into its
// environment, where the image never records it.
func writeRun(w *strings.Builder, secrets []string, command string) {
	if len(secrets) == 0 {
		fmt.Fprintf(w, "RUN %s\n", command)
		return
	}
	w.WriteString("RUN")
	for _, name := range slices.Sorted(slices.Values(secrets)) {
		fmt.Fprintf(w, " --mount=type=secret,id=%s,env=%s", name, name)
	}
	fmt.Fprintf(w, " \\\n    %s\n", command)
}

func writeEnv(w *strings.Builder, env []string) {
	w.WriteString("ENV " + strings.Join(env, " \\\n    ") + "\n")
}

func entrypointEnv(src *entity.DeploymentFunctionSource) []string {
	return []string{
		"HP_FN_ENTRYPOINT=" + quote(src.Entrypoint.File),
		"HP_FN_HANDLER=" + quote(src.Entrypoint.Handler),
	}
}

// limitsEnv is the call's limits, as the runtime reads them.
func limitsEnv(src *entity.DeploymentFunctionSource) []string {
	return []string{
		fmt.Sprintf("HP_FN_TIMEOUT_MS=%d", time.Duration(src.Timeout).Milliseconds()),
		fmt.Sprintf("HP_FN_MAX_CONCURRENCY=%d", src.MaxConcurrency),
		fmt.Sprintf("HP_FN_MAX_BODY_SIZE=%d", src.MaxBodySize.Bytes()),
	}
}

// quote is a value of a Dockerfile's ENV, which reads "$" as a variable.
func quote(value string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`, `$`, `\$`).Replace(value) + `"`
}

func readFile(dir, name string) (string, bool) {
	content, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return "", false
	}
	return string(content), true
}

// npmLocalDeps says whether package.json names a library by a path, or a
// workspace of the function's own.
func npmLocalDeps(content string) bool {
	var pkg struct {
		Dependencies         map[string]string `json:"dependencies"`
		DevDependencies      map[string]string `json:"devDependencies"`
		OptionalDependencies map[string]string `json:"optionalDependencies"`
		Workspaces           any               `json:"workspaces"`
	}
	if json.Unmarshal([]byte(content), &pkg) != nil {
		return false // npm says what is wrong with it, in the install step
	}
	if pkg.Workspaces != nil {
		return true
	}
	for _, deps := range []map[string]string{pkg.Dependencies, pkg.DevDependencies, pkg.OptionalDependencies} {
		for _, spec := range deps {
			if strings.HasPrefix(spec, "file:") || strings.HasPrefix(spec, "link:") || localPath(spec) {
				return true
			}
		}
	}
	return false
}

// pipLocalDeps says whether requirements.txt names a library by a path, or reads
// another file of the function.
func pipLocalDeps(content string) bool {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if i := strings.Index(line, " #"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		switch {
		case line == "", strings.HasPrefix(line, "#"):
		case localPath(line), strings.HasPrefix(line, "file:"), strings.Contains(line, "@ file:"),
			strings.HasPrefix(line, "-e"), strings.HasPrefix(line, "--editable"),
			strings.HasPrefix(line, "-r"), strings.HasPrefix(line, "--requirement"),
			strings.HasPrefix(line, "-c"), strings.HasPrefix(line, "--constraint"):
			return true
		}
	}
	return false
}

// goReplaceByDir is a replace directive of go.mod whose target is a directory.
var goReplaceByDir = regexp.MustCompile(`=>\s*(\.\.?/|/)`)

func goLocalDeps(content string) bool {
	return goReplaceByDir.MatchString(content)
}

func localPath(spec string) bool {
	return strings.HasPrefix(spec, "./") || strings.HasPrefix(spec, "../") || strings.HasPrefix(spec, "/")
}

// WriteInlineCode writes a function's inline files into dir. A file whose path
// leaves dir, or lands in the directory HivePaaS keeps, is refused.
func WriteInlineCode(dir string, code *entity.FunctionInlineCode) error {
	root, err := filepath.Abs(dir)
	if err != nil {
		return hperrors.Wrap(err)
	}
	for _, f := range code.Files {
		clean := filepath.Clean(filepath.FromSlash(f.Path))
		target := filepath.Join(root, clean)
		if f.Path == "" || filepath.IsAbs(filepath.FromSlash(f.Path)) ||
			!strings.HasPrefix(target, root+string(filepath.Separator)) ||
			clean == base.FunctionReservedDir || strings.HasPrefix(clean, base.FunctionReservedDir+string(filepath.Separator)) {
			return hperrors.Wrap(hperrors.ErrValueInvalid).WithParam("Name", f.Path)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil { //nolint:mnd
			return hperrors.Wrap(err)
		}
		if err := os.WriteFile(target, []byte(f.Content), 0o644); err != nil { //nolint:mnd,gosec
			return hperrors.Wrap(err)
		}
	}
	return nil
}
