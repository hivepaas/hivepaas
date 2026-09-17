# App template image override and registry version scan

**Status:** approved, not implemented
**Depends on:** [2026-09-17-app-templates-design.md](2026-09-17-app-templates-design.md) (phase 1, implemented)

## 1. The problem

A template pins exact images - `postgres:18.6-alpine3.24` - and `templatemodel.IsPinnedImage`
refuses moving tags, because a tag that moves changes what runs with no template revision to
show for it: on a redeploy, or when swarm reschedules a task onto another node.

That pin rots at the speed of whoever maintains the template. With a few hundred templates,
HivePaaS will not keep every patch level current, and a template stuck at 18.4 while upstream
ships 18.6 is not only an annoyance - it is a user running an image with published CVEs and no
obvious way out.

Two facts shape the design.

**Users are not blocked today.** An app created from a template carries an ordinary
`app-deployment` setting, and the deployment settings screen accepts any image. Anyone can point
their database at `postgres:18.6` right now. What is missing is the signal that they are behind,
and any help choosing what to move to.

**The work splits in two.** Patch and minor bumps inside a declared major line (18.4 → 18.6) are
frequent, and drop-in by construction: same data directory, same environment variables, same
health check. New major lines (18 → 19) are rare and genuinely need a human - PostgreSQL 18 moved
its data directory, which is why the template carries an `override` for 17 and 16. Only the first
kind is a capacity problem, and only the first kind can be automated. §7 automates it; the rest of
this spec is the escape hatch for when automation lags, and the discovery that makes it usable.

## 2. Decisions

| Question | Decision | Why |
|---|---|---|
| What may the user override to? | Any tag in the **same repository** as the template's pinned image, except `latest` | Everything else in the template - `POSTGRES_*` variables, `pg_isready`, the mount at `/var/lib/postgresql`, `settings.kind.engine` - is only correct for that software. A free-form image is a footgun wearing a feature's clothes |
| A moving tag such as `postgres:18`? | Allowed, labelled as moving | The user is choosing the risk for one app, knowingly. The template itself still may not use one: that would impose the risk on everybody |
| A different major line? | Allowed, with a different warning | The template's overrides were written for the lines it declares. For another line they may be wrong in ways nobody has tested |
| A digest (`@sha256:…`)? | Allowed | It is the strongest pin there is |
| How does the user find a newer tag? | A button that scans the registry on demand | No background polling: rate limits are only reached when somebody asks, an air-gapped install gets an error on one button instead of a broken store, and nothing calls out on its own |
| Which registries in phase 1? | Those that answer anonymously - Docker Hub and public V2 registries | Private registries reuse the existing `registry-auth` setting later (§10) |
| Who decides the repository to scan? | The server, from template + version + variant | A client that could name the address would turn the endpoint into a probe for the cluster's internal network |
| Is the override remembered? | Yes - in the `app-template` setting and in the audit entry | Phase 2's merge must see it as the user's change, and "I accept the risk" is an act worth recording |

## 3. Overriding at creation

`POST /projects/:projectID/:projectEnv/apps/from-template` gains one optional field:

```json
{"name": "main-db", "template": "postgres", "version": "18", "imageOverride": "postgres:18.7-alpine3.24"}
```

Empty means the template's own image, which stays the default in the dialog - the override lives
behind "Advanced", not in the main flow.

The value is checked by a pure function beside `IsPinnedImage`, because the linter and the
dashboard both need the same answer:

```go
// templatemodel
type ImageOverrideClass string   // "same-line" | "other-major" | "moving"

func ClassifyImageOverride(templateImage, override string) (ImageOverrideClass, error)
```

Rules, in order:

1. Both references are parsed with `imageref.Parse`.
2. The repository - registry host included - must be equal. Anything else is
   `ERR_APP_TEMPLATE_IMAGE_NOT_ALLOWED`, naming the repository the template uses.
3. An empty tag, or the tag `latest`, is refused by the same error: both mean "whatever is newest",
   which is a moving tag with no version to record.
4. A reference carrying a digest, or a tag `IsPinnedImage` accepts, whose leading number equals the
   chosen version's major → `same-line`.
5. Same, but a different leading number → `other-major`.
6. A tag `IsPinnedImage` refuses (`18`, `18-alpine`) → `moving`.

The class reaches the dashboard in the error-free case too, so the confirmation can say the right
thing. The wording matters more than the mechanism:

- `same-line` - "HivePaaS has not tested this exact release. The template's configuration applies."
- `other-major` - "This is a different major line than the template describes. Its configuration -
  data directory, environment, health check - may not apply, and the app may fail to start or
  refuse to read existing data."
- `moving` - "This tag moves. What runs will change without notice on a redeploy, or when a task
  is rescheduled."

## 4. Scanning the registry

```
GET /projects/:projectID/app-templates/:templateName/image-tags?version=18&variant=alpine
```

