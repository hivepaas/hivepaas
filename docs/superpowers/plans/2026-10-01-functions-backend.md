# Functions in the Backend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Part 2 of functions. HivePaaS knows a function: an app of kind `function` whose deployment settings hold its code, built through the existing pipeline from a Dockerfile HivePaaS writes on the runtime images `release.json` names, and run with the runtime's command, health check and port, and a stop grace period above its timeout.

**Architecture:**
- **The release** names the four runtime images by digest, in `functionRuntimes`; `make release-pin` pins them with the other images.
- **The model:** `base.AppCategoryFunction`, `base.DeploymentMethodFunction`, and `entity.DeploymentFunctionSource` (runtime, contract, entrypoint, inline code or a repository, Debian packages, the limits of one call, a registry) in the app's deployment settings. Inline code is kept in the setting, so a deployment's snapshot holds the code it ran.
- **The Dockerfile** is written by `functionbuild.Dockerfile` for each build: the runtime image, the Debian packages as root, `hivepaas-runtime deps` with every build secret mounted, the code, the `HP_FN_*` limits as `ENV`; Go compiles in its build image and runs on the slim one. It goes into the build as a manual Dockerfile at `.hivepaas/Dockerfile`.
- **The deployment** is a repository's, with a source step of its own (inline files written, or the repository checked out and its function's directory taken as the context), then the same build, lock, commands and service update; the container gets what is fixed for a function (`functioncontainer.ApplyFixed`).
- **The rules** keep a function a function: only `POST …/apps/function` makes one; its kind never changes; its method is `function`, and only its is; its port and its domains' are 8080; its container settings keep what is fixed; a clone that takes its code is one; a spec neither exports nor makes one yet.

**Tech Stack:** Go 1.27, the backend as it is (gin, bun, uber fx, `tiendc/go-validator`, the moby API), Docker buildx through the existing image build service.

**Spec:** `docs/superpowers/specs/2026-10-01-functions-design.md`, part 2 of section 9, amended by this plan's commit (its last section, "Changes after part 2's plan"). Part 1's plan is `docs/superpowers/plans/2026-10-01-function-runtimes.md`; the runtimes' contract is `CONTRACT.md` of `hivepaas/function-runtimes` v1.0.0.

## How the patches work

Every task's code was written and verified before this plan.

**Repository:** this one (the backend). **Base:** `main` at this plan's commit, which adds only documents to `6fc60091` ("docs(plans): functions part 1 - the runtimes, with verified patches"), the commit the patches were cut from. Work on a branch in a worktree.

Patches live in `docs/superpowers/plans/2026-10-01-functions-backend/`: `be-NN-tests.patch` then `be-NN-code.patch` for task NN. The commands below run from the repository's root (or the worktree's), with `P=/Users/tnt/go/src/github.com/hivepaas/hivepaas/docs/superpowers/plans/2026-10-01-functions-backend`.

The branch the patches were cut from is kept as `prep/functions-backend`. After the last task, `git diff prep/functions-backend -- . ':(exclude)docs/superpowers'` is empty: the two differ only by this plan's documents.

**What was checked, on a fresh worktree of `6fc60091`, at every task:**
- the tests alone fail, as each task's step 2 says;
- with the code, they pass and `go build ./...` passes;
- the task's lint commands of step 4 find 0 issues, `errcodelint` included where the step runs it, and `make gen-swag` changes nothing after Tasks 1, 5, 6 and 8.

**After the last task**, on the prep branch, whose tree the last task equals: `go build ./...`; `golangci-lint run ./...` (the whole repository) 0 issues; `go run ./tools/goroutinelint .` and `go run ./tools/errcodelint` clean; `go test ./...` all `ok`; `make gen-swag` changes nothing.

**What it needs on the machine:** Go 1.27, golangci-lint 2.13, Docker for `make gen-swag` (the `hivepaas-devtools` and `openapitools` images). The opt-in test `TestAFunctionOfEachRuntimeBuildsOnTheReleasesImages` (`HIVEPAAS_DOCKER_TESTS=1`) builds and calls a function of each runtime with docker's default builder: it pulls the four runtime images (the Go build image is 1.3 GB) and each function's libraries, and takes about a minute once they are pulled; it was run while preparing this plan.

## Global Constraints

- **Runtimes:** `node24`, `python313`, `go127`; Go's build image is keyed `go127-build`. **Contract:** `v1`. `release.json` and `base.StableVersion`/`base.BetaVersion` name the images by digest, as part 1 released them:
  - `ghcr.io/hivepaas/function-runtime-node24:1.0.0@sha256:a4560e0017d5c80f64973473bdf9612ed335b2ea79a7cf7cb39da259ea636970`
  - `ghcr.io/hivepaas/function-runtime-python313:1.0.0@sha256:9d3a28bc48cbbd41d1c7728e2b1a9bb53b5cf97a3d32ebfa8b1f84450c7b6bf2`
  - `ghcr.io/hivepaas/function-runtime-go127:1.0.0@sha256:31fce6af839d98fdad7ae982ea8ed2696b8f8aaefca0866a50a227f57bd3e169`
  - `ghcr.io/hivepaas/function-runtime-go127-build:1.0.0@sha256:c18ddbeb26c9d22b081cb21a330994952ab219c72fbcaf9b2bd2855fe30ec631`
