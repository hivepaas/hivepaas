# Releasing HivePaaS

How a version of HivePaaS reaches servers: what to set up once, then the steps
of every release, a beta (`v1.0.0-beta1`) or a stable one (`v1.0.0`).

## What a release is, and who can change it

| Piece | Where | Made by | Trusted because |
|---|---|---|---|
| App and agent images | GHCR `ghcr.io/hivepaas/hivepaas`, `ghcr.io/hivepaas/hivepaas-agent` | the Release workflow, on a tag | `release.json` names them by digest |
| `release.json` / `release.signed.json` | the `release` branch | you, signed offline with `make release-sign` | the app refuses it without both signatures ([releasekeys](../hivepaas_app/pkg/releasesig/releasekeys/README.md)), and so does the installer, through its verifier image |
| `install.sh` | the GitHub Release of the tag | the Release workflow | the tag, protected |
| The release verifier image | GHCR `ghcr.io/hivepaas/release-verify` | the Release verifier workflow, rarely | `install.sh` pins it by digest |
| `get.hivepaas.com` | a redirect to that `install.sh` | you | the domain's own protection |

The workflow never signs and never publishes: the signing keys stay on the
offline machine, so a compromised CI can build images but cannot make an
installation run them.

## Versions

- Tags: `vMAJOR.MINOR.PATCH` for a stable release, `vMAJOR.MINOR.PATCH-betaN` for
  a beta: `v1.0.0-beta1`, `v1.0.0-beta2`, then `v1.0.0`. **No dot before the
  number**: the updater compares versions with
  [pkg/version](../hivepaas_app/pkg/version/version.go), which reads `-beta1`
  and not `-beta.1`. The workflow refuses any other form.
- The backend and the dashboard carry the same tag. The app image is built with
  the dashboard of its tag.
- Image tags: `1.0.0-beta1` is pushed once and never moved. `beta`, or `stable`
  and `latest`, follow the newest release of the channel. `release.json` names
  images by version and digest, never by the moving tags.
- A beta is the `beta` channel: installs default to it
  (`HIVEPAAS_CHANNEL=beta`) and run `config/config.beta.toml`. A stable release is
  the `stable` channel and `config/config.production.toml`.

## Once, before the first release