```json
{
  "data": {
    "repository": "docker.io/library/postgres",
    "currentTag": "18.6-alpine3.24",
    "truncated": false,
    "tags": [
      {"tag": "18.7-alpine3.24", "class": "same-line", "newer": true},
      {"tag": "19.0-alpine3.24", "class": "other-major", "newer": true}
    ]
  }
}
```

**Fetching.** The standard V2 flow: `GET /v2/<repo>/tags/list`, and on a `401` read
`WWW-Authenticate`, fetch a token from the realm it names, retry. Docker Hub is that same flow with
an anonymous token. Nothing in HivePaaS speaks to a registry over HTTP today - the daemon does the
pulling - so this is a new, small client under `services/registry`.

**Filtering is the actual work.** `postgres` publishes thousands of tags: every version, every
base, every alias. From the pinned tag the server keeps only what a user would recognise as a
newer build of the same thing:

- tags that do not parse as a version are dropped, and so are `latest` and bare aliases;
- the suffix family of the current tag is required - `alpine*` against `alpine*`, `trixie` against
  `trixie` - because `18.7-trixie` is not a newer build of an app running alpine;
- the rest are sorted newest first, classified as in §3, and marked `newer` by
  `imageref.IsUpgrade`.

This is inference, not law. `pkg/imageref` says so in its own doc comment - tags are not semantic
versions and it answers only where it is confident. So the endpoint offers candidates and the
person chooses; it never picks for them.

**Limits.** At most 1000 tags read (paginating with `n=` and `Link`), at most 50 returned, with
`truncated` saying when more existed. Results are cached per repository for ten minutes in memory,
the way `hpappserviceimpl` caches release info. A `429` surfaces as
`ERR_REGISTRY_RATE_LIMITED`; anything else unreachable as `ERR_REGISTRY_UNAVAILABLE`. A failed
scan never blocks creating an app from the template's own pinned image.

## 5. Where the override is recorded

`entity.AppTemplateSettings` gains one field:

```go
// ImageOverride is the image the user chose instead of the template's, empty when
// the template's own image is in use. The class is not stored: it is derived, and
// a later HivePaaS may classify the same pair differently.
ImageOverride string `json:"imageOverride,omitempty"`
```

The audit entry for the creation gains `imageOverride` beside the existing `template`, `version`,
`release` and `revision`. Parameter values stay out, as before.

## 6. What phase 2 must do with it

Phase 2 merges three things per field: the template's previous render (the stored base), the
template's new render, and the app as it stands. The override lives only in the app's deployment
setting, so the merge already sees it as a user change and keeps it - which is the wanted
behaviour, and this spec adds no mechanism for it.

What phase 2 must add is honesty in the update dialog. When the template's pin moves from 18.6 to
18.8 and the app is running an override of 18.7, the diff must show both and let the user take the
template's image back, rather than listing "no change to image" because the user's value won.

`ImageOverride` exists so that screen can tell "the user chose this" from "the template said this",
without guessing from the strings.

## 7. Keeping the templates current

The escape hatch is not a substitute for maintenance. In `app-templates`:

```
apptemplate bump [-dry-run] <dir>
```

For every declared version line it asks the registry for tags in the same family with a higher
patch level, rewrites `release` and `images`, and prints what it changed. A scheduled workflow
runs it, and opens a pull request when the diff is non-empty; the existing CI - `lint` and
`index -check` - gates the merge. New major lines stay a human's job: they need a `versions` entry
and, often, an `override`.

This is the part that makes a few hundred templates survivable. `bump` shares the registry client
from §4, so it is mostly wiring.

## 8. Security

The trust chain is untouched. Release info is signed offline, pins the templates commit and the
hash of `index.json`, and every template file is verified against the hash the index names before
it is rendered. An override changes one field of one app, inside the repository the verified
template already names, and is recorded in the audit log.

Two properties the implementation must keep:

- the scan endpoint derives its address from a verified template, never from the request body, so
  it cannot be aimed at an internal host;
- `latest` and an empty tag are refused, so a recorded override always names something a human can
  read back and reason about, even when it is a moving tag.

## 9. Testing

Unit, with no network: a table for `ClassifyImageOverride` (same repository, other repository,
other registry host, `latest`, bare name, digest, same and other major, moving tags); tag
filtering against a recorded tag list served by a fake registry, including the `401` → token →
retry flow, pagination, a `429`, and a repository whose tags are all aliases.

Live, on the second backend: create a postgres app with an override one patch ahead, confirm the
service runs that image, the binding reports `imageOverride`, and the audit entry carries it;
then scan the tags of that template and confirm the pinned tag is not offered as newer than
itself.

## 10. Later

- **Private registries.** Reuse the `registry-auth` setting (`Address`, `Username`, encrypted
  `Password`) for the scan; the field already exists on app deployments.
- **Passive notice.** Once scanning is proven, the store could show "the template pins 18.6,
  upstream has 18.8" without a click - but only where a registry is reachable, and never as a
  background job on an air-gapped installation.
- **Per-template policy.** A template whose configuration is known to break across patch levels
  could declare `imageOverride: forbidden`, and the dialog would say why.
