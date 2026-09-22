# Templates that create several apps

**Status:** implemented, except the readiness wait in §4
**Depends on:** [2026-09-17-app-templates-design.md](2026-09-17-app-templates-design.md) (phase 1,
implemented), [2026-09-17-app-template-dependencies-design.md](2026-09-17-app-template-dependencies-design.md)
(implemented)

This is the expansion of one line in the first document's §12: *"Templates producing several
linked apps, which needs the reference rewriting of spec import."*

## 1. The problem

The store has 274 templates and every one of them is a single application, optionally with a
database beside it. The applications people keep asking for next are not that shape:

| Application | Services | Notes |
|---|---|---|
| Supabase | 13 | studio, api-gw, auth, rest, realtime, storage, imgproxy, meta, functions, db, supavisor, db-config, deno-cache |
| Appwrite | 29 | 23 of them are the **same image** with a different command - worker-mails, worker-builds, worker-deletes, and twenty more |

Today a template creates its own app plus at most
[`MaxDependencies = 5`](../../../hivepaas_app/service/apptemplateservice/templatemodel/dependency.go)
dependency apps, and a template named as a dependency may not have dependencies of its own
([lint_dependencies.go](../../../hivepaas_app/service/apptemplateservice/templaterepo/lint_dependencies.go)).
The ceiling is six apps. Supabase needs thirteen.

Raising the number would not help, because three things are wrong beyond the count.

**Dependencies fan out; a stack needs fan in.** Every dependency creates a *new* app. Two
components of one application cannot share one PostgreSQL: if five Appwrite components each
declared MariaDB, the request would create five MariaDBs. The thing Supabase needs is the
opposite - one `db`, and twelve services pointing at it.

**Shared secrets are copied by hand.** In Supabase's own compose file, counted:

| Value | Services that read it |
|---|---|
| `POSTGRES_PASSWORD` | 11 |
| `JWT_SECRET` | 10 |
| `ANON_KEY` | 5 |
| `SERVICE_ROLE_KEY` | 4 |

If each service is its own template, a person installing Supabase pastes secrets thirty times
and any one typo produces a stack that starts and then fails somewhere else. We have already
shipped two templates with a weaker form of this problem - `n8n-worker` and `langfuse-worker`
both ask the user to copy an encryption key from the app next door, because there is no way
for one template's generated secret to reach another template's app.

**The files would be copies of each other.** Appwrite as separate templates is 23 files
differing by one `command` line, all pinned to one image tag that has to be bumped in 23
places on every release. When a format makes us copy a file 23 times, the format is wrong,
not the catalog.

Hiding those files from the store - the `internal` flag in §7 - fixes how the store *looks*.
It fixes none of the three problems above.

## 2. Decisions

| Question | Decision | Why |
|---|---|---|
| Extend `dependencies`, or add a block? | Add `components` | They are different relationships. A dependency is another template's app: PostgreSQL is a thing on its own, reusable, and it outlives what created it. A component is one process of *this* application, meaningless alone, and it dies with it |
| Are components references to other templates, or inline? | Inline | 23 of Appwrite's 29 services are one image with one different argument. A reference means 23 near-identical files; inline means one file whose components are a list |
| Do components share the dependencies? | Yes. `dependencies` is declared once, and every component may refer to it | This is the whole point. One `db`, thirteen readers |
| May a component declare its own dependency? | No | It would be fan-out again, in a new spelling |
| How many components? | At most 30 | Appwrite is 29. The number is a guard against a template that should have been a Compose file, not a target to fill |
| Which component gets the domain? | Exactly one, marked `primary: true` | Routing is per app and an application has one front door. The other twelve are reachable inside the project by key, which is what they need |
| Is a component a real app? | Yes, an ordinary one | Console, logs, environment editing, replicas, volumes, metrics already work per app. A component that is anything else means building all of that a second time |
| How are components ordered? | `needs:` names other components; creation follows the order, and start waits on it | Supabase's `db` must have migrated before `auth` starts. Three shipped templates hand-roll a `getent` loop for the one-dependency version of this; thirteen components cannot each hand-roll it |
| What happens on delete? | The whole stack goes | Twelve apps left behind is not a deletion. It needed no work: a component is a logical child, and those already go with the app they serve - see §6, which also raises what that same cascade does to dependencies |
| How does phase 2's merge work? | Per component, keyed by component name | `AppTemplateBase` is already per app. Components make it N bases, not a new kind of base |
| What is a component's app called? | The primary takes the name the person gave; the rest are that name with the component's added, like a dependency | Somebody who types "supabase" should get an app called supabase. If every component were suffixed, the thing they named would not exist |
| Variants and components together? | Not in this phase | A variant multiplies the image table by the component table. Nothing we want needs both yet |

