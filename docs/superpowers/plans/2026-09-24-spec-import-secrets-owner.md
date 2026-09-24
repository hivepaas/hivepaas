# Spec Import: Secrets and Owner Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Validate says what becomes of the secrets a bundle does or does not carry, and who
owns each project it writes: an omitted secret left empty, a credential HivePaaS generates, a
credential the target keeps, an owner found by id or email or not at all, and an owner change
the operator may not make.

**Architecture:** Secret findings come from the bundle's `secretsMode` and the settings import
creates, read as their types with a new `entity.CountEmptySecrets`. The owner is resolved in the
planner through `UserRepo`; whether the operator may change an existing project's owner is a
third callback on `ValidateImportReq`, answered in `specuc` with the gate project update applies.

**Out of this plan, into the next:** apply.

**Spec:** [docs/superpowers/specs/2026-09-23-config-spec-import-design.md](../specs/2026-09-23-config-spec-import-design.md)
- §4 the project owner, §6 codes and notes, §7 secrets.

## Global Constraints

- Before a task is done: `go build ./...`; `golangci-lint run ./...` printing `0 issues.`;
  `go test ./hivepaas_app/entity/... ./hivepaas_app/service/specservice/...
  ./hivepaas_app/usecase/specuc/...`. The last task runs `go test ./...`.
- A setting import **creates** is one of a node being created, one the target lacks in a node
  being updated, or one closure pulled in. Only those are looked at for secrets: an existing
  setting keeps the target's secret, and an omitted one never overwrites a value with nothing.
- `omit` bundles: a created setting holding an empty secret is `SECRET_OMITTED` (fixable),
  except the values HivePaaS owns - the `app-kind` credentials and a project's `repo-webhook`
  secret - which are generated, note `SECRET_GENERATED`. An empty secret the source never had
  reads the same; that false positive is accepted (decided 2026-09-24).
- `encrypted` and `plaintext` bundles: an app matched by key whose target has an `app-kind`
  credential keeps it, note `CREDENTIAL_KEPT`; the diff compares that app's kind without its
  credential, since the credential is not what import writes.
- The owner: the active user with the bundle owner's id, else the active user with its email,
  else the operator with note `OWNER_NOT_FOUND`. A new project gets it. An existing project
  whose owner no user matches keeps its owner, with the same note - handing a project to
  whoever happens to import it is a change nobody asked for. An existing project whose resolved
  owner differs changes owner only if `MayChangeOwner` allows; otherwise `OWNER_NOT_PERMITTED`
  (fixable) and `owner` leaves its changes.
- Notes carry no severity and need no acceptance; they are not part of `planHash`.
- End every commit message with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

---

### Task 1: Counting empty secrets (`entity/setting_secrets_omit.go`)

`CountEmptySecrets(data SettingData) int` walks the same shape `OmitSecrets` does and counts the
`EncryptedField`s that hold nothing. Tests: a secret with a value and one without; a nil nested
struct holds none; a map and a slice of structs with secrets.

### Task 2: Secrets (`import_secrets.go`)

- The planner learns which settings of a node import creates: `created(node) []string`, from
  the action, `missing` and `pulled`.
- `checkSecrets()` after closure, for `omit`: `SECRET_OMITTED` on the node, detail
  `{setting, secrets: n}`; or note `SECRET_GENERATED` for `app-kind` and `repo-webhook`.
- For `encrypted` and `plaintext`: `CREDENTIAL_KEPT` on an app matched by key that writes, when
  the target's kind holds a credential; `appKindCredentials` names the credential paths, and a
  test holds them to the `EncryptedField`s of `entity.AppKindSettings`. The kind block is
  compared with the target's credential put in the bundle's place.

Tests: an omit bundle creating a secret and an app with a database credential; an existing
secret raises nothing; an encrypted bundle whose app is matched by key and by id.

### Task 3: The owner (`import_owner.go`, `specuc`)

`New` takes `repository.UserRepo`. `ValidateImportReq` gains `OperatorID` and
`MayChangeOwner func(ctx, *entity.Project) (bool, error)`. `resolveOwner(ctx, doc)` returns the
user and how it was found. `planProject` compares the resolved owner with the target's in place
of the bundle's owner id, which differs between installations for the same person. The usecase
sets `OperatorID` and answers `MayChangeOwner` with project update's gate: an admin, the current
owner, or Write on the Project module.

Tests: found by id, by email when the id is another installation's, not found for a new project
and for an existing one, a disabled user skipped; an owner change allowed and refused; the
usecase's gate for an admin, the owner and a member.

### Task 4: Spec, gates, merge

The import spec's §14 names this plan and records the owner decision in §4.
`go build ./... && golangci-lint run ./... && go test ./...`. Commit, merge into `main` locally,
retest, delete the branch.
