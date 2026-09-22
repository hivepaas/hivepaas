# What a built image is called

**Status:** draft, awaiting review

## 1. The problem

Three things, all in the same twenty lines of
[calcBuildImageTags](../../../hivepaas_app/service/imagebuildservice/imagebuildserviceimpl/helper.go).

**Two apps can own the same repository.** The repository an image is pushed to is the app's
key, through [GetAutoImageName](../../../hivepaas_app/entity/app.go). An app key is unique
inside an environment, not across the installation: the development database has `backend`,
`frontend` and `db` each existing in two different projects. Pushed to one registry under one
account they land in one repository. Their images mix; "keep the last 10 builds" counts them
together, so a busy project evicts a quiet one's images; and anyone who may read that
repository reads both projects' images. The system registry makes this worse, because it is
one account for the whole installation by design.

**`repoSource.imageTags` is configuration that does nothing.** The API accepts it, the setting
stores it, a spec bundle carries it - and no build ever reads it. `ImageBuildReq` has no field
for it, and `imageBuildData.ImageTags` is nil on every path, so `calcBuildImageTags` always
takes its other branch. The dashboard has no input for it either. It is a promise the product
does not keep, and the branch it feeds would do surprising things if it ever ran: it replaces
the commit tag rather than adding to it, and a tag containing `/` is pushed verbatim, which
sends an image to Docker Hub while the operator believes they are pushing to their own
registry.

**Nothing says the repository has to exist first.** HivePaaS never creates a repository and
never checks for one: the push path calls `ImagePush` and the registry-auth screen's test
button calls `RegistryLogin`, which authenticates and nothing more. On Docker Hub, GHCR, ACR,
Quay and any self-hosted zot or distribution the first push creates the repository. On ECR the
push is refused; on Harbor the project must exist; on Google Artifact Registry the repository
must exist. The operator finds out after the build has run, in the deploy log.

## 2. Decisions

| Question | Decision | Why |
|---|---|---|
| What is an image called? | Repository `<project key>-<app key>`, tag `<env key>-<commit sha7>` | Unique across the installation, derived from values that cannot change, and one repository per app rather than one per app per environment - which is what a registry that does not create repositories by itself has to be given by hand |
| Why is the environment in the tag and not the name? | Because cleanup can still be written per environment, and fewer repositories is worth more than what it costs | Measured: zot applies `mostRecentlyPushedCount` per keepTags rule, not per repository, so `^dev-.*` and `^prod-.*` are counted separately. The cost is that access control is per repository and never per tag, so a credential that reads an app reads all of its environments (§9) |
| Can the operator choose the name? | No. `imageName` leaves the deployment settings | A name that is a function of the app cannot collide, cannot drift and needs no validation of its own. The screen shows what it will be |
| What about the tag? | `<env key>-<commit sha7>`, and always first in the list | It is what the service spec runs, so it has to be immutable. A moving tag makes a redeploy a lottery |
| Does production get a bare tag? | No, every environment is prefixed | Nothing marks an environment as production - `ProjectEnv` carries a name, a key, a status, a colour and an order - so a bare tag would mean hardcoding the literal `prod`, and an installation whose environment is called `production` or `live` would never get it. One rule, and a tag that says which environment it came from |
| Custom tags? | A per-deployment argument, **added** to the commit tag, prefixed with the environment too, never stored | A release marker is a property of one build, not of the app. The prefix is what keeps `v1.4.0` from dev and `v1.4.0` from production overwriting each other in the one repository they now share |
| `imageTags` in the settings? | Removed | It is dead configuration today, so deleting it changes no behaviour |
| May a name or tag contain `/`? | No | A tag may not contain one under the OCI grammar at all. Forbidding it in a name keeps one repository per app and closes the path that pushed to Docker Hub by accident |
| Repositories that must exist first | Not created by HivePaaS; the screen says so before the build runs | Creating them means a second credential with far wider rights, and a different API for every vendor. §7 |

## 3. The name and the tag

```
<registry address>/<registry username>/<project key>-<app key>:<env key>-<commit sha7>
```

A new method on the app replaces `GetAutoImageName`:

