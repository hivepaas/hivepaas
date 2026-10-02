package appsettingsdto

import (
	"fmt"
	"path"
	"slices"
	"strings"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/githelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/vcsurl"
)

// DeploymentFunctionSourceReq is what a function is deployed from. Normalize it,
// then Validate it, then turn it into its entity.
type DeploymentFunctionSourceReq struct {
	Runtime        base.FunctionRuntime  `json:"runtime"`
	Contract       base.FunctionContract `json:"contract"`
	Entrypoint     FunctionEntrypointReq `json:"entrypoint"`
	Code           FunctionCodeReq       `json:"code"`
	SystemPackages []string              `json:"systemPackages"`
	Timeout        timeutil.Duration     `json:"timeout" swaggertype:"string"`
	MaxConcurrency int                   `json:"maxConcurrency"`
	MaxBodySize    unit.DataSize         `json:"maxBodySize" swaggertype:"string"`
	PushToRegistry basedto.ObjectIDReq   `json:"pushToRegistry"`
}

type FunctionEntrypointReq struct {
	File    string `json:"file"`
	Handler string `json:"handler"`
}

type FunctionCodeReq struct {
	Inline *FunctionInlineCodeReq `json:"inline"`
	Repo   *FunctionRepoCodeReq   `json:"repo"`
	Dir    string                 `json:"dir"`
}

type FunctionInlineCodeReq struct {
	Files []*FunctionFileReq `json:"files"`
}

type FunctionFileReq struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type FunctionRepoCodeReq struct {
	RepoType    base.RepoType            `json:"repoType"`
	RepoURL     string                   `json:"repoURL"`
	RepoRef     string                   `json:"repoRef"`
	CommitHash  string                   `json:"commitHash"`
	RepoOptions DeploymentRepoOptionsReq `json:"repoOptions"`
	Credentials basedto.ObjectIDReq      `json:"credentials"`
}

// Normalize folds the source into one spelling, and fills in what it leaves
// out with the runtime's defaults.
func (req *DeploymentFunctionSourceReq) Normalize() {
	if req == nil {
		return
	}
	req.Runtime = base.FunctionRuntime(strings.ToLower(strings.TrimSpace(string(req.Runtime))))
	req.Contract = base.FunctionContract(strings.TrimSpace(string(req.Contract)))
	if req.Contract == "" {
		req.Contract = base.FunctionContractV1
	}

	defaultFile, defaultHandler := req.Runtime.DefaultEntrypoint()
	req.Entrypoint.File = cleanFunctionPath(req.Entrypoint.File)
	if req.Entrypoint.File == "" {
		req.Entrypoint.File = defaultFile
	}
	req.Entrypoint.Handler = strings.TrimSpace(req.Entrypoint.Handler)
	if req.Entrypoint.Handler == "" {
		req.Entrypoint.Handler = defaultHandler
	}

	req.Code.Dir = cleanFunctionPath(req.Code.Dir)
	if req.Code.Dir == "." {
		req.Code.Dir = ""
	}
	req.Code.Inline.Normalize()
	if req.Code.Repo != nil {
		req.Code.Repo.RepoURL = strings.TrimSpace(req.Code.Repo.RepoURL)
		req.Code.Repo.RepoRef = strings.TrimSpace(req.Code.Repo.RepoRef)
	}

	packages := make([]string, 0, len(req.SystemPackages))
	for _, pkg := range req.SystemPackages {
		if pkg = strings.TrimSpace(pkg); pkg != "" && !slices.Contains(packages, pkg) {
			packages = append(packages, pkg)
		}
	}
	req.SystemPackages = packages

	if req.Timeout == 0 {
		req.Timeout = timeutil.Duration(base.FunctionTimeoutDefault)
	}
	if req.MaxConcurrency == 0 {
		req.MaxConcurrency = base.FunctionMaxConcurrencyDefault
	}
	if req.MaxBodySize == 0 {
		req.MaxBodySize = base.FunctionMaxBodySizeDefault
	}
}

// cleanFunctionPath trims a path and cleans it: "./src//index.js" is
// "src/index.js". A path that leaves the function stays one, for Validate.
func cleanFunctionPath(p string) string {
	if p = strings.TrimSpace(p); p == "" {
		return ""
	}
	return path.Clean(p)
}

