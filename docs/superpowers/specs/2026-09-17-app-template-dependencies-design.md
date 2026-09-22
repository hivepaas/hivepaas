# App template dependencies

**Status:** approved, implemented
**Depends on:** [2026-09-17-app-templates-design.md](2026-09-17-app-templates-design.md) (phase 1, implemented)

## 1. The problem

Every template shipped so far is one app: a database, a cache, an object store. A web
application is not. WordPress needs MySQL, Outline needs PostgreSQL and Redis, Plausible
needs PostgreSQL and ClickHouse. Today the store can offer those images, and the person
installing one is left to create the database themselves, copy its credentials into the
app's environment by hand, and discover the wiring by reading someone else's docker-compose
file. That is the part a template exists to remove.

The templates for those databases already exist and are tested. What is missing is a way for
one template to say "I also need that one", and for provisioning to act on it.

## 2. Decisions

| Question | Decision | Why |
|---|---|---|
| How does a template ask for a database? | It names another template in the same repository | The database templates are written, linted and tested. Re-declaring MySQL inside every web app template would copy an image pin, a health check and an environment into a dozen files that then rot apart |
| Inline apps, or a reference? | A reference | An inline second app is a second place to fix a bug |
| How deep may it go? | One level. A template named as a dependency may not have dependencies of its own | A tree is a package manager. One level covers every case we have, and it is the only depth a creation dialog can show honestly |
| How many? | At most three | One database is the common case, a database and a cache the next. Three is room to spare, not an invitation |
| May it reuse an app that already exists? | Not in this phase - a dependency is always created | One flow instead of two. The binding records an app id either way, so "attach the Postgres I already run" is an addition later, not a rewrite |
| What happens to the database when the app is deleted? | It is deleted with it | A dependency is a logical child of the app it was created for - `App.LogicalParentID` - and `appservice.DeleteApp` deletes an app's logical children recursively before the app itself. Keeping the data past the app that used it means detaching it first, not deleting around it |
| Where do the dependency's parameter values come from? | The template sets what it can; the dialog asks for what is left | Every database template has a required `dataVolume` with no default, so a dialog that asked nothing would be lying about what it needs |
| When is a dependency resolved? | At creation only | Phase 2 updates an app from its template. Moving a running application from MySQL to PostgreSQL is a data migration, not a merge, and nothing here will pretend otherwise |

> **Corrected 2026-09-22.** This row originally said deleting the app leaves the dependency
> where it is, on the theory that deleting an app is aimed at the app and data that outlives
> the click should be recoverable. The implementation never did that: `DeleteApp`'s cascade
> over `logical_parent_id` deletes dependencies along with everything else created for an
> app, and its own comment says so ("the dependencies a template created alongside it"). The
> `CreatedForAppID` comment in `entity.AppTemplateSettings` and this repository's `README.md`
> both said the old, wrong thing as well and are corrected with this. See
> [2026-09-22-app-template-components-design.md](2026-09-22-app-template-components-design.md)
> §6, which is where the mismatch was found.

## 3. The format

A template gains one optional top-level block, beside `parameters` and `versions`:

```yaml
dependencies:
  - name: db                      # the role, and the suffix of the created app's name
    title: Database               # what the dialog calls it
    template: mysql               # a template in this repository
    version: "8.4"                # optional; the dependency template's default when absent
    variant: alpine               # optional
    params:                       # values this template fixes for the dependency
      dbName: wordpress
      username: wordpress
```

`params` may use the main template's own placeholders - `dbName: "${{ params.siteName }}"` -
because the dependency is rendered after the main template's parameters are resolved. A
dependency parameter that is neither set here nor given a default by its own template, and is
not optional, is asked of the person creating the app: in practice the data volume, and
nothing else.

**Referring to the dependency.** The app that needs the credentials refers to them through
two placeholders:

```yaml
settings:
  envVars:
    data:
      - {k: WORDPRESS_DB_HOST, v: "${{ deps.db.ref.HIVEPAAS_HOST }}"}
      - {k: WORDPRESS_DB_USER, v: "${{ deps.db.ref.HIVEPAAS_USER }}"}
      - {k: WORDPRESS_DB_PASSWORD, v: "${{ deps.db.ref.HIVEPAAS_PASSWORD }}"}
      - {k: WORDPRESS_DB_NAME, v: "${{ deps.db.ref.HIVEPAAS_DATABASE_NAME }}"}
```

`${{ deps.db.ref.X }}` renders to `${blog_db.X}` - an ordinary HivePaaS environment
reference, resolved at deploy time against the dependency app. The password is never
rendered into the app's stored settings, exactly as it is not today when somebody wires two
apps together by hand. `${{ deps.db.key }}` renders to the app key alone, for a connection
string a template would rather assemble itself.

Only `HIVEPAAS_*` names are accepted after `ref.`, and only names the dependency's kind
actually publishes: `ref.HIVEPAAS_BUCKET` against a database is refused by the linter, which
can see both templates.

## 4. What provisioning does

`POST /projects/:projectID/:projectEnv/apps/from-template` gains one optional field:

```json
{"name": "blog", "template": "wordpress",
 "params": {"siteName": "blog"},
 "dependencyParams": {"db": {"dataVolume": "01J…"}}}
```

