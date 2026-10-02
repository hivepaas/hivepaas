package specserviceimpl

import (
	"fmt"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/githelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/vcsurl"
)

// normalizeFunctionSource spells an imported function's source the way
// creating a function does (appsettingsdto.DeploymentFunctionSourceReq's
// Normalize), and fills what it leaves out with its runtime's defaults. A
// service does not import a DTO, so the two are held together by a test.
func normalizeFunctionSource(src *entity.DeploymentFunctionSource) {
	if src == nil {
		return
	}
	src.Runtime = base.FunctionRuntime(strings.ToLower(strings.TrimSpace(string(src.Runtime))))
	src.Contract = base.FunctionContract(strings.TrimSpace(string(src.Contract)))
	if src.Contract == "" {
		src.Contract = base.FunctionContractV1
	}

	defaultFile, defaultHandler := src.Runtime.DefaultEntrypoint()
	src.Entrypoint.File = cleanFunctionPath(src.Entrypoint.File)
	if src.Entrypoint.File == "" {
		src.Entrypoint.File = defaultFile
	}
	src.Entrypoint.Handler = strings.TrimSpace(src.Entrypoint.Handler)
	if src.Entrypoint.Handler == "" {
		src.Entrypoint.Handler = defaultHandler
	}

	src.Code.Dir = cleanFunctionPath(src.Code.Dir)
	if src.Code.Dir == "." {
		src.Code.Dir = ""
	}
	if inline := src.Code.Inline; inline != nil {
		inline.Files = slices.DeleteFunc(inline.Files, func(f *entity.FunctionFile) bool { return f == nil })
		for _, f := range inline.Files {
			f.Path = cleanFunctionPath(f.Path)
		}
	}
	if repo := src.Code.Repo; repo != nil {
		repo.RepoURL = strings.TrimSpace(repo.RepoURL)
		repo.RepoRef = strings.TrimSpace(repo.RepoRef)
		if repo.RepoType == base.RepoTypeGit {
			repo.RepoRef = string(githelper.NormalizeRepoRef(repo.RepoRef))
		}
		if parsed, err := vcsurl.Parse(repo.RepoURL); err == nil {
			repo.RepoID = parsed.ID
		}
	}

	packages := make([]string, 0, len(src.SystemPackages))
	for _, pkg := range src.SystemPackages {
		if pkg = strings.TrimSpace(pkg); pkg != "" && !slices.Contains(packages, pkg) {
			packages = append(packages, pkg)
		}
	}
	src.SystemPackages = packages
	if len(src.SystemPackages) == 0 {
		src.SystemPackages = nil
	}

	if src.Timeout == 0 {
		src.Timeout = timeutil.Duration(base.FunctionTimeoutDefault)
	}
	if src.MaxConcurrency == 0 {
		src.MaxConcurrency = base.FunctionMaxConcurrencyDefault
	}
	if src.MaxBodySize == 0 {
		src.MaxBodySize = base.FunctionMaxBodySizeDefault
	}
}

// cleanFunctionPath trims a path and cleans it: "./src//index.js" is
// "src/index.js". A path that leaves the function stays one, for the check.
func cleanFunctionPath(p string) string {
	if p = strings.TrimSpace(p); p == "" {
		return ""
	}
	return path.Clean(p)
}