// Validate checks the source, once normalized. field is where it is in the
// request.
func (req *DeploymentFunctionSourceReq) Validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStrIn(&req.Runtime, true, base.AllFunctionRuntimes, field+"runtime")...)
	res = append(res, basedto.ValidateStrIn(&req.Contract, true, base.AllFunctionContracts, field+"contract")...)
	res = append(res, req.validateEntrypoint(field+"entrypoint.")...)
	res = append(res, req.Code.validate(field+"code")...)

	res = append(res, basedto.ValidateSliceEx(req.SystemPackages, true, 0, base.FunctionSystemPackagesMax, nil,
		field+"systemPackages")...)
	for i, pkg := range req.SystemPackages {
		res = append(res, vld.Must(base.FunctionSystemPackagePattern.MatchString(pkg)).OnError(
			vld.SetField(fmt.Sprintf("%ssystemPackages[%d]", field, i), nil),
			vld.SetCustomKey("ERR_VLD_DEBIAN_PACKAGE_INVALID"),
		))
	}

	res = append(res, basedto.ValidateDuration(&req.Timeout, true, timeutil.Duration(base.FunctionTimeoutMin),
		timeutil.Duration(base.FunctionTimeoutMax), field+"timeout")...)
	res = append(res, basedto.ValidateNumber(&req.MaxConcurrency, true, 1, base.FunctionMaxConcurrencyMax,
		field+"maxConcurrency")...)
	res = append(res, basedto.ValidateNumber(&req.MaxBodySize, true, base.FunctionMaxBodySizeMin,
		base.FunctionMaxBodySizeMax, field+"maxBodySize")...)
	res = append(res, basedto.ValidateObjectIDReq(&req.PushToRegistry, false, field+"pushToRegistry")...)
	return res
}

func (req *DeploymentFunctionSourceReq) validateEntrypoint(field string) (res []vld.Validator) {
	file, handler := req.Entrypoint.File, req.Entrypoint.Handler
	fileOK := base.FunctionPathOK(file, true)
	if exts, ok := base.FunctionEntrypointExts[req.Runtime]; ok {
		fileOK = base.FunctionPathOK(file, false) && slices.Contains(exts, path.Ext(file))
	}
	res = append(res, vld.Must(fileOK).OnError(
		vld.SetField(field+"file", nil),
		vld.SetCustomKey("ERR_VLD_FUNCTION_ENTRYPOINT_INVALID"),
	))
	if handlerRegex, ok := base.FunctionHandlerPatterns[req.Runtime]; ok {
		res = append(res, vld.Must(handlerRegex.MatchString(handler)).OnError(
			vld.SetField(field+"handler", nil),
			vld.SetCustomKey("ERR_VLD_FUNCTION_HANDLER_INVALID"),
		))
	}
	return res
}

func (req *FunctionCodeReq) validate(field string) (res []vld.Validator) {
	res = append(res, vld.Must((req.Inline != nil) != (req.Repo != nil)).OnError(
		vld.SetField(field, nil),
		vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED_ONLY"),
	))
	if req.Inline != nil {
		res = append(res, req.Inline.validate(field+".inline")...)
		// A directory picks a function among several in a repository.
		res = append(res, vld.Must(req.Dir == "").OnError(
			vld.SetField(field+".dir", nil),
			vld.SetCustomKey("ERR_VLD_FIELD_UNALLOWED"),
		))
	}
	if req.Repo != nil {
		res = append(res, req.Repo.validate(field+".repo.")...)
		if req.Dir != "" {
			res = append(res, vld.Must(base.FunctionPathOK(req.Dir, false)).OnError(
				vld.SetField(field+".dir", nil),
				vld.SetCustomKey("ERR_VLD_FUNCTION_FILE_PATH_INVALID"),
			))
		}
	}
	return res
}

// Normalize cleans the files' paths: "./src//index.js" is "src/index.js".
func (req *FunctionInlineCodeReq) Normalize() {
	if req == nil {
		return
	}
	req.Files = gofn.Filter(req.Files, func(f *FunctionFileReq) bool { return f != nil })
	for _, f := range req.Files {
		f.Path = cleanFunctionPath(f.Path)
	}
}

// Validate checks inline code against its limits, field naming the code.
func (req *FunctionInlineCodeReq) Validate(field string) []vld.Validator {
	return req.validate(field)
}

