# Spec Import: Availability and Permissions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Validate reports what the target cannot give an imported object as it stands - a
domain or a published port another app holds, a volume pinned to a node the target lacks,
storage that already holds data - and what the operator may not grant - capabilities, and a
mount into the storage of an app outside the import. Every finding is reported, never only the
first.

**Architecture:** The template checks' reading of a document moves to `specmodel` - the ports
a document publishes, the domains it answers at, the capabilities it grants - and template
creation calls it from there (spec §10). The service calls stay with each caller: template
creation refuses on the first finding, import asks once per domain and per port and reports
each. The permission halves stay in the usecase: `ValidateImportReq` carries two callbacks,
as it carries `AuthorizeSecrets`, and `specuc` answers them with `permissionManager`.

**Out of this plan, into the next:** secrets (`SECRET_OMITTED`, `SECRET_GENERATED`,
`CREDENTIAL_KEPT`) and the owner (`OWNER_NOT_FOUND`, `OWNER_NOT_PERMITTED`); then apply.

**Spec:** [docs/superpowers/specs/2026-09-23-config-spec-import-design.md](../specs/2026-09-23-config-spec-import-design.md)
- §3 permissions, §6 codes, §6 values tied to an installation, §10 the shared template checks.

## Global Constraints

- Before a task is done: `go build ./...`; `golangci-lint run ./...` printing `0 issues.`;
  `go test ./hivepaas_app/service/specservice/... ./hivepaas_app/usecase/specuc/...
  ./hivepaas_app/usecase/apptemplateuc/...`. The last task runs `go test ./...`.
- A check looks only at what a selected node writes: an app being created, or the block of an
  app being updated that changed; a volume setting being created, changed or pulled in.
- The objects the import itself updates are not in the way: a domain or port held by an app
  this import rewrites is left out of the check, and two objects of the import asking for the
  same domain or port are a finding on the second.
- Codes: `DOMAIN_IN_USE`, `PORT_IN_USE`, `NODE_NOT_FOUND` fixable; `STORAGE_NOT_EMPTY`,
  `STORAGE_UNCHECKED` warning; `CAPABILITY_NOT_PERMITTED`, `SHARED_MOUNT_NOT_PERMITTED` skipped
  (the app is not imported). A storage path that could not be read is unchecked, never empty.
- A nil callback allows.
- End every commit message with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

---

### Task 1: The template checks' reading moves to `specmodel`

`specmodel.PublishedPorts(doc) []PortConfig`, `specmodel.ActiveDomains(doc) ([]string, error)`
- which reads a routing body holding external references, replacing them with nothing, since a
domain's certificate is not what it answers at - and `specmodel.GrantedCapabilities(doc)
[]string`. `apptemplateuc` calls them in place of `publishedPortsOf`, `renderedDomains` and
`grantedCapabilities`, unchanged in behaviour. Tests move with them, plus one for a routing
body with an external certificate.

### Task 2: Domains and ports (`import_checks.go`)

`New` takes `domainservice.Service`. `checkAvailability(ctx)` runs after closure over the
selected nodes that write:

- the domains of each app whose routing is written: an import claim seen before is
  `DOMAIN_IN_USE` with the other app's path; otherwise `VerifyDomainsAvailable` for that domain,
  ignoring the apps whose routing the import rewrites;
- the published ports of each app whose networks are written: the same, through
  `VerifyPortsAvailable`, ignoring the services of the apps whose networks the import rewrites.

Details name the domain or `port/protocol`; the action says it is dropped. Tests with a fake
domain service and cluster service holding one of each.

### Task 3: Nodes and storage

- A volume setting written with a `nodeId` the target has no cluster node for is
  `NODE_NOT_FOUND`, found through a seam `nodeExists` (in production, an active cluster-node
  setting whose `ref_id` is the id).
- An app being created, in an env the target has, is asked `InspectAppStorage` for each managed
  mount of its own - not one with `sourceApp` - whose volume the target has: the volume's id
  is the target's entry for the mount's path, or the setting its external reference finds.
  `HasData` is `STORAGE_NOT_EMPTY`; not checked is `STORAGE_UNCHECKED`. Both warnings name
  the volume and the path.

Tests: a volume pinned to a node the target lacks, and one it has; a created app whose storage
holds data, and one whose storage could not be read.

### Task 4: Permissions (`import_checks.go`, `specuc`)

`ValidateImportReq` gains `MayGrantCapabilities func(ctx) (bool, error)` and
`MayWriteApp func(ctx, *entity.App) (bool, error)`.

- An app created with capabilities, or updated so that the capabilities it grants change, when
  `MayGrantCapabilities` says no: skipped, `CAPABILITY_NOT_PERMITTED`, naming what it asks for.
  The callback is asked once per plan.
- An app whose storage is written with a mount into the directory of an app the import does not
  write - an existing app of its env - when `MayWriteApp` says no for that app: skipped,
  `SHARED_MOUNT_NOT_PERMITTED`, naming the mount and the app.

The permission checks run before the availability checks, which skip what they skipped.
`specuc.ValidateImport` answers the first with Write on the cluster module and the second with
Write on the app, the gates template creation applies. Tests in the service for both codes and
their exemptions; in the usecase, that the callbacks ask the permission manager.

### Task 5: Spec, gates, merge

The import spec's §14 names this plan. `go build ./... && golangci-lint run ./... && go test ./...`.
Commit, merge into `main` locally, retest, delete the branch.