// functionSourceProblems says what is wrong with a normalized function source,
// field by field, as creating a function refuses it; nothing for a valid one.
func functionSourceProblems(src *entity.DeploymentFunctionSource) []string {
	var problems []string
	add := func(field, format string, args ...any) {
		problems = append(problems, field+": "+fmt.Sprintf(format, args...))
	}
	if !slices.Contains(base.AllFunctionRuntimes, src.Runtime) {
		add("runtime", "%q is not one of %v", src.Runtime, base.AllFunctionRuntimes)
	}
	if !slices.Contains(base.AllFunctionContracts, src.Contract) {
		add("contract", "%q is not one of %v", src.Contract, base.AllFunctionContracts)
	}
	problems = append(problems, entrypointProblems(src)...)
	problems = append(problems, codeProblems(&src.Code)...)

	if len(src.SystemPackages) > base.FunctionSystemPackagesMax {
		add("systemPackages", "at most %d", base.FunctionSystemPackagesMax)
	}
	for _, pkg := range src.SystemPackages {
		if !base.FunctionSystemPackagePattern.MatchString(pkg) {
			add("systemPackages", "%q is not a Debian package", pkg)
		}
	}
	timeout := time.Duration(src.Timeout)
	if timeout < base.FunctionTimeoutMin || timeout > base.FunctionTimeoutMax {
		add("timeout", "%s is not between %s and %s", timeout, base.FunctionTimeoutMin, base.FunctionTimeoutMax)
	}
	if src.MaxConcurrency < 1 || src.MaxConcurrency > base.FunctionMaxConcurrencyMax {
		add("maxConcurrency", "%d is not between 1 and %d", src.MaxConcurrency, base.FunctionMaxConcurrencyMax)
	}
	if src.MaxBodySize < base.FunctionMaxBodySizeMin || src.MaxBodySize > base.FunctionMaxBodySizeMax {
		add("maxBodySize", "%s is not between %s and %s", src.MaxBodySize.HR(),
			base.FunctionMaxBodySizeMin.HR(), base.FunctionMaxBodySizeMax.HR())
	}
	return problems
}

func entrypointProblems(src *entity.DeploymentFunctionSource) []string {
	var problems []string
	file := src.Entrypoint.File
	fileOK := base.FunctionPathOK(file, true)
	if exts, ok := base.FunctionEntrypointExts[src.Runtime]; ok {
		fileOK = base.FunctionPathOK(file, false) && slices.Contains(exts, path.Ext(file))
	}
	if !fileOK {
		problems = append(problems, fmt.Sprintf("entrypoint.file: %q is not a handler's file for %s", file, src.Runtime))
	}
	if pattern, ok := base.FunctionHandlerPatterns[src.Runtime]; ok && !pattern.MatchString(src.Entrypoint.Handler) {
		problems = append(problems, fmt.Sprintf("entrypoint.handler: %q is not a name %s calls",
			src.Entrypoint.Handler, src.Runtime))
	}
	return problems
}

func codeProblems(code *entity.FunctionCode) []string {
	if (code.Inline != nil) == (code.Repo != nil) {
		return []string{"code: inline code or a repository, one of them"}
	}
	var problems []string
	if repo := code.Repo; repo != nil {
		if !slices.Contains(base.AllRepoTypes, repo.RepoType) {
			problems = append(problems, fmt.Sprintf("code.repo.repoType: %q is not one of %v",
				repo.RepoType, base.AllRepoTypes))
		}
		if _, err := vcsurl.Parse(repo.RepoURL); err != nil || repo.RepoURL == "" {
			problems = append(problems, fmt.Sprintf("code.repo.repoURL: %q is not a repository's URL", repo.RepoURL))
		}
		if code.Dir != "" && !base.FunctionPathOK(code.Dir, false) {
			problems = append(problems, fmt.Sprintf("code.dir: %q is not a directory of the repository", code.Dir))
		}
		return problems
	}

	files := code.Inline.Files
	if code.Dir != "" {
		problems = append(problems, "code.dir: inline code is in no directory")
	}
	if len(files) == 0 || len(files) > base.FunctionInlineCodeMaxFiles {
		problems = append(problems, fmt.Sprintf("code.inline.files: from 1 to %d files, not %d",
			base.FunctionInlineCodeMaxFiles, len(files)))
	}
	seen := map[string]bool{}
	size := unit.DataSize(0)
	for i, f := range files {
		size += unit.DataSize(len(f.Content))
		switch {
		case !base.FunctionPathOK(f.Path, false) || base.FunctionPathReserved(f.Path):
			problems = append(problems, fmt.Sprintf("code.inline.files[%d].path: no function may have a file at %q",
				i, f.Path))
		case seen[f.Path]:
			problems = append(problems, fmt.Sprintf("code.inline.files[%d].path: %q is there twice", i, f.Path))
		}
		seen[f.Path] = true
	}
	if size > base.FunctionInlineCodeMaxSize {
		problems = append(problems, fmt.Sprintf("code.inline: %s of code, more than %s", size.HR(),
			base.FunctionInlineCodeMaxSize.HR()))
	}
	return problems
}
