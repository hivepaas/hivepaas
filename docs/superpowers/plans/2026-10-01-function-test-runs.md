# Function Test Runs Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Part 3 of functions. A function can be called once with code not yet saved: `POST …/apps/{appID}/function/test-run` sends the editor's files and a request to a build node, which installs the function's libraries into an image it keeps, runs `hivepaas-runtime invoke` in a throwaway container from it, and answers the response, the logs, the duration and a lock file the install made.

**Architecture:**
- **The libraries' image** is the function's Dockerfile up to its install (`functionbuild.Libraries`), named `hivepaas-function-libs:<hash>` after what it is built from, built through the image build service as a local image (`ImageBuildReq.LocalImage`) and kept on the node.
- **A run on a node** (`functiontest.Runner`) finds or builds that image, creates a container through Docker's API with the function's environment, limits, network and resources, copies the code into `/app` and the request into a file, runs `invoke` reading it, reads what it wrote and its lock files, and removes the container whatever happens.
- **The build node** is chosen as for a build (`functionservice.Service.TestRun`): this node runs the run itself, another node's agent runs it (`FunctionService.FunctionTestRun`, a gRPC call carrying the code and the request).
- **The API** (`appuc.TestRunFunction`) loads the function's saved settings, sends the code and the request, and answers what came back; the handler lifts the server's write timeout for it.

**Tech Stack:** Go 1.27, the backend as it is, the moby API client, gRPC and protoc (`protoc` 35.1, `protoc-gen-go` 1.36.12, `protoc-gen-go-grpc` 1.6.2, the versions the agent's generated files carry), Docker buildx through the image build service.

**Spec:** `docs/superpowers/specs/2026-10-01-functions-design.md`, section 4 and part 3 of section 9, amended by this plan's commit (its last section, "Changes after part 3's plan"). Part 2's plan is `docs/superpowers/plans/2026-10-01-functions-backend.md`; the runtimes' contract (`invoke`, its result line, its exit codes) is `CONTRACT.md` of `hivepaas/function-runtimes` v1.0.0.

## How the patches work

Every task's code was written and verified before this plan.

**Repository:** this one (the backend). **Base:** `main` at this plan's commit, which adds only documents to `ddc4330f` ("feat(functions): a function's container settings keep what is fixed for it"), the commit the patches were cut from. Work on a branch in a worktree.

Patches live in `docs/superpowers/plans/2026-10-01-function-test-runs/`: `tr-NN-tests.patch` then `tr-NN-code.patch` for task NN. The commands below run from the repository's root (or the worktree's), with `P=/Users/tnt/go/src/github.com/hivepaas/hivepaas/docs/superpowers/plans/2026-10-01-function-test-runs`.

The branch the patches were cut from is kept as `prep/function-test-runs`. After the last task, `git diff prep/function-test-runs -- . ':(exclude)docs/superpowers'` is empty.

**What was checked, on a fresh worktree of `ddc4330f`, at every task:**
- the tests alone fail, as each task's step 2 says;
- with the code, they pass and `go build ./...` passes;
- the task's lint commands of step 4 find 0 issues; Task 4's `function.proto` generates the committed files again unchanged; after Task 6 `make gen-swag` changes nothing.

**After the last task**, on the prep branch, whose tree the last task equals: `golangci-lint run ./...` 0 issues; `goroutinelint` and `errcodelint` clean; `go test ./...` all `ok` (213 packages).