> The dependencies document says "at most three"; the constant has since become 5. That
> document is stale on this point, not this one.

Two rules were added while implementing, both because the alternative would have had to
guess:

- **A version of a template with components cannot override `app`.** An override patches one
  app tree, and there are several. Which one it meant is not decidable.
- **Components and variants cannot be used together**, as the table says, and this is refused
  rather than assumed.

## 3. The format

A template gains one optional top-level block. A template has `components`, or it has the
single `app` block it has today - never both.

```yaml
metadata:
  name: supabase
  title: Supabase
  # ...

parameters:
  # Declared once. Every component sees them.
  - name: jwtSecret
    title: JWT secret
    type: secret
    minLength: 32
    generate: {length: 64}
  - name: domain
    type: domain
    optional: true

dependencies:
  # Created once, shared by every component below.
  - name: db
    title: Database
    template: postgres
    version: "18"
    params: {dbName: postgres, username: supabase}

versions:
  - name: "2026.09"
    release: "2026.09.1"
    default: true
    # One image per component, in one place, so a release is one edit.
    components:
      gw:      {image: kong:3.9.1}
      auth:    {image: supabase/gotrue:v2.181.0}
      rest:    {image: postgrest/postgrest:v13.0.9}
      studio:  {image: supabase/studio:2026.09.01}

components:
  - name: gw
    title: API gateway
    primary: true                 # this one owns the domain
    needs: [auth, rest]
    app:                          # the same AppDoc shape as a single-app template
      deployment:
        source:
          activeMethod: image
          imageSource: {image: "${{ image }}"}     # this component's image
      settings:
        routing: {port: 8000, exposePublicly: true, domains: [...]}

  - name: auth
    title: Auth
    app:
      deployment:
        source:
          activeMethod: image
          imageSource: {image: "${{ image }}"}
      settings:
        envVars:
          data:
            - {k: GOTRUE_JWT_SECRET, v: "${{ params.jwtSecret }}"}
            - {k: GOTRUE_DB_HOST, v: "${{ deps.db.ref.HIVEPAAS_HOST }}"}
            - {k: GOTRUE_DB_PASSWORD, v: "${{ deps.db.ref.HIVEPAAS_PASSWORD }}"}
            - {k: API_EXTERNAL_URL, v: "http://${{ comp.gw.key }}:8000"}
```

**What is new in placeholders.** Two spellings, mirroring `deps.` exactly:

| Placeholder | Renders to | For |
|---|---|---|
| `${{ comp.<name>.key }}` | the component app's key, a literal | a URL a template assembles itself |
| `${{ comp.<name>.ref.HIVEPAAS_* }}` | `${<key>.HIVEPAAS_*}`, an environment reference | a component that declares a `kind` publishing those values |

Everything else is unchanged. `${{ params.x }}` works in every component, which is how one
generated `jwtSecret` reaches ten of them. `${{ deps.db.ref.* }}` works in every component,
which is the fan-in. `${{ image }}` resolves per component from the version table.

**One limit worth stating in the format.** Environment references - `${key.HIVEPAAS_HOST}` -
are resolved only inside environment variable *values*, not inside config file or secret
bodies. A template whose component is configured by a file must build that file at start from
environment variables; `misskey` does exactly this today and the reason is written in its
comments. With thirteen components this will be hit more often, so §11 lists resolving
references in file bodies as the follow-up it deserves.