In order:

1. The template's parameters are resolved, so the dependencies' `params` can refer to them.
2. Each dependency, in declaration order, is loaded from the same index - so from the same
   revision - and rendered from what the template fixes plus `dependencyParams`, bound to
   the key its app will have: the slug of `<app name>-<dependency name>`, so `blog-db`
   becomes `blog_db`, which is what the references resolve against.
3. The template is rendered with `deps` bound to those keys and to what each dependency's
   kind shares.
4. Only then, in one transaction, the apps are provisioned - dependencies first - each exactly
   as a direct creation would be: settings, environment, routing, first deployment. Their
   ids are chosen before the first one exists, so each binding can name the others.

Every app of the request is rendered before any is created, so a parameter it refuses or a
reference that cannot resolve fails the request with nothing to clean up.

A request whose apps would share a key - slugifying truncates long names - is refused while
rendering, before anything is created. A name already taken by an app in the project is
refused when that app is provisioned, and the apps provisioned before it in the same request
are removed with the transaction. Retrying with a different name is the fix; a generated
suffix would hide a collision that means something.

**When it fails.** Everything runs inside the existing transaction, and the deferred cleanup
that removes a half-created app's swarm service today walks a list, newest first. A failure
after the database is running removes it. If that removal fails, the error says so by name -
"the database blog-db was created and could not be removed" - because an orphan nobody is
told about is worse than an orphan.

**Readiness is not ordered.** The application may start before its database is accepting
connections, and swarm will restart it until it succeeds. Waiting for a dependency's health
check before starting the app is a scheduler feature; a template whose image does not retry
says so in its description.

## 5. What is recorded

`entity.AppTemplateSettings` gains, on the main app:

```go
// Dependencies are the apps created alongside this one, by role.
Dependencies []AppTemplateDependency `json:"dependencies,omitempty"`

type AppTemplateDependency struct {
    Name     string `json:"name"`     // the role: db, cache
    AppID    string `json:"appId"`
    Template string `json:"template"`
}
```

and on each created dependency:

```go
// CreatedForAppID is the app this one was created to serve, empty when it was
// created on its own.
CreatedForAppID string `json:"createdForAppId,omitempty"`
```

Both directions are stored because both screens need them: the app's page shows what it
depends on, and the delete dialog of either one shows the other. The audit entry for the
creation carries the dependency app ids beside `template`, `version` and `revision`.

## 6. The index and the store

`IndexEntry` gains a `dependencies` list of `{name, title, template}`, so the store can say
"also creates: MySQL" on the card and in the dialog without reading every template file -
the same reason the license moved into the index. The linter fills it from the template.

`index.json` is read by every installation of a channel whatever its version, so it is
decoded tolerantly: an older HivePaaS ignores `dependencies` rather than refusing the whole
index. Template files stay strict, and `requires.versionCode` keeps an older HivePaaS from
provisioning a template it cannot read.

The creation dialog shows one section per dependency: what will be created, which template
and version it comes from, and the parameters it still needs. The confirmation names every
app that is about to exist, because this is the first template that creates more than one.

## 7. Linting

Beside the existing rules, `apptemplate lint` refuses:

- a dependency naming a template that is not in this repository;
- a dependency naming a template that itself declares dependencies;
- a dependency naming the template it is declared in;
- more than three dependencies, or two with the same `name`;
- a `version` or `variant` the dependency's template does not declare;
- a `params` key the dependency's template does not declare;
- `deps.<name>` for a dependency that is not declared, and `ref.X` where X is not a
  `HIVEPAAS_*` variable the dependency's kind publishes.

The last two are what stop a template from being published with wiring that cannot work.

## 8. Security

A dependency is resolved inside the revision the release info pins, and every file is
verified against the hash the index names, exactly as the main template is. A template may
not name a template from another source, so nothing new is fetched and nothing crosses a
trust boundary.

The dependency's secrets are generated by its own template as they are today, stored
encrypted on that app, and reach the main app only as an environment reference - which means
that revoking access is deleting the reference, and reading the database password still
requires permission on the database app.

The cap of three and the refusal of nested dependencies bound the work one request can
create: at most four apps, known before anything starts.

A dependency's `params` may take the declaring template's parameters, but never one of its
secrets: a secret belongs to one app, and copying it into another's parameters would store it
twice.

## 9. Testing

Unit, without a cluster: the linter's refusals in §7, each as its own case; the render of
`deps.<name>.ref.X` and `deps.<name>.key`; parameter resolution where a dependency parameter
refers to a main parameter; the split of dependency parameters into "fixed by the template"
and "asked of the user".

Live, on the second backend: create WordPress with its MySQL, confirm two apps exist, that
the web app's environment resolves to the database's credentials, and that WordPress answers
its install page; delete the web app and confirm the database is still there with its data;
then create a template whose main app fails to render and confirm no database is left behind.

## 10. Later

- **Attaching an existing app.** A parameter type that picks an app in the environment, with
  a check that its kind and engine match what the template asks for.
- **Ordered startup.** Waiting for a dependency's health check before the first deployment of
  the app that needs it.
- **Dependencies in phase 2.** Showing, in the update diff, that the template now asks for a
  different database - and doing nothing about it automatically.