- **Defaults:** entrypoint `index.js`/`default`, `main.py`/`handler`, `.`/`Handle`; timeout 30 s (1 s to 15 min); concurrency 16 (1 to 1000); body size 6 MB (1 KB to 100 MB).
- **Inline code:** at most 1 MB in all and 100 files; paths relative, plain (`^[A-Za-z0-9._/-]+$`, at most 255 characters), without `..`; `.hivepaas` is reserved. **Debian packages:** at most 50, each `^[a-z0-9][a-z0-9+.-]+(=[A-Za-z0-9.+~:-]+)?$`.
- **Fixed for a function:** port 8080; the runtime's command and working directory (none set); the image's health check (none set); a stop grace period of the timeout plus 10 s.
- **The Dockerfile** is at `.hivepaas/Dockerfile` in the build's context, a manual Dockerfile. Its last stage has `ENV HP_FN_ENTRYPOINT`, `HP_FN_HANDLER`, `HP_FN_TIMEOUT_MS`, `HP_FN_MAX_CONCURRENCY`, `HP_FN_MAX_BODY_SIZE`.
- **The backend's conventions** (`docs/ARCHITECTURE.md`): handler → dto → usecase → service → repository; a request's `ModifyRequest` then `Validate`; errors are `hperrors.NewErr` with a message in `pkg/translation/messages/en/*.toml` (`errcodelint` checks); exhaustive switches over enums, without `//nolint:exhaustive`; 120 columns, US spelling.
- **Tests** use `assert` (`testify/require` is not vendored) and fakes that embed the interface they fake; none needs a database.
- **Before calling it done:** `go build ./...`, `golangci-lint run ./...` (the whole repository), `go run ./tools/goroutinelint .`, `go run ./tools/errcodelint`, `go test ./...`, and `make gen-swag` when a DTO changed (`docs/openapi/swagger.json` is committed). The main checkout has a git-ignored `vendor/`: after merging, run `go test ./...` there too.
- **Git:** commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`; merge locally, never push; never `git stash`.

## Review Focus

Inputs the tests do not reach, or reach only in part:

1. **A build on another node, through the agent.** The function's context and Dockerfile go to the build node as a repository's do, and the build service writes the manual Dockerfile into the context there. The tests check the request the build is given; the opt-in test builds with docker's own builder on this machine.
   - Pinned by: `TestAFunctionIsBuiltFromTheDockerfileWrittenForIt` (Task 4), `TestAFunctionOfEachRuntimeBuildsOnTheReleasesImages` (Task 3, opt-in). The spec's live check on the Linux server runs the agent's path.
2. **A repository whose function directory has a `.dockerignore`.** It applies to the build's context: one that leaves out the manifest or the lock file fails the install step with docker's own message. A `.hivepaas/Dockerfile` of the repository's own is overwritten in the checkout.
   - Pinned by: nothing; the build log shows docker's error.
3. **Inline code without a lock file.** Each build resolves the libraries' versions again (the build log warns), so a redeploy can pick up a newer minor version. The spec saves the lock file back; this part does not (see "Decisions beyond the spec").
   - Pinned by: `TestALockFileThatIsMissingIsNoted` (Task 3).
4. **An old deployment redeployed.** Its snapshot holds its own code and limits, so it rebuilds what it ran, on the runtime images of the release running now; identical code and Dockerfile give the same tag.
   - Pinned by: `TestTheSameFunctionHasTheSameContentHash` (Task 4); the redeploy itself is the existing path.
5. **A runtime changed in the settings.** Defaults fill only what is empty: a Node.js function moved to `python313` keeps `index.js`, which is refused until its entrypoint is changed too.
   - Pinned by: the same rule the other way, the case `a Node.js handler file that is not JavaScript` of `TestAFunctionSourceThatCannotBeRunIsRefused` (Task 2).

## Decisions beyond the spec

- **The lock file is not saved back after a deployment's build** (spec §3, step 3). The build runs on a build node and returns an image; bringing a file back needs a second output of the build, through the agent. Part 3's test run returns the lock file its install made, which the editor adds to the code for the author to save; until then a build without one says so in its log.
- **`FunctionCode.Repo` is a `FunctionRepoCode`** (type, URL, ref, commit, options, credentials), not the `DeploymentRepoSource` the spec sketched: a repository source's Dockerfile and registry do not apply to a function. `RepoSource()` gives the checkout what it takes.
- **`PushToRegistry` is on the function's source**, as on a repository's. What a build needs - a registry on a cluster of several nodes, the repository and its ref - is checked by `appdeploymentservice.CheckBuildSource`, shared by both, when the settings are saved and before a function is created.
- **The `HP_FN_*` limits are the image's `ENV`**, written by the Dockerfile, rather than the service's: they take effect at a build, and a deployment's image carries the limits it was built with.
- **An inline function's image is tagged after a hash of its files and its Dockerfile** (`functionbuild.ContentHash`), where a repository's is tagged after its commit.
- **The install step copies only the manifest, its lock file and the package manager's settings** (`.npmrc`, `uv.toml`), so a change to the code alone reuses the installed libraries. A manifest that refers to the function's own files gets the whole code first, with a note in the build log: in `package.json` a `file:`, `link:` or path version, or workspaces; in `requirements.txt` a path, `file:`, `-e`, `-r` or `-c`; in `go.mod` a `replace` to a directory.
- **Creating a function** is `POST /projects/{project}/{env}/apps/function`: the app's fields and the source. The source is checked before anything exists; the response carries the first deployment and its task.
- **The container settings keep what is fixed for a function**, as a deployment does: they write the service between deployments.
- **A clone of a function is a function when it takes the function's deployment settings**; without them it is an app like any other's clone.
- **Not yet:** a function in a spec (export leaves it out and reports `FUNCTION_SKIPPED`; a template or an import refuses a function's kind or deployment); a function in a repository deployed by the repository's webhook; creating a function through MCP (the MCP tools read, change and redeploy one); a build cache mount for Go modules, which builds of different functions would share.

---

## Task 1: The runtime images, pinned in the release

**Files:**
- Create: `hivepaas_app/base/function.go`
- Modify: `hivepaas_app/base/version.go`, `release.json`, `tools/releasepin/main.go`, `docs/openapi/swagger.json`
- Test: `hivepaas_app/base/function_test.go`, `hivepaas_app/service/hpappservice/hpappserviceimpl/release_templates_test.go`, `tools/releasepin/main_test.go`

**Interfaces:**
- Produces, in `base`:
  - `type FunctionRuntime string`; `FunctionRuntimeNode24 = "node24"`, `FunctionRuntimePython313 = "python313"`, `FunctionRuntimeGo127 = "go127"`; `AllFunctionRuntimes`;
  - `func (r FunctionRuntime) Compiled() bool` (Go only); `func FunctionBuildImageKey(runtime FunctionRuntime) string` (`"go127-build"`);
  - `ReleaseInfo.FunctionRuntimes map[string]string` (`json:"functionRuntimes,omitempty"`), keyed by runtime and by build image key; `StableVersion` and `BetaVersion` set it to the four images above.
- Produces, in `tools/releasepin`: the values of a `functionRuntimes` block are pinned like `image` fields.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-01-tests.patch`

