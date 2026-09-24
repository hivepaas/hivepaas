# Spec Import: Apply, Part 1 - Scopes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `specservice.ApplyImport` plans the request again inside the caller's transaction,
refuses a plan that changed, is blocked or has issues nobody accepted, and writes what the plan
says for everything but apps: global settings, projects - created with the defaults the dashboard
gives one - envs, and their settings, with every reference remapped to this installation's ids
and every issue's fix applied.

**Architecture:** Project creation's defaults move from `projectuc` to `projectservice`, where
import calls them too (spec §10). Settings of a scope are written through import policies, a
registry keyed by setting type beside the export's: the default writes the row, `ssl-cert` also
writes its certificate files after the commit, and a type whose writing has effects import does
not reproduce yet is skipped with its reason - `TYPE_NOT_IMPORTABLE` in the plan. Ids are chosen
for everything created before the first write, and references are rewritten with `RemapRefs`.

**Out of this plan, into the next:** apps - created through `ProvisionApps`, updated in phase 2
through the builders and `ApplyAppConfiguration` - deployments, the usecase with its transaction
and audit record, the handlers and routes. Until then app nodes are left alone by apply, and no
route reaches it.

**Spec:** [docs/superpowers/specs/2026-09-23-config-spec-import-design.md](../specs/2026-09-23-config-spec-import-design.md)
- §8 phase 1 steps 1-4 and the refusals, §10 projects, envs and import policies, §6 fixes.

## Global Constraints

- Before a task is done: `go build ./...`; `golangci-lint run ./...` printing `0 issues.`;
  `go test ./hivepaas_app/service/specservice/... ./hivepaas_app/service/projectservice/...
  ./hivepaas_app/usecase/projectuc/...`. The last task runs `go test ./...`.
- Apply refuses, in this order: `ERR_SPEC_IMPORT_PLAN_CHANGED` when the recomputed `planHash`
  differs from the one sent; `ERR_SPEC_IMPORT_BLOCKED` when anything selected is blocked;
  `ERR_SPEC_IMPORT_ISSUES_NOT_ACCEPTED` when anything selected has an issue and `acceptIssues`
  is false.
- Only selected nodes whose action is `create` or `update` are written; in a node, only what
  its `changes` name, except that a node being created writes everything it holds.
- A setting that exists keeps its id, scope, creation time, and - in an `omit` bundle - every
  secret the bundle leaves empty. A setting created gets a new id.
- A reference that validate cleared (`REF_NOT_SELECTED`, `REF_NOT_FOUND`) is written empty, and
  the setting holding it is `pending`; so is a setting with `SECRET_OMITTED`.
  `NODE_NOT_FOUND` writes the volume unpinned and leaves it active.
- A project created takes the bundle's key, name and note, the resolved owner or the operator,
  the defaults `projectuc.CreateProject` gives, and then the bundle's project settings over
  those defaults, matched by type and key.
- Every node written gets `outcome: applied`; a selected node not written, `unchanged` or
  `skipped`.
- End every commit message with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

---

### Task 1: A new project's defaults move to `projectservice`

`projectservice.PrepareNewProject(ctx, req *NewProjectReq, out *PersistingProjectData) error`
builds the project row, its env rows, tags, default webhook, default notification and default
volume, from `NewProjectReq{Project, Envs []NewProjectEnv{Name, Color}, Tags}`; and
`projectservice.NewProjectEnv(project, name, color, index, now) *entity.ProjectEnv` builds one
env row, for project update and import. `projectuc.CreateProject` and `UpdateProject` call them;
their tests pass unchanged.

### Task 2: Import policies (`import_policy.go`)

`importPolicyFor(typ) (importPolicy, reason string)`. The default policy writes the row;
`ssl-cert`'s also writes its certificate files once committed (`sslService.WriteCertFiles`).
Skipped, each with its reason: `schedJobs`, `periodicJobs`, `backupRepoCleanup`, `sslRenewal`,
`systemBackup`, `systemCleanup` (writing one schedules tasks), `backupRepos` (writing one
initializes a repository), `logging`, `registry`, `traefikConfig`, `traefikService`,
`hivepaasService` (each runs a service HivePaaS manages), `networks` (created with an env's
first app, under this installation's name). A test holds every type export can write to one or
the other.

Validate raises `TYPE_NOT_IMPORTABLE` (skipped) with the reason for a setting of a skipped type
that the node would write, and leaves it out of the node's changes. Settings of apps are built
by `BuildApp`, not by these policies.

### Task 3: `ApplyImport` and its refusals

`specservice.ApplyImport(ctx, db, req *ApplyImportReq) (*ApplyImportResp, error)`, with
`ApplyImportReq{ValidateImportReq, OperatorID, PlanHash, AcceptIssues}` and
`ApplyImportResp{Plan, AfterCommit func(ctx) error}`. It plans again on the caller's
transaction, refuses as the constraints say, and otherwise writes. `PlanNode` gains `outcome`.
The three errors go in `hperrors/errors_spec.go` with their messages.

Tests: a hash from another request, a blocked plan, issues without acceptance; an unchanged plan
writes nothing and reports `unchanged`.

### Task 4: Writing a scope's settings (`import_write.go`)

- Before the first write, every setting the import creates is given an id, keyed by its path.
- For each node written, the target's own settings of its scope are loaded (`loadOwned`) and
  keyed the way export keys them; a bundle entry is the target's by entry id, then by key; a
  singleton by type.
- A reference resolves to an id: a path to the id this import gives that setting, else the
  target's; an external reference to what `findRef` finds; a raw id to itself if it exists;
  anything validate cleared to nothing. The body is read with its external blocks swapped for
  placeholders, and `RemapRefs` rewrites placeholders and paths together.
- `entity.KeepSecrets(dst, src)` copies the secrets `dst` holds nothing in from `src`, for an
  existing setting and an omit bundle.
- The fixes: `NODE_NOT_FOUND` clears the volume's node; a cleared reference or omitted secret
  makes the row `pending`.
- The row: name, kind, ref id, inheritable, default, expiry and version from `setting`; scope
  and object from the node; written with `UpsertMulti`, which rewrites its resource links.

Tests against a fake setting repository that records what it is given: a changed certificate
keeps its id; a new one gets a new id and the app's routing - once apps are written - can name
it; a reference to a cleared certificate is written empty and pending; an omit bundle leaves an
existing secret's value alone; a volume pinned to a missing node is written unpinned.

### Task 5: Global settings, projects and envs

- `global`: its settings, at global scope.
- A project created: `PrepareNewProject` with the bundle's key, name, note, the resolved owner
  or `OperatorID`, its created envs; then the bundle's project settings, each written over the
  default of the same type and key when there is one - a repo webhook keeps the secret the
  default generated when the bundle omitted it.
- A project updated: its name, note and owner as the plan says, `UpdateVer` incremented.
- An env created: its row from `NewProjectEnv`; updated: its name, colour and index.
- Env settings, then project settings, as written by Task 4.

Tests: a new project with one env and a certificate, from a bundle of another installation; an
existing project renamed; a new env in an existing project.

### Task 6: Spec, gates, merge

The import spec's §14 names this plan. `go build ./... && golangci-lint run ./... && go test ./...`.
Commit, merge into `main` locally, retest, delete the branch.