## 4. What provisioning does

`POST /projects/:projectID/:projectEnv/apps/from-template` does not change shape. The request
still names one template and one app name; what changes is how many apps come out.

[`planApps`](../../../hivepaas_app/usecase/apptemplateuc/app_create_from_template.go) already
builds the list - dependencies first, then the app - and sets `logicalParentID` on each
dependency. Components extend that list:

1. **Render everything before anything exists.** Unchanged. Parameters resolve once,
   dependencies render against them, then each component renders against parameters,
   dependency bindings and the keys of its sibling components. Sibling keys are computable
   before creation because a component's app name is `<app>-<component>`.
2. **Order by `needs`.** A topological sort of the component graph; a cycle is refused by the
   linter, not discovered here.
3. **Create.** Dependencies, then components in that order, then record the links.
4. **Parent.** Every component app except the primary carries `LogicalParentID` = the primary
   component's app id, the same field a dependency app carries today. The primary *is* the app
   the request named - same id, same name - so it is nobody's child.

The checks that already run over the planned list - capabilities, published ports, domains,
shared mounts - are loops over `apps` and need nothing new. The port check becomes more
valuable, not less: thirteen components are thirteen chances to collide on a node port.

**Start order - not implemented.** Creation order is not start order: Swarm starts what it is given and an app
whose database is not yet resolvable fails. Three shipped templates work around this with a
`getent` loop in their `command`, written by hand. `needs:` should drive a wait that HivePaaS
injects, so the template does not carry it. This can ship as its own change before components
land - it is useful to every template with a dependency today - and if it does, components
inherit it.

## 5. What is recorded

`AppTemplateSettings` gains one field, and reuses two.

```go
// Component is the role this app plays in a template that creates several:
// "auth", "rest". Empty for an app that is the template's only one.
Component string `json:"component,omitempty"`

// Components are the apps created alongside this one because its template
// named them, in creation order. Set on the primary app only, like
// Dependencies.
Components []AppTemplateComponent `json:"components,omitempty"`
```

`CreatedForAppID` already says which app a secondary app was created for, and
`AppTemplateDependency` is the shape `AppTemplateComponent` copies (name, app id). Every
component app records its own `Base` - the render it came from - because that is what phase
2's merge needs, per component, and recording it now is what keeps phase 2 from needing a
migration.

## 6. Lifecycle

**Reading.** Nothing to build. [`excludeChildApps`](../../../hivepaas_app/usecase/appuc/list.go)
already hides apps with a `logical_parent_id` from the environment's app list and nests them
under their parent, with the reason written there: *"a person looking at the apps of an
environment is looking for the ones they made, not the six a template created underneath
them."* Thirteen components appear as one entry that opens.

**Deleting.** Nothing to build here either, and for the same reason as reading: a component
app carries `LogicalParentID`, and
[`appservice.DeleteApp`](../../../hivepaas_app/service/appservice/appserviceimpl/deletion.go)
already deletes an app's logical children recursively before the app itself. Deleting the
primary deletes the whole stack, which is what deleting an application should mean.

> **A documentation conflict found and corrected here.** The same cascade also deletes a
> template's *dependencies* - it has all along, on the same `LogicalParentID` mechanism - and
> three places said it did not: the dependencies design's decision table, a comment in
> `entity.AppTemplateSettings`, and `README.md` in the templates repository. All three were
> wrong about the code and are now corrected to match it: deleting an app deletes its
> dependencies with it, the same as it deletes components. See the dependencies design's §2
> for the corrected row and why.

**Updating.** Phase 2 renders the current revision with the app's recorded parameters and
compares against `Base.RenderedSHA256`. With components that runs per component and the
results are collected: the stack is `update-available` if any component is. Applying merges
each component against its own base and redeploys the ones that changed, in `needs` order. A
component that the new revision no longer declares is removed, and one it has added is
created - both named explicitly in the review dialog, because a merge that silently deletes an
app is not a merge anybody should approve blind.

## 7. Templates that should not be offered on their own

