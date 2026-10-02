package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

// DeploymentFunctionSource is what a function is deployed from: its code, the
// runtime that runs it, and the limits of one call. HivePaaS writes the
// function's Dockerfile from it.
type DeploymentFunctionSource struct {
	// Runtime is the language and its line: node24, bun1, python313, go127.
	Runtime base.FunctionRuntime `json:"runtime"`
	// Contract is the version of the handler contract: v1.
	Contract base.FunctionContract `json:"contract"`
	// Entrypoint is the file and the name of the handler in it.
	Entrypoint FunctionEntrypoint `json:"entrypoint"`
	// Code is inline, as written in the dashboard's editor, or a repository.
	Code FunctionCode `json:"code"`
	// SystemPackages are Debian packages installed before the libraries:
	// "libpq-dev", or "ffmpeg=7:5.1.6-0+deb12u1".
	SystemPackages []string `json:"systemPackages,omitempty"`
	// Limits of one call.
	Timeout        timeutil.Duration `json:"timeout"`
	MaxConcurrency int               `json:"maxConcurrency"`
	MaxBodySize    unit.DataSize     `json:"maxBodySize"`
	// PushToRegistry is the registry the built image goes to, which a cluster of
	// several nodes needs.
	PushToRegistry ObjectID `json:"pushToRegistry,omitzero"`
}

// FunctionEntrypoint is where the runtime finds the handler.
type FunctionEntrypoint struct {
	// File is the handler's file, relative to the function's root; for Go, the
	// directory of the handler's package.
	File string `json:"file"`
	// Handler is the name of the handler in it.
	Handler string `json:"handler"`
}

// FunctionCode is where a function's code is: inline or in a repository.
type FunctionCode struct {
	Inline *FunctionInlineCode `json:"inline,omitempty"`
	Repo   *FunctionRepoCode   `json:"repo,omitempty"`
	// Dir is where the function is in the repository, for a repository holding
	// several.
	Dir string `json:"dir,omitempty"`
}

// FunctionInlineCode is code kept in the setting itself, so that a deployment's
// snapshot of its settings holds the code it ran.
type FunctionInlineCode struct {
	Files []*FunctionFile `json:"files"`
}

// FunctionFile is one file of inline code, by its path in the function.
type FunctionFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// FunctionRepoCode is a repository holding a function's code.
type FunctionRepoCode struct {
	RepoType    base.RepoType         `json:"repoType"`
	RepoID      string                `json:"repoId"`
	RepoURL     string                `json:"repoURL"`
	RepoRef     string                `json:"repoRef"` // can be branch name, tag...
	CommitHash  string                `json:"commitHash,omitempty"`
	RepoOptions DeploymentRepoOptions `json:"repoOptions"`
	Credentials RepoCredentials       `json:"credentials,omitzero"`
}

// RepoSource is the repository as a checkout takes it.
func (c *FunctionRepoCode) RepoSource() *DeploymentRepoSource {
	return &DeploymentRepoSource{
		RepoType:    c.RepoType,
		RepoID:      c.RepoID,
		RepoURL:     c.RepoURL,
		RepoRef:     c.RepoRef,
		CommitHash:  c.CommitHash,
		RepoOptions: c.RepoOptions,
		Credentials: c.Credentials,
	}
}