```go
// ImageRepoName is the repository a build is pushed to: one per app, shared by
// the app's environments, unique across the installation.
//
// It needs Project loaded. Both build paths load it - the deployment helper and
// the agent's own loader - and an app without it is a programming error rather
// than a name to guess at, so it is an error and not a shorter name.
func (app *App) ImageRepoName() (string, error)

// ImageTag is what one build is tagged with. The environment is here rather than
// in the repository name, so that one app is one repository; ImageTagPrefix is
// the same environment part, which a custom tag is given as well.
func (app *App) ImageTag(commitHash string) (string, error)
func (app *App) ImageTagPrefix() (string, error)
```

The keys are already slug form - lowercase alphanumerics with `_` or `-` - so both the joined
name and the joined tag are valid under the OCI grammar. Three rules keep them that way and
keep them short:

- **Normalize.** Collapse `__` to `_` and `--` to `-`, as `GetAutoImageName` does today, and
  drop any separator that would start or end the value or double up where segments meet.
- **Truncate the name with a suffix.** A project and an app may each be named with 100
  characters, so the joined name can pass what a registry accepts once `<address>/<username>/`
  is in front of it. Each segment is cut to 50 characters, and when any cut happened the name
  ends with `-` plus the first 6 characters of a sha256 of the app's global key. Two apps whose
  names share a long prefix therefore still get two repositories.
- **Truncate the environment in the tag.** A tag may hold 128 characters. The environment part
  is cut to 20, which leaves room for a commit or for a custom tag of up to 100.

`base.ImageNameMaxLen` (200) is replaced by the real limits: 255 for the whole repository name
minus the address and account already in front of it, and 128 for a tag.

An example, for an app `api` in project `shop`, built in two environments:

```
registry.example.com/hivepaas/shop-api:dev-9f3c1de
registry.example.com/hivepaas/shop-api:prod-9f3c1de
```

## 4. Removing the name from the settings

`DeploymentRepoSource.ImageName` and `DeploymentRepoSource.ImageTags` are removed from the
entity, from the update DTO, from the deploy DTO and from the dashboard's form state. The
setting's version is bumped and its migration drops both fields, so that a spec bundle written
before this change imports without carrying them.

Existing apps that set a name are the only ones this moves. Their next build pushes to
`<project>-<app>` instead, the old repository stays where it is - collected by the
system registry's retention, kept forever by an external one - and on a registry that does not
create repositories by itself the operator has one new repository to make. The release note
says exactly this; nothing is migrated automatically, because nothing can move images between
repositories.

## 5. Tags

A deployment may carry extra tags:

```
POST /_/projects/{projectID}/{env}/apps/{appID}/deploy
{ "repoSource": { "repoRef": "main" }, "imageTags": ["v1.4.0", "stable"] }
```

They are **tags**, not references: `v1.4.0`, never `api:v1.4.0` and never `team/api:v1.4.0`.
Each is validated against the OCI tag grammar, `[A-Za-z0-9_][A-Za-z0-9._-]{0,127}`, which
excludes `/` by construction, and is at most 100 characters so that the environment prefix
fits. At most five per deployment.

Every tag, the commit one and the extras alike, carries the environment. Without it the two
environments of an app - which now share one repository - would overwrite each other's
`v1.4.0`.

The build then produces, for `api` in project `shop`, environment `dev`:

```
registry.example.com/hivepaas/shop-api:dev-9f3c1de   <- always, and what the service runs
registry.example.com/hivepaas/shop-api:dev-v1.4.0    <- the deployment's extra tags
registry.example.com/hivepaas/shop-api:dev-stable
shop-api:dev-9f3c1de                                 <- local only, never pushed
```

Order matters: [the service spec takes `ImageTags[0]`](../../../hivepaas_app/service/appdeploymentservice/appdeploymentserviceimpl/repo_deploy_apply_svc.go),
so the commit reference has to stay first. Push keeps its present rule - only references
carrying a `/` are pushed - which now means "everything the registry prefix was added to".

Where they travel, since nothing stores them:

1. `appactiondto.DeployAppReq.ImageTags`, validated there, **not** applied to the setting.
   Today's `ApplyTo` writes the request's tags into the copy of the settings that is then
   persisted; that line goes.
2. The deployment row's own settings snapshot, which is what the build step reads.
3. `imagebuildservice.ImageBuildReq.ImageTags`, a new field.
4. `ImageBuildReq.image_tags`, a new repeated field in
   [image_build.proto](../../../hivepaas_app/interface/agent/proto/image_build.proto), for a
   build that runs on another node through the agent.

An agent older than the server ignores the new field, so a remote build would push the commit
tag and drop the extras. The build log says which tags were pushed, which is how that shows up
rather than as a silent difference.