Independent of components, and shippable before them.

Some templates are only meaningful next to another app: `n8n-worker` and `langfuse-worker`
today, and any helper a future stack keeps out of line. They are in the catalog because there
is nowhere else to put them, and the store offers them to people who cannot use them.

`metadata.internal: true` keeps a template out of the listing and out of search. It stays
fetchable by name, so a template that names it as a dependency still works, and so does a
direct link.

**On the name.** The first suggestion was `private`. In a platform with projects, teams and
organisations, `private` will be read as "private to my organisation" - which is a real
feature somebody will want, and it should be able to have that word. `internal` says what this
one means: internal to the template repository, not offered on its own.

The change is small: the field on `Metadata`, the same field on `IndexEntry`, one condition in
[`filterTemplates`](../../../hivepaas_app/usecase/apptemplateuc/template_list.go), and a lint
rule that an internal template must be referenced by at least one template that is not - a
file nobody can reach is a file nobody will maintain.

Note what it does *not* do: with components, Supabase needs no internal templates at all,
because its thirteen services are one file. `internal` is for the leftovers, and the catalog
looking tidy is not a reason to prefer thirteen hidden files over one honest one.

## 8. Linting

New rules, all decidable from the repository alone:

- `components` and a top-level `app` are mutually exclusive; exactly one of them.
- At most 30 components; names match the dependency name pattern (`^[a-z][a-z0-9]{0,15}$`),
  which keeps `<app>-<component>` short enough to survive slugifying.
- Exactly one `primary: true`.
- `needs` names declared components only, and the graph is acyclic.
- Every version declares an image for every component, and every component's `app` block
  refers to `${{ image }}` rather than pinning one itself - so a release is one edit.
- `${{ comp.<name>.* }}` names a declared component; `ref.HIVEPAAS_*` is accepted only where
  that component's `kind` publishes the value, which is the rule `deps.` already has.
- Routing appears in the primary component only.
- No two components publish the same node port.

The renderer keeps its existing key-collision check, which now compares the whole set: thirteen
names of the form `<app>-<component>` truncated to a DNS label is where two of them meet, and
that must be refused at render, not discovered by the second provision.

## 9. Security

Nothing here grants anything new. Capabilities are checked per app over the planned list, so a
component asking for one is refused exactly as a single app would be, and the index's
`requiresCapabilities` flag is computed over all components.

**Appwrite specifically is still out of reach, and not because of this document.** Three of
its services mount `/var/run/docker.sock`, because Functions spawns runtime containers. Giving
a container in a user's project the node's Docker socket is giving it root on the node. That
is a platform question - a broker with a narrow API, or a rootless runtime - and it is
untouched by anything here. Supabase is the acceptance target for components; Appwrite needs
that other answer first.

## 10. Testing

- Render fixtures: a two-component template, one with `needs`, one with a cycle, one with two
  primaries, one with a component missing from a version's image table.
- A collision fixture: component names that truncate to the same key.
- Provisioning: order follows `needs`; every non-primary app carries `LogicalParentID`; the
  shared dependency is created once and referred to by both components.
- Deletion: components go, dependencies stay, and a failure halfway leaves nothing orphaned.
- Acceptance, on the dev cluster: Supabase from one template - Studio reachable on the domain,
  `auth` and `rest` answering through the gateway, all of them against the one `db`.

## 11. Later

| Work | Extension point |
|---|---|
| Resolving `${key.HIVEPAAS_*}` inside config file and secret bodies, not only environment values. Removes the start-up file-writing dance from `misskey`, `opencloud` and every stack that configures by file | `envvarservice` reference resolution |
| A platform-injected readiness wait from `needs`, replacing hand-written `getent` loops | Provisioning, or its own change first |
| Per-component replica counts set from the template, and scaling one component from the stack view | `components[].app.deployment` |
| Attaching a stack to dependencies that already exist, rather than creating them | The dependency binding, which records an app id either way |
| Variants together with components | The version image table |
| Appwrite, once function execution has an answer that is not the Docker socket | Out of scope here |
