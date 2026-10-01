// Package functiontest runs a function's test run on the node it is given:
// the handler called once, with code not yet saved, in a throwaway container
// started from the function's libraries' image. See part 3 of
// docs/superpowers/specs/2026-10-01-functions-design.md.
package functiontest

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice/functioninvoke"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
)

const (
	// ResultMarker starts the last line invoke writes: its result.
	ResultMarker = functioninvoke.ResultMarker
	// RuntimeUID is the runtime images' user, who owns /app: the code a test
	// run copies in is that user's, since npm and go mod tidy write there.
	RuntimeUID = 10001
	// BodyMax and LogsMax are how much of a response's body and of a call's
	// log come back; the rest is cut.
	BodyMax = unit.MB
	LogsMax = unit.MB

	// requestPath is where the request goes in the container, for invoke to
	// read on its standard input.
	requestPath = "/tmp/hivepaas-request.json"
	// killMargin is how long a container may outlive the call's timeout, to
	// start and to write its result; compiledKillMargin is a compiled runtime's,
	// which builds the function first and may download its modules.
	killMargin         = 30 * time.Second
	compiledKillMargin = 5 * time.Minute
)

// Outcome is how a test run ended: as the runtime says, when it wrote a result,
// or else as HivePaaS saw it.
type Outcome string

const (
	OutcomeOK      Outcome = "ok"
	OutcomeError   Outcome = "error"
	OutcomeTimeout Outcome = "timeout"
	// OutcomeLibrariesFailed is an install that failed: there is no container.
	OutcomeLibrariesFailed Outcome = "libraries-failed"
	// OutcomeNotLoaded is a handler that could not be loaded, or built.
	OutcomeNotLoaded Outcome = "not-loaded"
	// OutcomeBadRequest is a request the runtime could not read.
	OutcomeBadRequest Outcome = "bad-request"
	// OutcomeKilled is a container that outlived the call's timeout and its
	// margin.
	OutcomeKilled Outcome = "killed"
	// OutcomeNoResult is a container that ended without a result: killed for
	// its memory, say. Its exit code says how.
	OutcomeNoResult Outcome = "no-result"
)

// Request is the request a test run calls the handler with, as invoke reads it.
type Request = functioninvoke.Request

// RunReq is a test run on the node it runs on.
type RunReq struct {
	// Source is the function's settings: its runtime, entrypoint, Debian
	// packages and limits. Its code is Files, not yet saved.
	Source *entity.DeploymentFunctionSource
	Files  []*entity.FunctionFile
	// Request is the request the handler is called with.
	Request *Request
	// Images are the release's runtime images, by runtime.
	Images map[string]string
	// Inputs are the build's, for the install of the libraries: its variables,
	// its secrets, the project's registries.
	Inputs *imagebuildservice.BuildInputs
	// BuildSettings are the app's build settings, for the install's resources.
	BuildSettings *entity.ImageBuildSettings
	// Env is the function's environment, as its service has it.
	Env []string
	// Network is the network the function's service is on, empty for docker's
	// default.
	Network string
	// NanoCPUs and MemoryBytes are the function's limits, zero for none.
	NanoCPUs    int64
	MemoryBytes int64
}

// RunResp is what a test run brings back.
type RunResp struct {
	Outcome Outcome
	// Status, Headers and Body are the handler's response, when the runtime
	// wrote a result.
	Status        int
	Headers       map[string][]string
	Body          []byte
	BodyTruncated bool
	RequestID     string
	DurationMs    float64
	// Logs are what the call wrote on its standard output, its result aside.
	Logs          string
	LogsTruncated bool
	// Error is what the runtime wrote on its standard error.
	Error string
	// ExitCode is the container's.
	ExitCode int64
	// LibrariesBuilt says whether this run installed the libraries, and
	// LibrariesLog is the install's log when it did.
	LibrariesBuilt bool
	LibrariesLog   string
	// LockFiles are the lock files the run made or changed, for the editor to
	// add to the code.
	LockFiles []*entity.FunctionFile
}

// LibrariesBuildReq is the build of a libraries image on this node.
type LibrariesBuildReq struct {
	// Image is the image's name, local to the node.
	Image      string
	Dockerfile string
	// ContextDir holds the function's files and the Dockerfile.
	ContextDir    string
	TempDir       string
	Inputs        *imagebuildservice.BuildInputs
	BuildSettings *entity.ImageBuildSettings
}

// LibrariesBuilder builds a libraries image, and says what the build wrote.
type LibrariesBuilder interface {
	BuildLibraries(ctx context.Context, req *LibrariesBuildReq) (log string, err error)
}