1. **GitHub environment `release`** (Settings › Environments):
   - required reviewers, the maintainers who may release. Every job that pushes
     an image waits for one;
   - **Deployment branches and tags › Selected branches and tags**: the tag
     rules `v*`, `verify-v*`, `placeholder-v*` and `kopia-v*`, and no branch.
     Only a release, verifier, placeholder or kopia tag can push an image; a run
     from a branch is refused ("not allowed to deploy to release due to
     environment protection rules").

   No secret is needed: the workflows push to GHCR with their own `GITHUB_TOKEN`.
2. **The release verifier** (see [below](#the-release-verifier)): tag `main`
   with `verify-v1` and push the tag, which runs the *Release verifier*
   workflow; pin what its summary prints in `deployment/release/install.sh`. The Release workflow refuses a tag
   while `VERIFY_IMAGE` is empty.
3. **Rulesets** (Settings › Rules), in both `hivepaas` and `hivepaas-dashboard`:
   - tags `v*` (and `verify-v*`, `placeholder-v*`, `kopia-v*` in `hivepaas`):
     only maintainers create them; no update, no deletion;
   - branch `release` (backend): no force push, no deletion, changes through a
     pull request.
4. **The `release` branch.** Installations and the installer read
   `release.signed.json` from it (`app_release_info.go`, `install.sh`). It does
   not exist until the first release creates it (step 7 below).
5. **Signing keys.** Check that the two `*.pub.pem` in
   [releasekeys](../hivepaas_app/pkg/releasesig/releasekeys/)
   are the public halves of the keys on the offline machine, not test keys: every
   binary trusts them until a later release rotates them.
6. **Dev deploys.** The dev workflow's secrets are the GitHub environment
   `development`'s, which only `dev-*` tags may use: `DOCKERHUB_USERNAME` and
   `DOCKERHUB_TOKEN` (a token that pushes to the `-dev` repositories only), and
   `DEV_SRV_SSH_KEY`, `DEV_SRV_IP`, `DEV_SRV_SSH_PORT`, `DEV_SRV_USERNAME`. The
   dev server's host key is not checked.
7. **`get.hivepaas.com`**: a redirect only, never a copy of the script.
   - HTTPS only; plain HTTP answers with a redirect to HTTPS and nothing else.
   - HSTS on `hivepaas.com`; a CAA record naming the CA you use; registrar lock
     and 2FA on the registrar and DNS accounts; DNSSEC where the provider has it.
   - The target is set per release (step 9).

## Every release

The example is `v1.0.0-beta1`; for a stable release read `stable` for `beta`.

1. **Main is green.** The `Go Source Check` workflow passes on the commit you
   release, in both repositories where it applies.

2. **The release commit (backend).** [release.json](../release.json) is the one
   place a release is declared: the binary is built with it
   ([base/version.go](../hivepaas_app/base/version.go) reads it, embedded), and
   installations fetch it, signed. In its `beta` entry:
   - `appVersion: "v1.0.0-beta1"`, exactly the tag;
   - `releaseDate`: the day you publish;
   - `appImage: "ghcr.io/hivepaas/hivepaas:1.0.0-beta1"` and
     `agentImage: "ghcr.io/hivepaas/hivepaas-agent:1.0.0-beta1"`, the tags (the
     digests do not exist yet; the binary names these two by its version anyway);
   - the other images this release runs, and `templates` if the app templates
     moved: `go run ./tools/apptemplate pin <app-templates checkout>` prints the
     pin of its commit. Keep what the entry of the other channel says.

   Then pin every image but the app's and the agent's:
   ```bash
   make release-pin        # writes tag@sha256:…; make release-pin-check only reports
   ```
   It also checks each `templates` pin, and stops on one whose index.json at
   that commit does not hash to its `indexSha256`: every server that updates
   would be left without templates.

   Name each image by a tag that says its version (`traefik:v3.7.13`, not
   `traefik:v3.7`): the digest fixes the image, the tag is what people read, and
   the updater decides "newer" from the tag. To ship a rebuild under the same tag,
   the new digest is enough: an installation pinned to the old one updates.

   Merge it to `main`. The workflow refuses a tag that is not this `appVersion`,
   and an image other than the app's and the agent's that is not pinned
   (`releasepin -check -deps`).

3. **Tag the dashboard**, on the commit to release:
   ```bash
   git -C hivepaas-dashboard tag -a v1.0.0-beta1 -m "HivePaaS v1.0.0-beta1"
   git -C hivepaas-dashboard push origin v1.0.0-beta1
   ```

4. **Tag the backend** on the release commit, the same way. The push starts the
   **Release** workflow:
   - *Check the tag*: the tag is the compiled version, the dashboard has the tag,
     `go test ./...`, shellcheck and the installer's tests;
   - *Build* (after a reviewer approves the `release` environment): app and
     agent, amd64 and arm64, with SBOM and provenance. Each image is scanned
     with Trivy before it is pushed: a HIGH or CRITICAL flaw that has a fix -
     in the Alpine packages, or in a binary's Go or modules, kopia's and
     sql-migrate's too - stops the release. Update what carries it, or, once
     reviewed, accept it in `.trivyignore` with why and an expiry date;
   - *Multi-arch images*: `1.0.0-beta1` and `beta`, and the digests in the run's
     summary and in `digests.txt`;
   - *Draft*: a draft GitHub Release (pre-release for a beta) with `install.sh`
     already reading this tag, `install.env`, `digests.txt` and `SHA256SUMS`.

5. **Pin the app and the agent.** Now that they are built, `make release-pin`
   pins `appImage` and `agentImage` to the digests `digests.txt` gives, and
   changes nothing else: the binary does not read those two digests, so the
   release commit stays the one that was built.

6. **Sign, on the offline machine**, with the reviewed tool pinned in the
   Makefile:
   ```bash
   make release-sign KEYS="/offline/2026_ed.key /offline/2026_ml.key"
   ```
   It writes `release.signed.json` and checks it again with openssl. Commit
   `release.json` and `release.signed.json` to `main` through a pull request.

7. **Check on a clean server before anyone else can install it.**
   - `make test-installer-e2e` runs a whole install in a docker:dind container
     with this checkout's `release.signed.json`, which now names the pushed
     images.
   - Then on a real VM, with the draft's installer:
     ```bash
     gh release download v1.0.0-beta1 -p install.sh   # a draft needs gh, signed in
     scp install.sh vm: && ssh vm 'sudo HIVEPAAS_RELEASE_BRANCH=main bash install.sh'
     ```
     `HIVEPAAS_RELEASE_BRANCH=main` reads the release info you just merged, before
     the `release` branch has it. Check the dashboard, a deployment, a backup.

8. **Point the `release` branch at it.** Open a pull request from `main` (or the
   commit with the signed file) into `release`; the first time, create
   `release` from that commit. From this moment installations are offered the
   release.

9. **Publish.**
   - Publish the draft GitHub Release.
   - Point `get.hivepaas.com` at the installer:
     - a beta: `https://github.com/hivepaas/hivepaas/releases/download/v1.0.0-beta1/install.sh`.
       GitHub's `releases/latest` skips pre-releases, so a beta is named by its
       tag and the redirect moves with every beta;
     - a stable release: `https://github.com/hivepaas/hivepaas/releases/latest/download/install.sh`,
       set once.
   - The landing page and the docs give `curl -fsSL https://get.hivepaas.com | sudo bash`.
   - The docs' silent install (`hivepaas-website`,
     `docs/docs/installation/silent-install.md`) downloads `install.env` from
     this release's assets: point it at the new tag. `releases/latest` skips a
     beta, so the link names its release.

## Test the update, not only the install

An install proves a release installs; only the next release proves it updates.
Release `v1.0.0-beta2` early, even with little in it, and update a `beta1` server
to it from the dashboard before `v1.0.0`: the image switch, the database
migration and `blockMajorUpgrade` run for the first time there.

## Undoing a release

- **Before step 8**: delete the draft, and the tags if the release is abandoned
  (the ruleset has to allow it to a maintainer). Nothing was offered to anyone.
- **After step 8**: point `release` back at the previous signed `release.json`
  (a pull request reverting step 8). Installations stop being offered the release;
  those that already updated stay on it until a fixed release (`beta2`, or
  `1.0.1`) supersedes it. Move the channel's image tag back with
  `docker buildx imagetools create -t ghcr.io/hivepaas/hivepaas:beta ghcr.io/hivepaas/hivepaas:<previous>`,
  and the redirect of `get.hivepaas.com`.
- Never move or delete a pushed version tag or image tag: installations and
  `release.json` refer to them.

## GHCR packages

The first push of each image creates its package under the `hivepaas`
organization, private. Once, for each of `hivepaas`, `hivepaas-agent`,
`release-verify` and `kopia` (github.com/orgs/hivepaas/packages › the package ›
Package settings):
- **Change visibility › Public**, or servers cannot pull it. If GitHub refuses,
  allow public packages in Organization settings › Packages.
- **Manage Actions access**: `hivepaas/hivepaas` with **Write**, if the package
  is not already linked to the repository (the images' `source` label links it).

Public packages cost nothing in storage or transfer. A public package version
cannot be deleted once the package has more than 5,000 downloads, and a
version's per-architecture tags (`1.0.0-beta1-amd64`, `-arm64`) are what its
multi-arch tag points to: delete neither.

## The release verifier

A server's openssl is too old for the release signatures (ML-DSA-65 and
Ed25519ctx), so the installer checks `release.signed.json` with a small image,
`tools/releaseverify`, built from `deployment/release/Dockerfile.verify`. It
carries the keys of [releasekeys](../hivepaas_app/pkg/releasesig/releasekeys/)
and runs with no network, no root and nothing writable. `install.sh` pins it:

```bash
VERIFY_IMAGE=ghcr.io/hivepaas/release-verify@sha256:…
VERIFY_KEYS=sha256:…    # the fingerprint of the keys inside it
```

It is built once and reused by every release. Build a new one (the *Release
verifier* workflow, through a tag `verify-v<N>` with the next `N`: the
`release` environment accepts no branch) only when:
- a key is added or removed in `releasekeys/`: `go test ./tools/releaseverify/`
  fails until `VERIFY_KEYS` in `install.sh` is the new fingerprint, and so does
  every release until then;
- the signature format changes;
- Go fixes a vulnerability in the crypto it uses.

Then put the two lines the workflow's summary prints into `install.sh`, through
a pull request, and release as usual.

## The placeholder

A new app is created before anything is deployed to it, and runs
`ghcr.io/hivepaas/placeholder` until then: `tools/placeholder`, one static
binary built from `deployment/release/Dockerfile.placeholder`. It answers a page
saying the app is not deployed yet on ports 80, 3000, 8000 and 8080 - a new app
has no port yet, so these are the ones apps listen on most often - and stops on
SIGTERM. It needs no command, so the image an app is given later runs its own.
`release.json` pins it as `placeholderImage`.

It is built once and reused by every release. Build a new one (the
*Placeholder* workflow, through a tag `placeholder-vX.Y.Z`; the `release`
environment must allow the tag rule `placeholder-v*`) when `tools/placeholder`
changes or Go fixes a vulnerability it is built with. The first push creates
the package private: make it public, as the others (see GHCR packages). Then
put the line the workflow's summary prints into `release.json`, or the tag and
`make release-pin`, and release as usual.

## Kopia

The app and the agent ship kopia, the backup engine, built from source by
`deployment/kopia/Dockerfile` with Go 1.27 and the modules it names raised past
kopia's release: the release's own binaries keep the Go and the modules of
their day. The image, `ghcr.io/hivepaas/kopia`, holds the binary and its
license only; the app's and the agent's Dockerfiles, release and dev, copy the
binary from it, pinned by digest:

```dockerfile
ARG KOPIA_IMAGE=ghcr.io/hivepaas/kopia:0.23.1-1@sha256:…
```

It is built once and reused by every release. Build a new one (the *Kopia*
workflow, through a tag `kopia-v<kopia version>-<build>`, as `kopia-v0.23.1-2`;
the `release` environment must allow the tag rule `kopia-v*`) when:
- the *Kopia* workflow's weekly scan of the pinned image fails: a flaw with a fix
  was found since. Raise the module in `KOPIA_MODULES`, or Go;
- kopia releases: set `KOPIA_VERSION`, and drop from `KOPIA_MODULES` what the
  release has caught up with.

The workflow tests each architecture with the backup engine's integration
tests - a filesystem repository, S3, kopia's repository server - and scans it
before pushing. The first push creates the package private: make it public, as
the others (see GHCR packages). Then put the line its summary prints into the
four Dockerfiles, through a pull request, and release as usual.

## Known gaps before 1.0.0

- **`agentImage` must be in every release from now on.** An update moves the
  agent to it before the app; a release that names none leaves the agent where
  it is, and the new app then runs against the old agent.
- **Images other than app and agent** are compiled in as release.json pins
  them, by digest; the app's and the agent's by their version's tag.
