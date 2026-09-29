# Releasing HivePaaS

How a version of HivePaaS reaches servers: what to set up once, then the steps
of every release, a beta (`v1.0.0-beta1`) or a stable one (`v1.0.0`).

## What a release is, and who can change it

| Piece | Where | Made by | Trusted because |
|---|---|---|---|
| App and agent images | Docker Hub `hivepaas/hivepaas`, `hivepaas/hivepaas-agent` | the Release workflow, on a tag | `release.json` names them by digest |
| `release.json` / `release.signed.json` | the `release` branch | you, signed offline with `make release-sign` | the app refuses it without both signatures ([releasekeys](../hivepaas_app/service/hpappservice/hpappserviceimpl/releasekeys/README.md)) |
| `install.sh` | the GitHub Release of the tag | the Release workflow | the tag, protected; see [Known gaps](#known-gaps-before-100) |
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

1. **Docker Hub.** Create `hivepaas/hivepaas` and `hivepaas/hivepaas-agent`, and
   an access token that can push to those two repositories only.
2. **GitHub environment `release`** (Settings › Environments): required reviewers
   (the maintainers who may release), and the secrets `DOCKERHUB_USERNAME` and
   `DOCKERHUB_TOKEN` with the token above. The build jobs wait for a reviewer
   before they can read them.
3. **Rulesets** (Settings › Rules), in both `hivepaas` and `hivepaas-dashboard`:
   - tags `v*`: only maintainers create them; no update, no deletion;
   - branch `release` (backend): no force push, no deletion, changes through a
     pull request.
4. **The `release` branch.** Installations and the installer read
   `release.signed.json` from it (`app_release_info.go`, `install.sh`). It does
   not exist until the first release creates it (step 7 below).
5. **Signing keys.** Check that the two `*.pub.pem` in
   [releasekeys](../hivepaas_app/service/hpappservice/hpappserviceimpl/releasekeys/)
   are the public halves of the keys on the offline machine, not test keys: every
   binary trusts them until a later release rotates them.
6. **Dev deploys.** Add the secret `DEV_SRV_KNOWN_HOSTS`: the dev server's host
   keys, from `ssh-keyscan -p <port> <host>` run where you already trust that
   server. The dev workflow refuses any other host key.
7. **`get.hivepaas.com`**: a redirect only, never a copy of the script.
   - HTTPS only; plain HTTP answers with a redirect to HTTPS and nothing else.
   - HSTS on `hivepaas.com`; a CAA record naming the CA you use; registrar lock
     and 2FA on the registrar and DNS accounts; DNSSEC where the provider has it.
   - The target is set per release (step 9).

## Every release

The example is `v1.0.0-beta1`; for a stable release read `StableVersion` for
`BetaVersion` and `stable` for `beta`.

1. **Main is green.** The `Go Source Check` workflow passes on the commit you
   release, in both repositories where it applies.

2. **The release commit (backend).** In
   [base/version.go](../hivepaas_app/base/version.go), `BetaVersion`:
   - `AppVersion: "v1.0.0-beta1"`, exactly the tag;
   - `ReleaseDate`: the day you publish;
   - `AppImage: "hivepaas/hivepaas:1.0.0-beta1"` and
     `AgentImage: "hivepaas/hivepaas-agent:1.0.0-beta1"` (the tags; the digests
     do not exist yet), and the other images this release runs.

   Merge it to `main`. The workflow refuses a tag that is not this `AppVersion`.

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
     agent, amd64 and arm64, with SBOM and provenance;
   - *Multi-arch images*: `1.0.0-beta1` and `beta`, and the digests in the run's
     summary and in `digests.txt`;
   - *Draft*: a draft GitHub Release (pre-release for a beta) with `install.sh`
     already reading this tag, `install.env`, `digests.txt` and `SHA256SUMS`.

5. **release.json.** Update the `beta` entry: `appVersion`, `releaseDate`,
   `appImage` and `agentImage` as `digests.txt` gives them
   (`hivepaas/hivepaas:1.0.0-beta1@sha256:…`), the other images (by digest too,
   preferably), and `templates` if the app templates moved. Keep what the entry of
   the other channel says.

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
  `docker buildx imagetools create -t hivepaas/hivepaas:beta hivepaas/hivepaas:<previous>`,
  and the redirect of `get.hivepaas.com`.
- Never move or delete a pushed version tag or image tag: installations and
  `release.json` refer to them.

## Known gaps before 1.0.0

- **The installer does not verify `release.signed.json`'s signatures.** It checks
  the sha256 inside the same file, which catches corruption, not a forged file.
  `scripts/release-sign.sh` already verifies these signatures with openssl, so
  the installer can do the same with the public keys written into it.
- **`agentImage` must be in every release from now on.** An update moves the
  agent to it before the app; a release that names none leaves the agent where
  it is, and the new app then runs against the old agent.
- **Images other than app and agent** are pinned by tag in the compiled
  `version.go`; pin them by digest in `release.json` at least.
