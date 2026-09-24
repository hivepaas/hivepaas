# Spec Import: Apply, Part 2 - Apps and the Endpoint Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Apply writes apps - a new one provisioned the way template creation provisions one, an
existing one's settings written and, once committed, its running service brought to the bundle -
queues the deployments the options ask for, and is reachable at the three apply routes, recorded
as `spec-import`.

**Architecture:** An app's settings go through the same writer as a scope's, at app scope, with
the app-level fixes: domains and ports taken elsewhere dropped, the target's credential kept for
an app matched by key, owned secrets generated, swarm object ids kept only where they are this
installation's. Its deployment blocks are prepared - volumes named by id, a cleared mount or port
dropped - and built by `BuildApp` in import mode: into a new service's spec inside
`ProvisionApps`, into an existing service's spec after the commit. Phase 2 then updates the
swarm files, the routing and, once per env touched, the environment of every app in it.

**Out of this plan:** the dashboard screens.

**Spec:** [docs/superpowers/specs/2026-09-23-config-spec-import-design.md](../specs/2026-09-23-config-spec-import-design.md)
- §8 phase 1 step 5-6, phase 2, phase 3; §3 routes and apply body; §6 app fixes; §7 secrets.

## Global Constraints

- Before a task is done: `go build ./...`; `golangci-lint run ./...` printing `0 issues.`;
  `go test` of the packages touched. The last task runs `go test ./...` and `make gen-swag`.
- A new app keeps the bundle's key. Its id is chosen before anything is written, so settings of
  the import can name it.
- Phase 1 runs in the caller's transaction: scopes and existing apps' settings persisted, then
  new apps provisioned (their services and swarm files are created in docker, and `Cleanup`
  removes them when the transaction rolls back), deployments created, app-scope scheduled jobs
  scheduled. Nothing in docker changes for an existing app in phase 1.
- Phase 2 runs after the commit, for each updated app that has a service: its deployment blocks
  when any changed (one `ServiceUpdate`), its secrets and config files, its routing; then the
  environment of every app in each env touched. A failure marks that app `failed` with its error
  and the rest go on.
- Phase 3 schedules what phase 1 created: first deployments, deployments of changed sources,
  certificate tasks.
- End every commit message with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

---

### Task 1: A provisioned app can keep its key

`ProvisionAppReq.Key`: the app's key, derived from its name when empty. Test in
`appprovisionserviceimpl`.

### Task 2: App settings through the writer

- The writer's scope handling learns apps: rows from `loadOwned([app], appID)`, nothing for a new
  app; the setting written is `settings.<name>` or the deployment source (`app-deployment`).
- New app ids are chosen with the setting ids, and a reference to an app of the bundle maps to
  its id here.
- Fixes: `DOMAIN_IN_USE` drops the domain; an existing secret's or config file's swarm ids are
  the target's, a new one's are cleared; in `encrypted` and `plaintext` bundles an app matched by
  key keeps the target's app-kind credential; in `omit` bundles a created `app-kind` or
  `repo-webhook` setting gets the secrets it holds nothing in generated.

Tests: a routing setting with a taken domain; a created app-kind with generated credentials; an
app matched by key keeps its credential; a created repo webhook gets a secret.

### Task 3: The deployment an app is built from

`preparedDeployment(node)`: a copy of the bundle's deployment without its source, each managed
mount's volume named by the id it has here - the setting this import writes, the target's, or
what its external reference finds - a mount validate cleared dropped, a port validate dropped
dropped. Tests for each.

### Task 4: Phase 1 for apps

- Existing apps: their settings persisted with the scopes'; scheduled jobs of theirs that changed
  scheduled; with `deployChangedSource`, a deployment created for each whose source changed.
- New apps, after the scopes are persisted: `ProvisionApps` with the chosen id and key, the
  settings from Task 2, `BuildApp` in import mode on the prepared deployment, and a first
  deployment when the node deploys.
- `ApplyImportResp` gains `Cleanup`, `Tasks` and `Deployments`; `AfterCommit` takes the database
  to run phase 2 on.

Tests with fakes: a new app is provisioned with its key, settings and spec; an existing app's
changed settings are persisted with their ids; deployments are created as the options say.

### Task 5: Phase 2

For each updated app with a service, loaded with its project, env and settings: the deployment
blocks built onto its inspected spec and updated once; each secret and config file written
updated or created in docker; routing applied. Then, once per env touched - an app updated in it,
or its env's or project's variables or secrets changed - the environment of every app in it
built and applied. Outcomes: `applied`, or `failed` with the error for the app that failed.
`PlanNode` gains `error`.

Tests with fakes: a changed resource block updates the service once; a failing update marks that
app failed and leaves the next one applied.

### Task 6: The endpoint

- `specuc.ApplyImport`: the validate gates, the transaction, `Cleanup` when it does not commit,
  the `spec-import` audit record in it (digest, scope, selection, options, summary), phase 2,
  then the tasks scheduled. `base.AuditLogTypeSpecImport`.
- `specdto.ApplyImportReq` (validate's body with `planHash` and `acceptIssues`) and its response
  (the plan with outcomes, the deployments queued).
- `POST /spec/import/apply`, `/projects/:projectID/spec/import/apply`,
  `/projects/:projectID/:projectEnv/spec/import/apply`, with the 25 MB limit; `make gen-swag`.

Tests: the usecase records the import and cleans up when the service fails.

### Task 7: Spec, gates, merge

The import spec's §14 names this plan. `go build ./... && golangci-lint run ./... && go test ./...`.
Commit, merge into `main` locally, retest, delete the branch.