## 6. The dashboard

The build section of an app's deployment settings loses the *Image Repository Name* input and
gains a read-only line:

```
Image           shop-api
                registry.example.com/hivepaas/shop-api                 [copy]
                Tagged dev-<commit> for this environment.
```

The second line, the full reference, appears only when a registry is chosen, because that is
the thing an operator has to create or grant access to. The copy button copies it.

Under it, when a registry is chosen and it is not one HivePaaS knows creates repositories on
first push:

> **Create this repository first.** `registry.example.com/hivepaas/shop-api` — this registry
> does not create repositories on push, so a build would fail at the push step.

Which registries that covers is decided from the address:

| Address | Note |
|---|---|
| The system registry, `docker.io`, `index.docker.io`, `ghcr.io`, `*.azurecr.io`, `quay.io` | Not shown |
| `*.dkr.ecr.*.amazonaws.com`, `*-docker.pkg.dev` | Shown |
| Anything else | **Shown** |

Unknown addresses get the note. Most of them are self-hosted registries that do create
repositories on push, so the note is unnecessary there - but being told to check something that
is already true costs a glance, and staying quiet costs a build.

## 7. Why HivePaaS does not create the repository

There is no repository-creation endpoint in the OCI distribution specification: a repository
exists once a manifest is in it. Every registry that needs one created first has its own
management API outside `/v2/` - `ecr:CreateRepository` over IAM, Harbor's
`POST /api/v2.0/projects`, Artifact Registry's `repositories.create` over a service account -
each with credentials that are not the ones used to push, and each with far wider rights than
pushing. That is three vendor integrations and a second credential type, to save one action
taken once per app. The note in §6 is the whole of the answer instead.

## 8. Testing

- **The name.** Two apps with the same key in different projects produce different
  repositories; the same app key in two environments of one project produces the same
  repository. Long names are cut to the limit and end with the hash suffix; two long names
  sharing a prefix stay distinct. A name never starts or ends with a separator and never
  contains `//` or `__`. An app without Project loaded returns an error.
- **Tags.** Every tag carries the environment, including production. The commit reference is
  first, extras follow, the local reference is last and is not pushed. A long environment key
  is cut to 20 characters and the tag stays inside 128. `api:v1`, `team/api:v1`, an empty tag,
  a 200-character tag and a sixth tag are each refused with a message naming the field.
- **Nothing is remembered.** A deploy with tags leaves the app's deployment settings
  unchanged, and the next deploy without tags pushes only the commit reference.
- **The migration.** A setting written before this change loads, loses both fields, and its
  version is bumped. A spec bundle carrying them imports without them.
- **The note.** A table test over addresses: each of the listed vendors, one unknown address,
  and no registry chosen at all.
- **End to end**, on the development cluster: an app in two projects with the same key, built
  and pushed to the system registry, gives two repositories with the right names; the same app
  built in two environments gives one repository with two tags; a deploy with `["v1.4.0"]` puts
  `dev-9f3c1de` and `dev-v1.4.0` on one image and the service runs the commit one.

## 9. Not in this

- Retention by repository pattern, which is its own task now that a project's images share a
  prefix: `hivepaas/<project>-*`. With the environment in the tag it also wants rules per tag
  pattern, which zot counts separately - measured: two rules of `mostRecentlyPushedCount: 2`
  over `^dev-.*` and `^[0-9a-f]{7}$` kept two of each, so one environment cannot evict
  another's images.
- Access control per environment. zot's `accessControl` matches repositories, never tags, so a
  credential that may read an app may read every environment of it. Per-project and per-app
  scoping still work, because the repository carries both. An installation that needs
  production images walled off from a development credential would have to put the environment
  back into the repository name, which is the trade this design makes deliberately.
- An explicit "this environment is production" flag on `ProjectEnv`. Without one there is no
  honest way to treat production differently - which is why every tag is prefixed here.
- ECR's twelve-hour token. A registry auth for ECR stores a password that expires the same day,
  so pushes and pulls stop working by the next morning. Supporting it means a registry-auth
  type that fetches its own token from an IAM key, which is a larger piece of work than any of
  this.
- A check that the repository exists. It is cheap - `GET /v2/<name>/tags/list` before the build
  - but it answers "not yet", which is also what an empty repository on an auto-creating
  registry answers, so it adds a request without replacing the note.
