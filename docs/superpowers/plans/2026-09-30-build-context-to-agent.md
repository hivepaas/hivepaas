# The Build Context Goes to the Build Node Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A deployment from a repository builds on another node: the app sends the checkout to that node's agent inside the build call, instead of the path of a directory the agent cannot see.

**Architecture:**
- **Packing:** `pkg/srcpack` writes a source tree as a tar stream compressed with zstd, and unpacks one without trusting it.
- **The call:** `ImageBuild` is replaced by `ImageBuildFromSource`, a stream in both directions: the request, then the source in chunks of 1 MiB; logs and the result come back.
- **The agent:** unpacks the source in `/tmp/hivepaas/<YYYY-MM-DD>/<random>/checkout`, runs the build that exists today on it, and removes the directory when the call ends.
- **Left-overs:** the agent removes dated temporary directories when it starts and at each node cleanup, with the function the app's cleanup now shares.

**Tech Stack:** Go (gRPC with protoc 35 and protoc-gen-go, `archive/tar`, `klauspost/compress/zstd`, already vendored).

**Spec:** `docs/superpowers/specs/2026-09-30-build-context-to-agent-design.md`

## How the patches work

Every task's code was written and verified before this plan.

**Base:** the patches apply in order to `main` at `32da4c26`. This plan's commit changes only `docs/`.

**What was checked, on a fresh worktree, at every task:**
- the tests alone fail (a build error: the code they call does not exist yet);
- with the code, the task's packages pass;
- `golangci-lint run ./...` finds 0 issues, and `go run ./tools/goroutinelint .` and `go run ./tools/errcodelint` pass.

**After the last task:** `go test ./...` passes; `make gen-swag` and `go generate ./hivepaas_app/interface/agent/proto/...` change nothing; the tree equals the branch the patches were cut from. The new tests also pass five times under `-race`.

Patches live in `docs/superpowers/plans/2026-09-30-build-context-to-agent/`: `be-NN-tests.patch` then `be-NN-code.patch` for task NN.

The commands below run from the backend repo root, with `P=docs/superpowers/plans/2026-09-30-build-context-to-agent`.

## Global Constraints