// ToEntity is the code as it is kept, once normalized and validated.
func (req *FunctionInlineCodeReq) ToEntity() *entity.FunctionInlineCode {
	return &entity.FunctionInlineCode{
		Files: gofn.MapSlice(req.Files, func(f *FunctionFileReq) *entity.FunctionFile {
			return &entity.FunctionFile{Path: f.Path, Content: f.Content}
		}),
	}
}

func (req *FunctionInlineCodeReq) validate(field string) (res []vld.Validator) {
	files := req.Files
	res = append(res, basedto.ValidateSliceEx(files, false, 1, base.FunctionInlineCodeMaxFiles, nil,
		field+".files")...)
	res = append(res, basedto.ValidateObjectSliceBy(files, false, true, nil,
		func(f *FunctionFileReq) string { return f.Path }, field+".files")...)

	size := unit.DataSize(0)
	for i, f := range files {
		size += unit.DataSize(len(f.Content))
		pathField := fmt.Sprintf("%s.files[%d].path", field, i)
		if !base.FunctionPathOK(f.Path, false) {
			res = append(res, vld.Must(false).OnError(
				vld.SetField(pathField, nil),
				vld.SetCustomKey("ERR_VLD_FUNCTION_FILE_PATH_INVALID"),
			))
			continue
		}
		res = append(res, vld.Must(!base.FunctionPathReserved(f.Path)).OnError(
			vld.SetField(pathField, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_RESERVED"),
			vld.SetParam("Value", base.FunctionReservedDir),
		))
	}
	res = append(res, vld.Must(size <= base.FunctionInlineCodeMaxSize).OnError(
		vld.SetField(field, nil),
		vld.SetCustomKey("ERR_VLD_FUNCTION_CODE_TOO_LARGE"),
		vld.SetParam("Max", base.FunctionInlineCodeMaxSize.HR()),
		vld.SetParam("Actual", size.HR()),
	))
	return res
}

func (req *FunctionRepoCodeReq) validate(field string) (res []vld.Validator) {
	res = append(res, basedto.ValidateStrIn(&req.RepoType, true, base.AllRepoTypes, field+"repoType")...)
	res = append(res, basedto.ValidateGitRepoURL(&req.RepoURL, true, field+"repoURL")...)
	res = append(res, basedto.ValidateStr(&req.RepoRef, false, 1, repoRefMaxLen, field+"repoRef")...)
	res = append(res, basedto.ValidateGitCommitHash(&req.CommitHash, false, field+"commitHash")...)
	res = append(res, basedto.ValidateObjectIDReq(&req.Credentials, false, field+"credentials")...)
	return res
}

// ToEntity is the source as it is stored, once normalized and validated.
func (req *DeploymentFunctionSourceReq) ToEntity() (*entity.DeploymentFunctionSource, error) {
	if req == nil {
		return nil, nil
	}
	code := entity.FunctionCode{Dir: req.Code.Dir}
	if req.Code.Inline != nil {
		code.Inline = req.Code.Inline.ToEntity()
	}
	if req.Code.Repo != nil {
		repo, err := req.Code.Repo.toEntity()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		code.Repo = repo
	}
	var packages []string
	if len(req.SystemPackages) > 0 {
		packages = slices.Clone(req.SystemPackages)
	}
	return &entity.DeploymentFunctionSource{
		Runtime:        req.Runtime,
		Contract:       req.Contract,
		Entrypoint:     entity.FunctionEntrypoint{File: req.Entrypoint.File, Handler: req.Entrypoint.Handler},
		Code:           code,
		SystemPackages: packages,
		Timeout:        req.Timeout,
		MaxConcurrency: req.MaxConcurrency,
		MaxBodySize:    req.MaxBodySize,
		PushToRegistry: entity.ObjectID{ID: req.PushToRegistry.ID},
	}, nil
}

func (req *FunctionRepoCodeReq) toEntity() (*entity.FunctionRepoCode, error) {
	repoRef := req.RepoRef
	switch req.RepoType { //nolint:gocritic
	case base.RepoTypeGit:
		repoRef = string(githelper.NormalizeRepoRef(repoRef))
	}
	parsedRepoURL, err := vcsurl.Parse(req.RepoURL)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &entity.FunctionRepoCode{
		RepoType:    req.RepoType,
		RepoID:      parsedRepoURL.ID,
		RepoURL:     req.RepoURL,
		RepoRef:     repoRef,
		CommitHash:  req.CommitHash,
		RepoOptions: req.RepoOptions.ToEntity(),
		Credentials: entity.RepoCredentials{ID: req.Credentials.ID},
	}, nil
}
