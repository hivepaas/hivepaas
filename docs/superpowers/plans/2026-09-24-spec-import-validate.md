# Spec Import: Validate Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `POST .../spec/import/validate` at global, project and env scope reads an uploaded
bundle, matches what it holds against the installation, and answers with the plan: every node,
what import would do to it, what changed, whether a running app would restart, the issues found so
far, and the `planHash` apply will be bound to.

**Architecture:** Three layers, each testable alone. `specmodel` gains the plan's types and the
selection grammar. `specserviceimpl` gains a bundle reader (base64 → age → tar.gz → documents),
the matcher (id, then key, per object kind) and the planner, which diffs the uploaded documents
against the same documents `buildBundle` writes for the target - so both sides are rendered by one
piece of code. `specuc` and `spechandler` add the endpoint, its permissions and its body limit.

**Out of this plan, into the next:** reference resolution and closure (`REF_NOT_SELECTED`,
`REF_NOT_FOUND`), the availability issues (domains, ports, storage, nodes, capabilities, shared
mounts), owner and credential notes, and apply.

**Spec:** [docs/superpowers/specs/2026-09-23-config-spec-import-design.md](../specs/2026-09-23-config-spec-import-design.md)
- §1 selection, §2 transport, §3 API, §4 matching, §5 validate.

## Global Constraints

- Before a task is done: `go build ./...`; `golangci-lint run ./...` printing `0 issues.`;
  `go test ./hivepaas_app/service/specservice/... ./hivepaas_app/usecase/specuc/...` whole. The last
  task runs `go test ./...` and `make gen-swag`.
- The body is capped at 25 MB with `http.MaxBytesReader`. The bundle arrives base64 in JSON; the
  passphrase is never stored or logged.
- A bundle is refused whole - `ERR_SPEC_PASSPHRASE_REQUIRED`, `ERR_SPEC_PASSPHRASE_INVALID`,
  `ERR_SPEC_BUNDLE_INVALID`, `ERR_SPEC_API_VERSION_UNSUPPORTED` - only when it cannot be read;
  anything about one object is an issue on its node.
- Node paths: `global`, `projects/<p>`, `projects/<p>/settings`, `projects/<p>/envs/<e>`,
  `projects/<p>/envs/<e>/settings`, `projects/<p>/envs/<e>/apps/<a>`. A path selects its node and
  every node below; `*` matches one segment; exclude wins; an empty include selects everything.
- Matching: project by id then key; env by key within its project (its id is derived); app by id
  then key within its env. A project, env or app whose id matches a target object with another key
  is `KEY_MISMATCH`, skipped. A project to be created whose name another project holds is
  `NAME_IN_USE`, skipped.
- Diff: both sides parsed from YAML; a collection entry's `id`, and the `id` inside an `external`
  reference, are left out of the comparison.
- `planHash` is the SHA-256 of the bundle bytes' digest, the selection, the options, and each
  node's path, action, restart, deploy and issue codes, sorted by path.
- End every commit message with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

---

### Task 1: Plan types and selection (`specmodel/importplan.go`)

Types: `Selection{Include, Exclude}` with `Selects(path) bool`; `ImportOptions{Existing,
DeployCreated, DeployChangedSource}` with `ExistingUpdate`/`ExistingKeep`; node kinds and
actions; `PlanNode`; `ImportPlan{Bundle BundleInfo, Nodes, Summary, PlanHash}`; issue codes
`KEY_MISMATCH`, `NAME_IN_USE`, `SETTING_VERSION_NEWER`, `TYPE_NOT_IMPORTABLE`; `SeverityWarning`.
Tests: every row of the spec's selection table, `*`, exclude winning, an ancestor not selected by a
deeper include.

### Task 2: Reading a bundle (`specserviceimpl/bundle_read.go`)

`readBundle(content []byte, passphrase string) (*specmodel.ImportBundle, error)`: age-decrypts when
the content starts with the age header, gunzips, reads the tar, refuses an entry path that is
absolute or leaves the archive, caps the file count and each file's size, parses `spec.yaml`,
refuses an `apiVersion` it does not read, and parses `global.yaml`, `projects/*/project.yaml` and
`projects/*/envs/*.yaml`. `parseBundleFiles(files)` does the parsing for both the upload and the
target's own export. Tests: a round trip through `Export` for each secrets mode; a wrong
passphrase; a missing passphrase; garbage; a path escaping the archive; a future `apiVersion`.

### Task 3: Planning (`specserviceimpl/import_plan.go`, `import_match.go`, `import_diff.go`)

`(*service).ValidateImport(ctx, db, req) (*specservice.ValidateImportResp, error)` with
`ValidateImportReq{Scope, Bundle []byte, Passphrase, Selection, Options}`:

1. read the bundle; restrict it to the route's scope (project and env routes find their project and
   env in the bundle by id then key, else `ErrSpecImportScopeNotInBundle`);
2. export the target at the route's scope in memory with the bundle's secrets mode and parse it;
3. build the nodes, match each, diff matched ones, set `restart` and `deploy`, add the issues this
   plan covers, apply the selection and the `existing: keep` option;
4. compute the summary and `planHash`.

Tests against the export fixture: an unchanged import is all `unchanged` with no restart; a changed
env var makes the app `update` with `settings.envVars` in its changes and `restart`; a new app is
`create` with `deploy` when `deployCreated`; `keep`; `KEY_MISMATCH`; `SETTING_VERSION_NEWER`;
`TYPE_NOT_IMPORTABLE`; the same request twice gives the same `planHash`, a different option another.

### Task 4: The endpoint

`specdto.ValidateImportReq` (`bundle` base64, `passphrase`, `selection`, `options`) and its response;
`specuc.ValidateImport` - write access at the route's scope, `AuthorizeSecretReveal` when the
bundle's secrets mode reveals secrets; `spechandler` validate handlers for the three scopes with the
25 MB cap; routes beside export; the new errors in `errors_spec.go` and their messages; swagger.
Tests: the usecase refuses a secret-bearing bundle without the reveal permission.

### Task 5: Record, verify

The import spec §14 names this plan; `go build ./... && golangci-lint run ./... && go test ./...`;
`make gen-swag`; commit.