Tests added: `TestEachReleaseNamesAnImageForEveryFunctionRuntime`, `TestOnlyGoIsCompiled`, `TestRepoReleaseJSONNamesTheCompiledFunctionRuntimes`, `TestFunctionRuntimesArePinnedToo`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/base/ ./hivepaas_app/service/hpappservice/hpappserviceimpl/ ./tools/releasepin/`
Expected: FAIL, `undefined: base.AllFunctionRuntimes`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-01-code.patch`

`release.json`'s beta gets `functionRuntimes`; `docs/openapi/swagger.json` is `make gen-swag`'s output for the new field.

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/base/ ./hivepaas_app/service/hpappservice/hpappserviceimpl/ ./tools/releasepin/`
Expected: three `ok`

Run: `golangci-lint run ./hivepaas_app/base/ ./hivepaas_app/service/hpappservice/hpappserviceimpl/ ./tools/releasepin/`
Expected: `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(release): the images functions are built on, pinned in the release

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 2: A function's source, checked as a request

**Files:**
- Create: `hivepaas_app/entity/app_deployment_function.go`, `hivepaas_app/usecase/appsettingsuc/appsettingsdto/deployment_function_source.go`
- Modify: `hivepaas_app/base/function.go`, `hivepaas_app/entity/app_deployment_settings.go`, `hivepaas_app/pkg/translation/messages/en/errors.validation.en.toml`
- Test: `hivepaas_app/entity/app_deployment_function_test.go`, `hivepaas_app/usecase/appsettingsuc/appsettingsdto/deployment_function_source_test.go`

**Interfaces:**
- Consumes: Task 1's `base.FunctionRuntime`.
- Produces, in `base`: `func (r FunctionRuntime) DefaultEntrypoint() (file, handler string)`; `type FunctionContract string`, `FunctionContractV1 = "v1"`, `AllFunctionContracts`; the constants `FunctionPort`, `FunctionTimeoutDefault/Min/Max`, `FunctionMaxConcurrencyDefault/Max`, `FunctionMaxBodySizeDefault/Min/Max`, `FunctionInlineCodeMaxSize`, `FunctionInlineCodeMaxFiles`, `FunctionSystemPackagesMax`, `FunctionReservedDir` (`".hivepaas"`), `FunctionSystemPackagePattern`.
- Produces, in `entity`:
  - `type DeploymentFunctionSource struct { Runtime; Contract; Entrypoint FunctionEntrypoint; Code FunctionCode; SystemPackages []string; Timeout timeutil.Duration; MaxConcurrency int; MaxBodySize unit.DataSize; PushToRegistry ObjectID }`;
  - `FunctionEntrypoint{File, Handler}`, `FunctionCode{Inline *FunctionInlineCode; Repo *FunctionRepoCode; Dir string}`, `FunctionInlineCode{Files []*FunctionFile}`, `FunctionFile{Path, Content}`;
  - `FunctionRepoCode{RepoType, RepoID, RepoURL, RepoRef, CommitHash, RepoOptions, Credentials}` with `RepoSource() *DeploymentRepoSource`;
  - `AppDeploymentSettings.FunctionSource *DeploymentFunctionSource` (`json:"functionSource,omitempty"`); its registry and its repository's credentials are among the settings' references.