- **Before a task is done:** `go build ./...`; `golangci-lint run ./...` over the whole repo; `go run ./tools/goroutinelint .`; `go run ./tools/errcodelint`; the task's tests.
- **No import the main checkout has not vendored**, above all in tests: `testify/assert` is vendored, `testify/require` is not. After the merge, `go test ./...` runs in the main checkout too.
- **The whole checkout is sent**, not filtered by `.dockerignore`.
- **The source on the wire:** tar, compressed with zstd at its fastest level, in chunks of 1 MiB.
- **The agent's copy:** `/tmp/hivepaas/<YYYY-MM-DD>/<random>/checkout`, directories of mode `0700`, removed when the call ends.
- **Unpacking refuses** an absolute path, a path with `..`, an entry written through or over a symbolic link, and any entry that is not a file, a directory or a symbolic link.
- **Dated temporary directories** are kept 3 days by the cleanups; an agent that starts removes them all.
- **A build on the app's own node does not change.**
- **Generated code:** `image_build.pb.go` and `image_build_grpc.pb.go` come from `go generate ./hivepaas_app/interface/agent/proto/...` (protoc 35.1); they are in the code patch.
- **Git:** commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`; merge locally, never push; never `git stash`; leave the user's uncommitted work alone.

## Review Focus

1. **A real build on another node.** No test runs buildx; the tests stop at "the build is given a directory holding the source".
   - Pinned by: `TestAnAgentBuildsTheSourceItIsSent` (Task 3, over a real gRPC connection, with a 3 MiB file that does not compress), and the live check of Task 3, step 6.
2. **An agent of an older image.** It must answer `Unimplemented` at once, so that the deployment falls back to the current node rather than fail.
   - Pinned by: `TestAnAgentWithoutTheCallSaysSo` (Task 3). The fallback itself, and its warning in the deployment's log, by the live check.
3. **A deployment canceled while the source is being sent or built.** The agent's copy must go.
   - Pinned by: `TestACanceledBuildLeavesNoSourceOnTheAgent`, `TestASourceThatFailsToPackIsNotBuilt` (Task 3).
4. **A hostile or broken source stream.** Nothing may be written outside the build's directory.
   - Pinned by: `TestAnEntryLeavingTheDirectoryIsRefused` (five cases), `TestAnEntryThatIsNotSourceIsRefused`, `TestACutStreamIsAnError` (Task 2).
5. **A large repository.** The agent's copy is in the agent container's own filesystem, and the build node's disk holds it for the length of the build.
   - Pinned by: the live check, with the size the deployment's log reports ("Source sent: N files, X").

## Decisions beyond the spec

- **The client takes a function, `sendSource func(w io.Writer) error`**, not a reader: the app packs straight into the stream. Each chunk is copied before it is sent, since gRPC may still use a message after `Send` returns.
- **When packing fails on the app, that is the error the deployment gets**; the agent only sees the call canceled.
- **`imagebuildagentuc.UC.WithTempBaseDir`** sets where the agent unpacks; the agent itself uses `/tmp/hivepaas`.
- **An entry whose place is already taken is refused** (a duplicate in the stream), and a file's modification time travels with it.
- **`fileutil.TempDirRetentionDays = 3`** names the rule the app's cleanup had inline.
- **A bug the shared function fixes:** the app's cleanup stopped at the first base directory that did not exist and never looked at the second.
- **The agent's sweep at node cleanup runs only when the cluster cleanup is enabled**, as the rest of that call.
- **The size and time of the transfer are logged by the app**, not by the agent.

---

## Task 1: Dated temporary directories are removed by one function, on the agent too

**Files:**
- Modify: `hivepaas_app/pkg/fileutil/temp_dir.go`, `hivepaas_app/service/syscleanupservice/syscleanupserviceimpl/sys_cleanup_files.go`, `hivepaas_app/usecaseagent/nodecleanupagentuc/node_cleanup.go`, `hivepaas_app/cmd/agent/main.go`
- Test: `hivepaas_app/pkg/fileutil/temp_dir_test.go`

**Interfaces:**
- Produces:
  - `fileutil.RemoveDatedTempDirs(baseDir string, before time.Time) (removed int, err error)`;
  - `fileutil.TempDirRetentionDays = 3`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-01-tests.patch`

Tests added: `TestRemoveDatedTempDirsRemovesTheDaysBeforeTheThreshold`, `TestRemoveDatedTempDirsOfAMissingBaseIsNothing`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/pkg/fileutil/`
Expected: FAIL, build error `undefined: RemoveDatedTempDirs`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-01-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/pkg/fileutil/ ./hivepaas_app/service/syscleanupservice/... ./hivepaas_app/cmd/... && golangci-lint run ./...`
Expected: PASS; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(cleanup): an agent removes the temporary directories of past days

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 2: Packing a source tree, and unpacking one without trusting it

**Files:**
- Create: `hivepaas_app/pkg/srcpack/srcpack.go`
- Test: `hivepaas_app/pkg/srcpack/srcpack_test.go`