**What it needs on the machine:** Go 1.27, golangci-lint 2.13, Docker for `make gen-swag`. The opt-in `TestAFunctionOfEachRuntimeIsTestRun` (`HIVEPAAS_DOCKER_TESTS=1`) test-runs a function of each runtime twice on this machine's Docker, building the libraries with docker's default builder (not HivePaaS's builder): about 20 seconds once the runtime images are pulled; it passed while preparing this plan and removes the images it builds.

## Global Constraints

- **The contract:** `hivepaas-runtime invoke` reads the request as JSON on standard input (`method`, `path`, `query`, `headers`, `body` in base64) and writes `#hivepaas-result {…}` as the last line of standard output; exit 0 with a result, 2 when the request cannot be read, 3 when the handler cannot be loaded or a Go function built. The runtime images run as uid 10001, which owns `/app`.
- **A test run's container:** the libraries' image; the function's environment, then `HP_FN_ENTRYPOINT`, `HP_FN_HANDLER`, `HP_FN_TIMEOUT_MS`, `HP_FN_MAX_CONCURRENCY`, `HP_FN_MAX_BODY_SIZE` from its settings; its network (`<project>_<env>` network of the project); its CPU and memory limits; no published port; named `hivepaas-cont-fn-…` with the temporary resource labels; removed whatever happens; killed when it outlives the timeout by 30 s (Go: 5 min).
- **What comes back:** the body and the logs cut at 1 MB each (the logs keep their end), the install's log cut at 64 KB (its end), lock files (`package-lock.json`; `requirements.lock`; `go.mod` and `go.sum`) that the run made or changed.
- **Limits:** a test run's request body at most 1 MB, its code the inline code's limits (1 MB, 100 files); a test run lasts at most 15 minutes, its HTTP answer is written within 16.
- **The backend's conventions** (`docs/ARCHITECTURE.md`), the linters (`golangci-lint run ./...`, `goroutinelint`, `errcodelint`), `make gen-swag` when a DTO changes, `go test ./...` in the main checkout after merging (it has `vendor/`), tests with `assert` and fakes, no database.
- **Git:** commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`; merge locally, never push; never `git stash`.

## Review Focus

Inputs the tests do not reach, or reach only in part:

1. **A test-run container on a worker node joining the project's network.** The network is an attachable overlay, so a plain container may join it by name; the opt-in test runs on the default bridge of one machine.
   - Pinned by: `TestATestRunCallsTheHandlerInAContainerOfItsOwn` (the network asked for, Task 3). The live check on the Linux server, with two nodes, runs it.
2. **A first test run that pulls the Go build image (1.3 GB) and installs modules** through the proxy in front of HivePaaS: within the 15 minutes, and with no proxy cutting a long answer.
   - Pinned by: nothing; the handler lifts the server's own write timeout (Task 6).
3. **A test run on another node, through a real agent:** the code, the build's secrets and the function's environment travel in one gRPC message, as a build's inputs do.
   - Pinned by: `TestATestRunCrossesToTheAgentAndBack`, `TestALargeAnswerComesBack` (Task 4), in process.
4. **The image prune removing a libraries image between its build and the container's creation**, or two runs installing the same new libraries at once: the second build is a cache hit; a pruned image fails the run's container creation with Docker's "No such image", which a second run fixes.
   - Pinned by: nothing.
5. **Code sent for a runtime the saved settings do not have** - the editor switched runtime without saving: the run uses the saved settings, and the runtime cannot load the code (`not-loaded`, with its message).
   - Pinned by: `TestAHandlerThatCannotBeLoadedSaysWhy` (Task 3).

## Decisions beyond the spec

- **A test run takes the code, as files, and a request; the rest is the function's saved settings**: runtime, entrypoint, Debian packages, limits. A function in a repository is test-run with the files the editor sends.
- **The function's environment, CPU and memory limits are its service's**, read from Swarm: what the deployed function has. Its container joins the project's network, so the function reaches the project's other apps by name.
- **The request reaches `invoke` through a file** (`/tmp/hivepaas-request.json`, read by `sh -c "exec hivepaas-runtime invoke < …"`), so that the container needs no attached standard input. Its body is text; the response's body is base64.
- **The call is synchronous**: the handler lifts the server's write timeout to 16 minutes, the use case bounds the run to 15. The install's log comes back with the answer (its last 64 KB), not as it happens.
- **The libraries' image is the runtime's own when there is nothing to install**, pulled if the node lacks it; otherwise it is built by the image build service on HivePaaS's builder, with the build's inputs, as a local image never pushed (`ImageBuildReq.LocalImage`). Its name hashes the Dockerfile and the files the install copies.
- **The cluster cleanup's image prune removes the libraries' images**, like any image no container uses, rather than its build cache step: they are images.
- **A test run does not wait for a build slot**: with every build node at its maximum it fails with "unavailable".
- **Outcomes beyond the runtime's** `ok`, `error`, `timeout`: `libraries-failed` (with the install's log), `not-loaded` (exit 3), `bad-request` (exit 2), `killed` (past the timeout's margin), `no-result` (another end, with its exit code: 137 for memory).
- **The agent's call is unary**, the answer allowed up to 16 MB; the request stays under gRPC's 4 MB, which the 1 MB code and the 1 MB body keep it.
- **Not yet:** a test run through MCP; an audit entry per test run (who may run one may deploy the same code); the install's log as it happens.

---

## Task 1: The libraries' image a function's test runs start from

**Files:**
- Create: `hivepaas_app/service/functionservice/functionbuild/libraries.go`
- Modify: `hivepaas_app/service/functionservice/functionbuild/functionbuild.go` (the stages both Dockerfiles share: `writeStages`)
- Test: `hivepaas_app/service/functionservice/functionbuild/libraries_test.go`

**Interfaces:**
- Produces, in `functionbuild`:
  - `const LibrariesRepo = "hivepaas-function-libs"`;
  - `type LibrariesResp struct { Dockerfile, Image string; Notes []string }` and `func Libraries(req *DockerfileReq) (*LibrariesResp, error)`: the function's Dockerfile up to its install, without code or limits; `Image` is `LibrariesRepo:<24 hex>`, a hash of the Dockerfile and of the files the install copies (all files when the manifest names the function's own); with nothing to install and no package, `Dockerfile` is empty and `Image` the runtime's (Go: its build image);
  - `func LockFiles(runtime base.FunctionRuntime) []string`.
- `Dockerfile` (part 2) writes the same content as before.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/tr-01-tests.patch`

Tests added: `TestALibrariesImageIsTheInstallAlone` (the exact Dockerfile of a Node.js function with a package, a build argument and a secret), `TestAGoFunctionsLibrariesAreInTheBuildImage`, `TestALibrariesImageIsNamedAfterWhatItIsBuiltFrom`, `TestLibrariesFromTheFunctionsOwnFilesAreNamedAfterTheCode`, `TestAFunctionWithoutLibrariesRunsOnItsRuntimesImage`, `TestEachRuntimeNamesItsLockFiles`, `TestALibrariesImageNeedsItsRuntimesImage`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/service/functionservice/functionbuild/`
Expected: FAIL to build: `undefined: LibrariesResp`, `undefined: Libraries`, `undefined: LibrariesRepo`, `undefined: LockFiles`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/tr-01-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/functionservice/functionbuild/`
Expected: `ok` (part 2's Dockerfile tests included)

Run: `golangci-lint run ./hivepaas_app/service/functionservice/...`
Expected: `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(functions): the libraries image a function's test runs start from

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 2: A build can make a local image

**Files:**
- Modify: `hivepaas_app/service/imagebuildservice/types.go`, `hivepaas_app/service/imagebuildservice/imagebuildserviceimpl/{build_image.go,helper.go,push_image.go}`
- Test: `hivepaas_app/service/imagebuildservice/imagebuildserviceimpl/local_image_test.go`

**Interfaces:**
- Produces: `imagebuildservice.ImageBuildReq.LocalImage string`: when set, the image's only name; no app names it and it is never pushed. `imageReferences(req, inputs) ([]string, error)` in the impl: `[LocalImage]`, or the app's names as before.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/tr-02-tests.patch`

Tests added: `TestALocalImageIsNamedOnlyItsName`, `TestALocalImageIsNotPushed`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/service/imagebuildservice/...`
Expected: FAIL to build: `undefined: imageReferences`, `unknown field LocalImage in struct literal of type imagebuildservice.ImageBuildReq`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/tr-02-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/imagebuildservice/...`
Expected: `ok`

Run: `golangci-lint run ./hivepaas_app/service/imagebuildservice/...`
Expected: `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(build): a build can make a local image, named only its own name

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 3: A test run on a node, in a throwaway container from the libraries' image

**Files:**
- Create: `hivepaas_app/service/functionservice/functiontest/{functiontest.go,run.go,builder.go}`
- Test: `hivepaas_app/service/functionservice/functiontest/{fakes_test.go,run_test.go,builder_test.go,real_run_test.go}`

**Interfaces:**
- Consumes: Task 1's `functionbuild.Libraries`, `LockFiles`; Task 2's `LocalImage`; part 2's `functionbuild.WriteInlineCode`, `DockerfilePath`.
- Produces, in `functiontest`:
  - `type Request struct { Method, Path string; Query, Headers map[string][]string; Body []byte }` (invoke's JSON);
  - `type RunReq struct { Source *entity.DeploymentFunctionSource; Files []*entity.FunctionFile; Request *Request; Images map[string]string; Inputs *imagebuildservice.BuildInputs; BuildSettings *entity.ImageBuildSettings; Env []string; Network string; NanoCPUs, MemoryBytes int64 }`;
  - `type RunResp struct { Outcome Outcome; Status int; Headers map[string][]string; Body []byte; BodyTruncated bool; RequestID string; DurationMs float64; Logs string; LogsTruncated bool; Error string; ExitCode int64; LibrariesBuilt bool; LibrariesLog string; LockFiles []*entity.FunctionFile }`;
  - `type Outcome string` with `OutcomeOK`, `OutcomeError`, `OutcomeTimeout`, `OutcomeLibrariesFailed`, `OutcomeNotLoaded`, `OutcomeBadRequest`, `OutcomeKilled`, `OutcomeNoResult`;
  - `type Runner struct { Docker docker.Manager; Builder LibrariesBuilder; TempDir string; KillMargin time.Duration }` with `Run(ctx, *RunReq) (*RunResp, error)`;
  - `type LibrariesBuilder interface { BuildLibraries(ctx, *LibrariesBuildReq) (log string, err error) }`, `LibrariesBuildReq{Image, Dockerfile, ContextDir, TempDir, Inputs, BuildSettings}`, and `NewLibrariesBuilder(builds imagebuildservice.Service) LibrariesBuilder`;
  - constants `ResultMarker`, `RuntimeUID` (10001), `BodyMax`, `LogsMax` (1 MB), `LibrariesLogMax` (64 KB).

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/tr-03-tests.patch`

Tests added: `TestATestRunCallsTheHandlerInAContainerOfItsOwn`, `TestTheCodeIsTheRuntimeUsers`, `TestTheLibrariesAreBuiltOnce`, `TestAFunctionWithoutLibrariesRunsOnItsRuntimesImage`, `TestLibrariesThatDoNotInstallEndTheRun`, `TestAHandlerThatCannotBeLoadedSaysWhy`, `TestAContainerThatOutlivesItsTimeoutIsKilled`, `TestABodyAndLogsAreCutAt1MB`, `TestTheLockFileARunMadeComesBack`, `TestTheLimitsAreTheFunctionsSettings`, `TestARunWithoutAResultSaysHowItEnded`, `TestTheLibrariesAreBuiltAsALocalImage`, `TestALongInstallLogKeepsItsEnd`, and the opt-in `TestAFunctionOfEachRuntimeIsTestRun`. The fakes stand for Docker (images, the container, its logs as Docker multiplexes them, `/app` after the run) and the libraries' build.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/service/functionservice/functiontest/`
Expected: FAIL to build: `undefined: LibrariesBuildReq`, `undefined: RunReq`, `undefined: Runner`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/tr-03-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/functionservice/functiontest/`
Expected: `ok`

Run: `golangci-lint run ./hivepaas_app/service/functionservice/...`
Expected: `0 issues.`

Optional, with Docker: `HIVEPAAS_DOCKER_TESTS=1 go test -count=1 -run TestAFunctionOfEachRuntimeIsTestRun -v ./hivepaas_app/service/functionservice/functiontest/`
Expected: `--- PASS` for `node24`, `python313` and `go127`; afterwards `docker ps -a --filter label=hivepaas.system.temp=true` lists nothing

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(functions): a test run on a node, in a throwaway container from the libraries' image

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 4: A test run on another node, through its agent

**Files:**
- Create: `hivepaas_app/interface/agent/proto/function.proto` and its generated `function.pb.go`, `function_grpc.pb.go`; `hivepaas_app/interface/agent/client/functionservice/service.go`; `hivepaas_app/interface/agent/server/functionservice/test_run.go`; `hivepaas_app/interface/agent/server/server_function.go`; `hivepaas_app/usecaseagent/functionagentuc/uc.go`
- Modify: `hivepaas_app/interface/agent/proto/generate.go`; `hivepaas_app/interface/agent/client/imagebuildservice/service.go` and `hivepaas_app/interface/agent/server/imagebuildservice/image_build.go` (the build's inputs and settings mappings, exported for the function service: `InputsToProto`, `BuildSettingsToProto`, `InputsFromProto`, `BuildSettingsFromProto`); `hivepaas_app/interface/agent/server/server.go`; `hivepaas_app/cmd/agent/main.go`; `hivepaas_app/registry/provides.go`
- Test: `hivepaas_app/interface/agent/client/functionservice/service_test.go`

**Interfaces:**
- Consumes: Task 3's `functiontest.Runner`, `RunReq`, `RunResp`, `NewLibrariesBuilder`.
- Produces:
  - `service FunctionService { rpc FunctionTestRun(FunctionTestRunReq) returns (FunctionTestRunResp); }`, its messages reusing `ImageBuildInputs` and `ImageBuildSettings`;
  - `functionagentuc.New(logger, dockerManager, imageBuildService) *UC` (a `functiontest.Runner` on the agent's node, temporary files under `base.BaseTempDirDefault`), `(*UC).WithRunner(Runner) *UC`, `(*UC).TestRun(ctx, *functiontest.RunReq) (*functiontest.RunResp, error)` (a panic is an error);
  - client: `functionservice.FunctionServiceClient` (`TestRun`, `Close`) and `NewFunctionServiceClient(agentAddr string)`; it refuses a run without inputs, and reads answers up to 16 MB.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/tr-04-tests.patch`

Tests added: `TestATestRunCrossesToTheAgentAndBack` (every field of a run, both ways, through an in-process gRPC server), `TestALargeAnswerComesBack` (a 4 MB answer), `TestARunThatCannotBeMadeIsAnError`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/interface/agent/... ./hivepaas_app/usecaseagent/... ./hivepaas_app/cmd/internal/`
Expected: FAIL: `no required module provides package …/interface/agent/server/functionservice` for `client/functionservice`; the other packages `ok`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/tr-04-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/interface/agent/... ./hivepaas_app/usecaseagent/... ./hivepaas_app/cmd/internal/`
Expected: every package `ok` (`cmd/internal` validates both fx graphs)

Run: `golangci-lint run ./hivepaas_app/interface/agent/... ./hivepaas_app/usecaseagent/... ./hivepaas_app/registry/... ./hivepaas_app/cmd/... && go run ./tools/errcodelint`
Expected: `0 issues.`; errcodelint prints nothing and exits 0

Run: `(cd hivepaas_app/interface/agent/proto && protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative function.proto) && git status --short hivepaas_app/interface/agent/proto | grep -v '^[AM]  '`
Expected: no output (the committed files are what protoc writes)

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(functions): a test run on another node, through its agent

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 5: A test run on a build node, as a build picks one

**Files:**
- Create: `hivepaas_app/service/functionservice/service.go`, `hivepaas_app/service/functionservice/functionserviceimpl/{service.go,test_run.go}`
- Modify: `hivepaas_app/registry/provides.go`
- Test: `hivepaas_app/service/functionservice/functionserviceimpl/test_run_test.go`

**Interfaces:**
- Consumes: Task 3's runner, Task 4's `NewFunctionServiceClient`; `imagebuildservice.ResolveBuildInputs`, `SelectBuildWorkerNode`; `clusterservice.ServiceInspect`; `networkservice.GetProjectNetworkName`; `agentservice.GetAgentAddrForNode`; `systemappservice.CurrentRelease().FunctionRuntimes`.
- Produces: `functionservice.Service` with `TestRun(ctx, db, *TestRunReq) (*functiontest.RunResp, error)`, `TestRunReq{App (with Project and ProjectEnv), Source, Files, Request}`; `functionserviceimpl.New(dockerManager, settingRepo, imageBuildService, clusterService, networkService, agentService) functionservice.Service`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/tr-05-tests.patch`

Tests added: `TestATestRunIsTheFunctions` (the build's inputs, the service's environment, limits and network, the release's images, the project's build settings, the build slot given back), `TestATestRunOnAnotherNodeGoesToItsAgent`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/service/functionservice/... ./hivepaas_app/cmd/internal/`
Expected: FAIL: `no required module provides package …/service/functionservice` for `functionserviceimpl`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/tr-05-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/functionservice/... ./hivepaas_app/cmd/internal/`
Expected: every package `ok`

Run: `golangci-lint run ./hivepaas_app/service/functionservice/... ./hivepaas_app/registry/...`
Expected: `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(functions): a test run on a build node, as a build picks one

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 6: The test run API

**Files:**
- Create: `hivepaas_app/usecase/appuc/{test_run_function.go}`, `hivepaas_app/usecase/appuc/appdto/test_run_function.go`, `hivepaas_app/interface/api/handler/apphandler/test_run_function.go`
- Modify: `hivepaas_app/usecase/appuc/uc.go` (the function service), `hivepaas_app/usecase/appsettingsuc/appsettingsdto/deployment_function_source.go` (`FunctionInlineCodeReq.Normalize`, `Validate`, `ToEntity`), `hivepaas_app/interface/api/server/router_apps.go`, `hivepaas_app/hperrors/errors_project_app.go`, `hivepaas_app/pkg/translation/messages/en/errors.project_app.en.toml`, `docs/openapi/swagger.json`
- Test: `hivepaas_app/usecase/appuc/test_run_function_test.go`, `hivepaas_app/usecase/appuc/appdto/test_run_function_test.go`, `hivepaas_app/interface/api/server/router_apps_test.go`

**Interfaces:**
- Consumes: Task 5's `functionservice.Service`.
- Produces:
  - `POST /projects/{projectID}/{projectEnv}/apps/{appID}/function/test-run` → `apphandler.Handler.TestRunFunction` (write permission on the app; the write deadline lifted to 16 minutes) → `appuc.UC.TestRunFunction(ctx, auth, *appdto.TestRunFunctionReq) (*appdto.TestRunFunctionResp, error)`, bounded to 15 minutes, 200;
  - `appdto.TestRunFunctionReq{ProjectID, ProjectEnvID, AppID, Code *appsettingsdto.FunctionInlineCodeReq, Request *TestRunRequestReq{Method, Path, Query, Headers, Body string}}`: `code` required; the method one of GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS (GET by default); the path from `/` (`/` by default); header names in lower case; the body at most 1 MB;
  - `appdto.TestRunFunctionDataResp{Outcome, Status, Headers, Body (base64), BodyTruncated, RequestID, DurationMs, Logs, LogsTruncated, Error, ExitCode, LibrariesBuilt, LibrariesLog, LockFiles []*appsettingsdto.FunctionFileResp}`;
  - `hperrors.ErrAppNotFunction` (`ERR_APP_NOT_FUNCTION`, 412, param `Name`).

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/tr-06-tests.patch`

Tests added: `TestATestRunTakesTheCodeAndARequest`, `TestATestRunRequestIsNormalized`, `TestWhatATestRunCannotTakeIsRefused`, `TestATestRunCallsTheFunctionWithTheCodeSent`, `TestAnAppThatIsNotAFunctionHasNoTestRun`, `TestAFunctionIsTestRunOnItsApp`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/usecase/appuc/... ./hivepaas_app/usecase/appsettingsuc/appsettingsdto/ ./hivepaas_app/interface/api/server/ ./hivepaas_app/cmd/internal/`
Expected: FAIL: `undefined: TestRunFunctionReq`, `unknown field functionService in struct literal of type UC`, and `--- FAIL: TestAFunctionIsTestRunOnItsApp`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/tr-06-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/usecase/appuc/... ./hivepaas_app/usecase/appsettingsuc/appsettingsdto/ ./hivepaas_app/interface/api/server/ ./hivepaas_app/cmd/internal/`
Expected: every package `ok`

- [ ] **Step 5: The whole repository, and the tree**

Run: `golangci-lint run ./... && go run ./tools/goroutinelint . && go run ./tools/errcodelint`
Expected: `0 issues.`; the two custom linters print nothing and exit 0

Run: `go test ./...`
Expected: every package `ok`

Run: `make gen-swag && git status --short | grep -v '^[AM]  '`
Expected: no output

Run: `git diff --stat prep/function-test-runs -- . ':(exclude)docs/superpowers'`
Expected: no output

- [ ] **Step 6: Commit**

```bash
git commit -m "feat(functions): the test run API

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## After the plan (the user)

- Merge into `main` locally; run `go test ./...` in the main checkout.
- The agent changes: the build nodes' agents need this release before a test run reaches them.
- The live check on the Linux server: a test run of each runtime on the node HivePaaS runs on and on another build node; one whose function reads another app of its project; the first run of a Go function on a fresh node.
- Part 4 (the dashboard) starts from this API: the editor's files and the request in, the answer, the logs and the lock file out.
