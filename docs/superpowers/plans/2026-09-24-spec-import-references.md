# Spec Import: References and Closure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Validate follows every reference a selected object makes - a setting to a setting, an
app to a sibling app, a mount to a volume or to another app's directory - and says what import
will do about each: resolve it to what the bundle brings, resolve it to what the target already
has, pull in what the target lacks, or clear it with `REF_NOT_SELECTED` or `REF_NOT_FOUND`. The
records of the projects and envs a selected object needs are created with it.

**Architecture:** References are read from a bundle body with the type's own
`GetRefObjectIDs()`: the body is parsed as its setting type, after each `external` block is swapped
for a token, so only the strings the type calls references are followed. Each reference is then a
path into the bundle, an external reference, or a raw id. The planner resolves them after it has
planned every node, pulling the settings the target lacks into their settings node, and repeating
for what those settings reference, until nothing new is pulled. The target is looked up through a
seam, as export's loaders are, so the fixture can answer.

**Out of this plan, into the next:** the availability and permission issues (domains, ports,
storage, nodes, capabilities, shared mounts), secrets and owner notes, and apply - which will remap
the references this plan resolves.

**Spec:** [docs/superpowers/specs/2026-09-23-config-spec-import-design.md](../specs/2026-09-23-config-spec-import-design.md)
- §1 the ancestor rule, §4 matching of collection settings, §5 closure, §6 `REF_NOT_SELECTED` and
  `REF_NOT_FOUND`.

## Global Constraints

- Before a task is done: `go build ./...`; `golangci-lint run ./...` printing `0 issues.`;
  `go test ./hivepaas_app/service/specservice/...` whole. The last task runs `go test ./...`.
- A reference is one of:
  - a **path** into the bundle: `global/<block>[/<key>]`, `projects/<p>/<block>[/<key>]`,
    `projects/<p>/envs/<e>/<block>[/<key>]`, `projects/<p>/envs/<e>/apps/<a>/<block>[/<key>]`,
    where `<block>` is a known block name and `<key>` may hold `/`;
  - an **external** block `{external: {type, name, kind, id}}` standing where the id was;
  - a **raw id** export could not resolve, because the setting was gone.
- Closure follows references, never containment; pulls in only what the target lacks; never
  reaches outside the route's scope; never pulls in an app. Individual settings are not nodes: a
  settings node pulled in by closure is `selectedBy: dependency`, and its `changes` name only the
  settings pulled in.
- "The operator deselects" means an `exclude` pattern covers the node. A node an `include` merely
  leaves out is not selected, and closure may pull from it.
- A reference that cannot be pulled in resolves against the target as an external one would - by
  id, then by type, name and kind, as the referencing scope sees settings - and otherwise is
  `REF_NOT_SELECTED` (fixable) when the bundle holds it, with `availableIn` naming its node, and
  `REF_NOT_FOUND` (fixable) when it does not.
- Issues name what holds the reference (`setting` or `mount`) and what it named, never a value.
- End every commit message with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

---

### Task 1: What the target lacks, and `keep` creating it (`import_diff.go`)

`hasEntry(current map[string]any, key string, body any) bool`: a collection entry is on the target
when `current` holds its key, or an entry with the same `id` (a rename since the export, §4). A
singleton is on the target when the block is. `settingsChanges` records, per node, which of the
changes are settings the target lacks (`planner.missing[path]`).

`applyExisting` with `existing: keep`: an app or a project or env record stays `keep`; a settings
node whose scope exists creates what it lacks - `update`, with `changes` narrowed to those
settings - and is `keep` when it lacks nothing. That is the spec's "only what is missing is
created", at the grain of a setting.

Tests: a new global certificate under `keep` makes `global` an `update` naming only it, while a
changed one is left alone; a renamed entry with the same id is not missing.

### Task 2: Reading references from a bundle (`import_refs.go`)