- Produces, in `appsettingsdto`: `DeploymentFunctionSourceReq` (with `FunctionEntrypointReq`, `FunctionCodeReq`, `FunctionInlineCodeReq`, `FunctionFileReq`, `FunctionRepoCodeReq`) and its methods `Normalize()`, `Validate(field string) []vld.Validator`, `ToEntity() (*entity.DeploymentFunctionSource, error)`.
- Validation keys added: `ERR_VLD_FUNCTION_ENTRYPOINT_INVALID`, `ERR_VLD_FUNCTION_HANDLER_INVALID`, `ERR_VLD_FUNCTION_FILE_PATH_INVALID`, `ERR_VLD_FUNCTION_CODE_TOO_LARGE`, `ERR_VLD_DEBIAN_PACKAGE_INVALID`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-02-tests.patch`

Tests added: `TestAFunctionsRegistryAndCredentialsAreItsReferences`, `TestAFunctionsRepoIsCheckedOutAsARepoSource`, `TestAFunctionsInlineCodeIsKeptInItsSetting`, `TestAFunctionSourceTakesItsRuntimesDefaults`, `TestAFunctionSourceIsNormalized`, `TestAFunctionSourceOfEachRuntimeIsValid`, `TestAFunctionSourceThatCannotBeRunIsRefused` (24 refusals: runtime, contract, entrypoint per runtime, handler per runtime, code neither or both, files, paths, `.hivepaas`, size, a directory with inline code, the repository's URL, packages, the limits), `TestAFunctionSourceBecomesItsEntity`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/entity/ ./hivepaas_app/usecase/appsettingsuc/appsettingsdto/`
Expected: FAIL, `unknown field FunctionSource in struct literal of type AppDeploymentSettings`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-02-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/entity/ ./hivepaas_app/usecase/appsettingsuc/appsettingsdto/`
Expected: two `ok`

Run: `golangci-lint run ./hivepaas_app/entity/ ./hivepaas_app/usecase/appsettingsuc/appsettingsdto/ ./hivepaas_app/base/ && go run ./tools/errcodelint`
Expected: `0 issues.`; errcodelint prints nothing and exits 0

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(functions): a function's source, checked as a request

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 3: The Dockerfile HivePaaS writes for a function

**Files:**
- Create: `hivepaas_app/service/functionservice/functionbuild/functionbuild.go`
- Modify: `hivepaas_app/hperrors/errors_project_app.go`, `hivepaas_app/pkg/translation/messages/en/errors.project_app.en.toml`
- Test: `hivepaas_app/service/functionservice/functionbuild/functionbuild_test.go`, `hivepaas_app/service/functionservice/functionbuild/real_build_test.go`

**Interfaces:**
- Consumes: Task 2's `entity.DeploymentFunctionSource`, `entity.FunctionInlineCode`, `base.FunctionReservedDir`, `base.FunctionSystemPackagePattern`.
- Produces, in `functionbuild`:
  - `const DockerfilePath = ".hivepaas/Dockerfile"`;
  - `type DockerfileReq struct { Source *entity.DeploymentFunctionSource; Images map[string]string; SourceDir string; BuildArgs, BuildSecrets []string }` and `type DockerfileResp struct { Content string; Notes []string }`;
  - `func Dockerfile(req *DockerfileReq) (*DockerfileResp, error)`: `hperrors.ErrFunctionRuntimeUnavailable` when the release names no image for the runtime, `hperrors.ErrValueInvalid` for a package name that is not one;
  - `func WriteInlineCode(dir string, code *entity.FunctionInlineCode) error`.
- Produces, in `hperrors`: `ErrFunctionRuntimeUnavailable` (`ERR_FUNCTION_RUNTIME_UNAVAILABLE`, param `Runtime`).

The Dockerfile of a Node.js function with two Debian packages, a build argument and two build secrets is, exactly (the test holds it):

```dockerfile
# The Dockerfile HivePaaS writes for a function: runtime node24, contract v1.
FROM <node24 image> AS base
USER root
RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg libpq5 \
    && rm -rf /var/lib/apt/lists/*
USER hivepaas

FROM base AS deps
COPY --chown=hivepaas:hivepaas package.json package-lock.json .npmrc /app/
ARG NPM_CONFIG_LOGLEVEL
RUN --mount=type=secret,id=GH_TOKEN,env=GH_TOKEN --mount=type=secret,id=NPM_TOKEN,env=NPM_TOKEN \
    hivepaas-runtime deps

FROM deps
COPY --chown=hivepaas:hivepaas . /app/
ENV HP_FN_ENTRYPOINT="index.js" \
    HP_FN_HANDLER="default" \
    HP_FN_TIMEOUT_MS=30000 \
    HP_FN_MAX_CONCURRENCY=16 \
    HP_FN_MAX_BODY_SIZE=6291456
```

Only the manifest files present in the function are copied. Go's: `FROM <go127-build> AS base` (with the packages), `FROM base AS deps` (`go.mod go.sum`, `deps`), `FROM deps AS build` (the code, the entrypoint's `ENV`, `hivepaas-runtime build`), then `FROM <go127>` with the packages and `COPY --from=build --chown=hivepaas:hivepaas /app /app` and all five `ENV`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-03-tests.patch`

Tests added: `TestANodeFunctionInstallsItsLibrariesBeforeItsCode`, `TestAFunctionWithoutLibrariesHasNoInstallStep`, `TestALockFileThatIsMissingIsNoted`, `TestLibrariesFromTheFunctionsOwnFilesAreInstalledWithItsCode`, `TestAGoFunctionIsCompiledThenRunOnTheSlimImage`, `TestAHandlerNameStaysAsItIs`, `TestWhatCannotBeBuiltIsRefused`, `TestInlineCodeIsWrittenWhereItSays`, `TestInlineCodeOutsideTheFunctionIsNotWritten`, and the opt-in `TestAFunctionOfEachRuntimeBuildsOnTheReleasesImages`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/service/functionservice/functionbuild/`
Expected: FAIL, `undefined: Dockerfile`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-03-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/functionservice/functionbuild/`
Expected: `ok`

Run: `golangci-lint run ./hivepaas_app/service/functionservice/... ./hivepaas_app/hperrors/ && go run ./tools/errcodelint`
Expected: `0 issues.`; errcodelint prints nothing and exits 0

Optional, with Docker (about a minute after the pulls): `HIVEPAAS_DOCKER_TESTS=1 go test -count=1 -run TestAFunctionOfEachRuntimeBuildsOnTheReleasesImages -v ./hivepaas_app/service/functionservice/functionbuild/`
Expected: `--- PASS` for `node24`, `python313` and `go127`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(functions): the Dockerfile HivePaaS writes for a function

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 4: A function is deployed from its code, built on its runtime's image

**Files:**
- Create: `hivepaas_app/service/appdeploymentservice/appdeploymentserviceimpl/function_deploy.go`, `.../function_deploy_apply_svc.go`
- Modify: `hivepaas_app/base/app_deployment.go`, `hivepaas_app/service/functionservice/functionbuild/functionbuild.go` (`ContentHash`), `.../appdeploymentserviceimpl/{deployment.go,deployment_notification.go,repo_deploy.go,repo_deploy_build_image.go}`, `hivepaas_app/interface/mcp/tools_app_actions.go`
- Test: `.../appdeploymentserviceimpl/function_deploy_test.go`, `hivepaas_app/service/functionservice/functionbuild/functionbuild_test.go`

**Interfaces:**
- Consumes: Task 3's `functionbuild.Dockerfile`, `WriteInlineCode`, `DockerfilePath`; Task 1's release images through `systemappservice.CurrentRelease().FunctionRuntimes`.
- Produces:
  - `base.DeploymentMethodFunction = "function"`, in `AllDeploymentMethods`;
  - `func functionbuild.ContentHash(code *entity.FunctionInlineCode, dockerfile string) string` (64 hexadecimal characters);
  - in `appdeploymentserviceimpl`: `repoDeploymentData.ContextDir` (the build's context when it is not the whole checkout); `deployStepPrepareBuild(ctx, db, data, pushToRegistry entity.ObjectID)` and `deployStepImageBuild(ctx, db, data, target *imageBuildTarget)` with `imageBuildTarget{CommitHash, Dockerfile, PushToRegistry, Inputs}`, the repository's steps cut so that a function's use them; `deployFromFunction`, `functionDeployStepSource`, `functionDeployStepImageBuild`, `functionDeployStepServiceApply`; `applyFunctionContainer(contSpec, source)` (moved in Task 8);
  - the MCP's `deploySource.describe()` says "the function's code" for a function without a repository.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-04-tests.patch`

Tests added: `TestAFunctionsInlineCodeIsTheSourceOfItsBuild`, `TestAFunctionInARepositoryIsBuiltFromItsDirectory`, `TestAFunctionsDirectoryThatIsNotInItsRepositoryFailsTheDeployment`, `TestAFunctionIsBuiltFromTheDockerfileWrittenForIt`, `TestAFunctionsContainerIsItsRuntimes`, `TestTheSameFunctionHasTheSameContentHash`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/service/appdeploymentservice/... ./hivepaas_app/service/functionservice/...`
Expected: FAIL to build: `undefined: ContentHash` in `functionbuild`, `undefined: base.DeploymentMethodFunction` in `appdeploymentserviceimpl`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-04-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/appdeploymentservice/... ./hivepaas_app/service/functionservice/...`
Expected: `ok` for `appdeploymentserviceimpl` and `functionbuild`

Run: `golangci-lint run ./hivepaas_app/service/appdeploymentservice/... ./hivepaas_app/service/functionservice/... ./hivepaas_app/interface/mcp/ ./hivepaas_app/base/`
Expected: `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(functions): a function is deployed from its code, built on its runtime's image

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 5: An app's settings know a function

**Files:**
- Create: `hivepaas_app/usecase/appsettingsuc/function_rules.go`, `hivepaas_app/usecase/appsettingsuc/appsettingsdto/deployment_function_source_get.go`, `hivepaas_app/service/appdeploymentservice/appdeploymentserviceimpl/build_source_check.go`
- Modify: `hivepaas_app/base/{app_kind.go,env_var.go}`, `hivepaas_app/entity/app_kind_settings.go`, `hivepaas_app/hperrors/errors_project_app.go`, `hivepaas_app/pkg/translation/messages/en/errors.project_app.en.toml`, `hivepaas_app/service/appdeploymentservice/{service.go,types.go}`, `hivepaas_app/service/appcloneservice/appcloneserviceimpl/clone_2_settings.go`, `hivepaas_app/service/envvarservice/envlink/recipes.go`, `hivepaas_app/usecase/appsettingsuc/appsettingsdto/{deployment_settings_get.go,deployment_settings_update.go,kind_settings_update.go}`, `hivepaas_app/usecase/appsettingsuc/{deployment_settings_update.go,kind_settings_update.go,routing_settings_update.go}`, `hivepaas_app/interface/mcp/tools_app_actions.go`, `docs/openapi/swagger.json`
- Test: `hivepaas_app/entity/app_kind_function_test.go`, `hivepaas_app/usecase/appsettingsuc/function_rules_test.go`, `hivepaas_app/usecase/appsettingsuc/appsettingsdto/deployment_settings_function_test.go`, `hivepaas_app/service/appdeploymentservice/appdeploymentserviceimpl/build_source_check_test.go`, `hivepaas_app/service/appcloneservice/appcloneserviceimpl/clone_2_settings_test.go`, `hivepaas_app/service/envvarservice/envlink/envlink_test.go`, `hivepaas_app/interface/mcp/tools_function_test.go`

**Interfaces:**
- Consumes: Task 2's source and request, Task 4's method.
- Produces:
  - `base.AppCategoryFunction = "function"`, in `AllAppCategories`; a function shares no variables of its kind, and another app links to it by its address;
  - `func entity.IsFunctionKind(setting *entity.Setting) bool`;
  - `appdeploymentservice.Service.CheckBuildSource(ctx, *CheckBuildSourceReq) error` with `CheckBuildSourceReq{RepoSource *entity.DeploymentRepoSource; PushToRegistry entity.ObjectID; RefObjects *entity.RefObjects}`: the registry a cluster of several nodes needs, then the repository and its ref (`gittool`), the check a repository's settings made in the use case before;
  - `DeploymentSettingsReq.FunctionSource *DeploymentFunctionSourceReq` (`json:"functionSource"`), normalized by `UpdateAppDeploymentSettingsReq.ModifyRequest`, required (`ERR_VLD_VALUE_REQUIRED` at `functionSource`) when the method is `function`;
  - `DeploymentSettingsResp.FunctionSource *DeploymentFunctionSourceResp` (`json:"functionSource,omitempty"`), its registry and credentials answered as settings (missing ones as missing);
  - `UpdateAppKindSettingsReq.ModifyRequest`: a function's port is 8080;
  - in `appsettingsuc`: `checkDeploymentOfKind(function bool, settings *entity.AppDeploymentSettings) error`, `checkKindCategoryChange(current *entity.AppKindSettings, next base.AppCategory) error`, `fixFunctionRouting(routing *entity.AppRoutingSettings)`;
  - errors `ErrDeploymentMethodFunctionRequired`, `ErrDeploymentMethodFunctionUnallowed` (412), `ErrAppKindFunctionUnchangeable` (412, from `ErrNonEditable`);
  - the MCP's `deploySource.Code`, a digest of a function's inline code, and the repository's URL, ref and commit for one in a repository.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-05-tests.patch`

Tests added: `TestAnAppIsAFunctionByItsKind`, `TestAFunctionIsDeployedAsOne`, `TestAFunctionStaysAFunction`, `TestAFunctionsRoutingPointsAtItsRuntimesPort`, `TestAFunctionsDeploymentSettingsCarryItsSource`, `TestTheFunctionMethodWithoutASourceIsRefused`, `TestAFunctionsSourceIsSentWithItsSettings`, `TestAFunctionsKindKeepsItsRuntimesPort`, `TestABuildOnSeveralNodesNeedsARegistry`, `TestAFunctionsCloneIsAFunctionWhenItTakesItsCode`, `TestAFunctionIsLinkedToByItsAddress` (and a `function` case in two envlink tables), `TestARedeployOfAFunctionSeesItsCode`, `TestARedeployOfAFunctionInARepositorySeesItsCommit`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/entity/ ./hivepaas_app/interface/mcp/ ./hivepaas_app/service/appcloneservice/... ./hivepaas_app/service/appdeploymentservice/... ./hivepaas_app/service/envvarservice/... ./hivepaas_app/usecase/appsettingsuc/...`
Expected: FAIL to build in every package, among them `undefined: IsFunctionKind`, `undefined: base.AppCategoryFunction`, `unknown field FunctionSource in struct literal of type appsettingsdto.DeploymentSettingsResp`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-05-code.patch`

What the use cases now do:
- **Deployment settings:** the app is loaded with its kind; `checkDeploymentOfKind` refuses another method for a function (`ERR_DEPLOYMENT_METHOD_FUNCTION_REQUIRED`), and the method `function` or a function's source for another app (`ERR_DEPLOYMENT_METHOD_FUNCTION_UNALLOWED`); the repository's and the function's build are checked by `CheckBuildSource`.
- **Kind settings:** `checkKindCategoryChange` refuses a kind that makes a function another app, or another app a function, a new kind included (`ERR_APP_KIND_FUNCTION_UNCHANGEABLE`).
- **Routing settings:** a function's port and its domains' container ports are 8080 whatever the request says.
- **Clone:** a function's kind is copied when its deployment settings are.

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/entity/ ./hivepaas_app/interface/mcp/ ./hivepaas_app/service/appcloneservice/... ./hivepaas_app/service/appdeploymentservice/... ./hivepaas_app/service/envvarservice/... ./hivepaas_app/usecase/appsettingsuc/...`
Expected: every package `ok`

Run: `golangci-lint run ./hivepaas_app/... && go run ./tools/errcodelint`
Expected: `0 issues.`; errcodelint prints nothing and exits 0 (the exhaustive linter checks the new category in every switch over categories)

Run: `make gen-swag && git status --short docs/`
Expected: no output (the patch carries the regenerated `swagger.json`)

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(functions): an app's settings know a function

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 6: Creating a function

**Files:**
- Create: `hivepaas_app/usecase/appuc/create_function.go`, `hivepaas_app/usecase/appuc/appdto/create_function.go`, `hivepaas_app/interface/api/handler/apphandler/create_function.go`
- Modify: `hivepaas_app/usecase/appuc/uc.go` (the task queue and the deployment service), `hivepaas_app/interface/api/server/router_apps.go`, `docs/openapi/swagger.json`
- Test: `hivepaas_app/usecase/appuc/create_function_test.go`, `hivepaas_app/usecase/appuc/appdto/create_function_test.go`, `hivepaas_app/interface/api/server/router_apps_test.go`

**Interfaces:**
- Consumes: Task 2's `DeploymentFunctionSourceReq`, Task 5's `CheckBuildSource` and `base.AppCategoryFunction`, `appprovisionservice.ProvisionApps` with `Configure` and `FirstDeployment`.
- Produces:
  - `POST /projects/{projectID}/{projectEnv}/apps/function` → `apphandler.Handler.CreateFunction` → `appuc.UC.CreateFunction(ctx, auth, *appdto.CreateFunctionReq) (*appdto.CreateFunctionResp, error)`, 201;
  - `appdto.CreateFunctionReq{ProjectID, ProjectEnvID, *AppBaseReq, Source *appsettingsdto.DeploymentFunctionSourceReq}` (`source` required); `CreateFunctionResp{Meta, Data *CreateFunctionDataResp{ID, DeploymentID, TaskID}}`;
  - `appuc.New` takes `queue.TaskQueue` and `appdeploymentservice.Service` too (fx provides both);
  - in `appuc`: `checkFunctionSource(ctx, projectID, projectEnvID, source) error` (the source's references in its environment, then `CheckBuildSource`), `functionSettings(app, source, timeNow) []*entity.Setting` (kind `function`; deployment `function` with the source, inheritable; routing at 8080, inheritable).

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-06-tests.patch`

Tests added: `TestAFunctionIsCreatedFromItsSource`, `TestAFunctionWithoutASourceIsRefused`, `TestAFunctionsSourceIsCheckedAsItsSettingsWouldBe`, `TestAFunctionWithoutANameIsRefused`, `TestAFunctionsSourceIsCheckedInItsEnvironmentFirst`, `TestAFunctionIsCreatedWithItsKindSourceAndPort`, `TestAFunctionIsCreatedAmongTheEnvsApps`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/interface/api/server/ ./hivepaas_app/usecase/appuc/... ./hivepaas_app/cmd/internal/`
Expected: FAIL: `undefined: CreateFunctionReq`, `unknown field appDeploymentService in struct literal of type UC`, and `--- FAIL: TestAFunctionIsCreatedAmongTheEnvsApps` (the route is not there)

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-06-code.patch`

`CreateFunction` checks the source before the transaction (the repository check reaches the network, and nothing is locked meanwhile), provisions the app with `functionSettings` and a first deployment, records the creation, schedules the deployment and certificate tasks once committed, and removes what provisioning made in docker when the transaction does not commit.

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/interface/api/server/ ./hivepaas_app/usecase/appuc/... ./hivepaas_app/cmd/internal/`
Expected: every package `ok` (`cmd/internal` validates the fx graph with `appuc.New`'s new parameters)

Run: `golangci-lint run ./hivepaas_app/usecase/appuc/... ./hivepaas_app/interface/api/...`
Expected: `0 issues.`

Run: `make gen-swag && git status --short docs/`
Expected: no output

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(functions): creating a function

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 7: A function is not part of a spec yet

**Files:**
- Modify: `hivepaas_app/service/specservice/specmodel/issue.go`, `hivepaas_app/service/specservice/specserviceimpl/{build_settings.go,walk.go}`
- Test: `hivepaas_app/service/specservice/specserviceimpl/{build_test.go,build_import_test.go,export_test.go}`

**Interfaces:**
- Consumes: Task 5's `entity.IsFunctionKind`, `base.AppCategoryFunction`, `base.DeploymentMethodFunction`.
- Produces: `specmodel.CodeFunctionSkipped = "FUNCTION_SKIPPED"`, an export issue (`skipped`, at the app's path); a template's or an import's kind `function`, and an import's deployment settings with the method `function` or a function's source, refused with `ERR_SPEC_BLOCK_INVALID`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-07-tests.patch`

Tests added: `TestExportSkipsAFunction` (the export fixture gains a function, `hello`), `TestBuildAppDoesNotImportAFunction`, and the case `a function` of `TestBuildAppRefuses`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/service/specservice/...`
Expected: FAIL, `undefined: specmodel.CodeFunctionSkipped`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-07-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/specservice/...`
Expected: two `ok`

Run: `golangci-lint run ./hivepaas_app/service/specservice/...`
Expected: `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(functions): a function is not part of a spec yet

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 8: A function's container settings keep what is fixed for it

**Files:**
- Create: `hivepaas_app/service/functionservice/functioncontainer/functioncontainer.go`
- Modify: `hivepaas_app/service/appdeploymentservice/appdeploymentserviceimpl/function_deploy_apply_svc.go`, `hivepaas_app/usecase/appsettingsuc/container_settings_update.go`
- Test: `hivepaas_app/service/functionservice/functioncontainer/functioncontainer_test.go` (`TestAFunctionsContainerIsItsRuntimes`, moved from `function_deploy_test.go`), `hivepaas_app/usecase/appsettingsuc/container_settings_function_test.go`

**Interfaces:**
- Consumes: Task 4's `applyFunctionContainer`, which moves; Task 5's `entity.IsFunctionKind`.
- Produces: `functioncontainer.StopGraceMargin` (10 s) and `functioncontainer.ApplyFixed(contSpec *swarm.ContainerSpec, source *entity.DeploymentFunctionSource)`; `updateAppContainerSettingsData.FunctionSource`, loaded with the app's kind and deployment settings.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-08-tests.patch`

Tests added: `TestAFunctionsContainerIsItsRuntimes` (in its new package), `TestAFunctionsContainerSettingsKeepWhatIsFixedForIt`, `TestAnAppsContainerSettingsAreTheRequests`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test -count=1 ./hivepaas_app/service/functionservice/... ./hivepaas_app/usecase/appsettingsuc/ ./hivepaas_app/service/appdeploymentservice/...`
Expected: FAIL to build: `undefined: ApplyFixed`, `unknown field FunctionSource in struct literal of type updateAppContainerSettingsData`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-08-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/functionservice/... ./hivepaas_app/usecase/appsettingsuc/ ./hivepaas_app/service/appdeploymentservice/...`
Expected: every package `ok`

- [ ] **Step 5: The whole repository, and the tree**

Run: `golangci-lint run ./... && go run ./tools/goroutinelint . && go run ./tools/errcodelint`
Expected: `0 issues.`; the two custom linters print nothing and exit 0

Run: `go test ./...`
Expected: every package `ok` (no `FAIL`)

Run: `make gen-swag && git status --short`
Expected: no output

Run: `git diff --stat prep/functions-backend -- . ':(exclude)docs/superpowers'`
Expected: no output

- [ ] **Step 6: Commit**

```bash
git commit -m "feat(functions): a function's container settings keep what is fixed for it

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## After the plan (the user)

- Merge into `main` locally; run `go test ./...` in the main checkout, whose `vendor/` the worktree did not have.
- The live check (spec §10), on the Linux server, once the dashboard can create one or through the API: a function of each runtime created inline and deployed, called over HTTP; one from a repository with a private library (a build secret); one with a Debian package; a cluster of two nodes with a registry.
- Part 3 starts from here: the test run (API, agent, the libraries' image) and the lock file it returns.