**Interfaces:**
- Produces:
  - `srcpack.Stats{Files int; Bytes int64}`;
  - `srcpack.Pack(ctx context.Context, dir string, w io.Writer) (Stats, error)`;
  - `srcpack.Unpack(ctx context.Context, r io.Reader, dir string) (Stats, error)`, `dir` existing.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-02-tests.patch`

Tests added: `TestASourceTreeArrivesAsItLeft`, `TestAnEntryLeavingTheDirectoryIsRefused`, `TestAnEntryThatIsNotSourceIsRefused`, `TestACutStreamIsAnError`, `TestPackingStopsWhenCanceled`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/pkg/srcpack/`
Expected: FAIL, build errors `undefined: Pack` and `undefined: Unpack`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-02-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/pkg/srcpack/ && golangci-lint run ./...`
Expected: PASS; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(build): a source tree is packed to travel, and unpacked without trusting it

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 3: The checkout travels in the build's call

**Files:**
- Create: `hivepaas_app/usecaseagent/imagebuildagentuc/image_build_source.go`
- Modify:
  - `hivepaas_app/interface/agent/proto/image_build.proto`, and the generated `image_build.pb.go`, `image_build_grpc.pb.go`;
  - `hivepaas_app/usecaseagent/imagebuildagentuc/uc.go`;
  - `hivepaas_app/interface/agent/server/imagebuildservice/image_build.go`, `hivepaas_app/interface/agent/server/server_image_build.go`;
  - `hivepaas_app/interface/agent/client/imagebuildservice/service.go`;
  - `hivepaas_app/service/appdeploymentservice/appdeploymentserviceimpl/repo_deploy_build_image.go`.
- Test: `hivepaas_app/usecaseagent/imagebuildagentuc/image_build_source_test.go`, `hivepaas_app/interface/agent/client/imagebuildservice/service_test.go`

**Interfaces:**
- Consumes: Task 2's `srcpack.Pack` and `srcpack.Unpack`; `fileutil.CreateTempDir`.
- Produces:
  - gRPC: `ImageBuildService.ImageBuildFromSource(stream ImageBuildMsg) returns (stream ImageBuildResp)`; `ImageBuildMsg{oneof: req ImageBuildReq, source_chunk bytes}`; `ImageBuildReq` fields 10 and 11 reserved; the `ImageBuild` rpc is gone;
  - agent use case: `(*imagebuildagentuc.UC).ImageBuildFromSource(ctx, req *imagebuildagentdto.ImageBuildReq, source io.Reader) (*imagebuildagentdto.ImageBuildResp, error)` and `(*UC).WithTempBaseDir(dir string) *UC`;
  - server: `imagebuildservice.ImageBuildFromSource(uc, stream)` (package `interface/agent/server/imagebuildservice`);
  - client: `ImageBuildServiceClient.ImageBuild(ctx, req, sendSource func(w io.Writer) error) (*imagebuildagentdto.ImageBuildResp, error)`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-03-tests.patch`

Tests added:
- agent use case: `TestTheBuildRunsOnTheSourceItWasSent`, `TestTheSourceIsRemovedAfterAFailedBuild`, `TestASourceThatDoesNotUnpackIsNotBuilt`;
- the call: `TestAnAgentBuildsTheSourceItIsSent`, `TestAFailedBuildGivesTheAgentsReason`, `TestACanceledBuildLeavesNoSourceOnTheAgent`, `TestASourceThatFailsToPackIsNotBuilt`, `TestAnAgentWithoutTheCallSaysSo`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/usecaseagent/imagebuildagentuc/ ./hivepaas_app/interface/agent/client/imagebuildservice/`
Expected: FAIL, build errors `type *UC has no field or method WithTempBaseDir` and `too many arguments in call to c.ImageBuild`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-03-code.patch`

The patch carries the generated proto code. `go generate ./hivepaas_app/interface/agent/proto/...` must leave it unchanged.

- [ ] **Step 4: Check everything**

Run: `go build ./... && go test ./... && golangci-lint run ./... && go run ./tools/goroutinelint . && go run ./tools/errcodelint`
Expected: PASS; `0 issues.` `make gen-swag` changes nothing.

Run: `go test -race -count=5 ./hivepaas_app/interface/agent/client/imagebuildservice/ ./hivepaas_app/usecaseagent/imagebuildagentuc/ ./hivepaas_app/pkg/srcpack/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(build): the checkout travels to the build node in the build's call

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

- [ ] **Step 6: Live checks** (the user, on the Linux server, with new app and agent images)
  - **Check 1:** with two nodes, set the image build settings' worker to the other node and deploy an app from a repository with a Dockerfile: the log says "Starting build process on worker node", "Sending the source to the build node...", "Source sent: N files, X, in T", and the build succeeds (Review Focus 1 and 5).
  - **Check 2:** the same with a generated Dockerfile (source "auto").
  - **Check 3:** on the build node, during the build `/tmp/hivepaas/<today>/` in the agent's container holds one directory; after it, none.
  - **Check 4:** cancel a deployment during the build: the directory goes (Review Focus 3).
  - **Check 5:** with the agent of the build node still on the old image, deploy: the log warns and the build falls back to the current node (Review Focus 2).