- `decodeImportedSetting(block, typ, key, body) (*entity.Setting, entity.SettingData, error)`,
  taken out of `addImportedSetting`: the row from `setting`, the data brought to this version with
  `Migrate`, parsed. `addImportedSetting` keeps the rest.
- `settingRefs(typ, key, body) ([]bundleRef, error)`: copies the body with every `external` block
  swapped for a token, decodes it, and reads `GetRefObjectIDs()`: setting ids become path, external
  or raw-id references; app ids become app references. A body from a newer HivePaaS has none -
  `SETTING_VERSION_NEWER` already blocks it - and a body that does not parse is
  `ERR_SPEC_BUNDLE_INVALID` with the setting's path.
- `mountRefs(storage)`: each managed mount's volume - `source` as a path, or `external` - and its
  `sourceApp`.
- `parseRefPath(path) (refTarget, bool)`: the node the path's setting lives in, its settings map's
  location, the block and the key.

Tests: a routing body with an external certificate and a path one gives both, and the body is not
changed; a raw id; a newer version gives nothing; a path with `/` in its key; paths that are not
references (`global`, an unknown block).

### Task 3: Resolving references, and closure (`import_refs.go`, `import_plan.go`, `service.go`)

- The planner keeps the bundle as it was before `restrictToScope` (`full`), each node's settings
  map and app document by path, and the nodes by path.
- A seam `findRef func(ctx, db, scope, ref *specmodel.ExternalRef) (*entity.Setting, error)`, nil
  when nothing matches; in production `settingRepo.GetByID` then `GetByName` with the kind, both
  through the repository's scope visibility. The scope is the referencing node's nearest ancestor
  that exists on the target: its env, else its project, else global.
- `resolveRefs(ctx)`, after `applyExisting`: a worklist of the selected nodes that write -
  `create` or `update` - and then of each setting pulled in. For every reference:
  - a path whose node is selected and imported: resolved;
  - a path the target has at the same place (`hasEntry` on the target's export): resolved;
  - a path whose node is not selected, not excluded, not skipped and inside the route: pulled in -
    the node becomes `selectedBy: dependency`, `create` when its scope is new and `update`
    otherwise, its `changes` the settings pulled, its issues those of the settings pulled - and
    the setting's own references join the worklist;
  - an app's path (`…/apps/<a>/…`), or anything else that cannot be pulled: the bundle's entry
    becomes an external reference, found with `findRef` or `REF_NOT_SELECTED`;
  - an external reference: `findRef`, else `REF_NOT_FOUND`;
  - a raw id: `loadByIDs`, else `REF_NOT_FOUND`;
  - an app id: an imported app of the bundle with that id, else the target's app with it, else
    `REF_NOT_SELECTED` if the bundle holds it and `REF_NOT_FOUND` if not;
  - a `sourceApp`: an imported app of that key in the same env of the bundle, else the target's,
    else the same two codes.

Tests against the export fixture: an unchanged import still raises nothing; selecting only the app
pulls in a certificate the target lacks, and nothing else of `global`; excluding `global` makes it
`REF_NOT_SELECTED` with `availableIn: global`; an external volume nobody has and a raw certificate
id are `REF_NOT_FOUND`; a `sourceApp` nobody has is `REF_NOT_FOUND`; a project bundle imported at
its env's route resolves the project's volume on the target.

### Task 4: Ancestor records (`import_plan.go`)

`selectAncestors()`, after closure: the project and env nodes above a selected node that writes,
when they are `create` and not selected, become `selectedBy: dependency`. Their settings nodes are
not pulled in (§1: the record, not its settings).

Test: a new project in the bundle, only one of its apps selected - the project and the env are
created by dependency, the project's settings are not selected.

### Task 5: Spec, gates, merge

The import spec's §14 names this plan. `go build ./... && golangci-lint run ./... && go test ./...`.
A throwaway check against the local stack plans an export of it with no issues, and is deleted.
Commit, merge into `main` locally, retest, delete the branch.
