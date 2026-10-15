# Image scanning, phase 2

Status: approved 2026-10-15, being built. Phase 1 is
[2026-10-11-image-scanning-design.md](2026-10-11-image-scanning-design.md):
a container per scan, Trivy's data on a chosen volume, an app's Security tab,
ignores at four scopes, the daily scan. Nothing below is built.

## What phase 2 adds

Phase 1 answers "what does this app carry" for whoever opens the app in the
dashboard. Phase 2 takes the same answer where else it is asked:

1. **A project's Security tab**: every app of the project, the worst first - for
   a project's members, who cannot open System.
2. **HivePaaS's own images**: what the installation itself runs - Traefik, the
   database, Redis, the registry, the logging backend, the app and the agent.
3. **An SBOM to download**, CycloneDX, for an image an app runs.
4. **Ignores in spec export and import**, with the project they belong to.
5. **MCP**: an assistant reads an app's vulnerabilities and scans it.
6. **The CLI**: `hivepaas app vulns` and `hivepaas app scan`, with a CI gate.

Not in phase 2: secret scanning of images (its findings - a key in a layer, no
package, no fix - fit none of phase 1's screens, and a first look is noisy), and
the deployment gate, which is phase 3.

## 1. A project's Security tab

A tab of the project, after Operations: the project's apps with what their
images carry, as System > Image Scanning > Apps shows them - the worst first,
CRITICAL and HIGH with their fixable part, last scanned, status - each a link
to the app's Security tab. A Beta badge, as the app's tab has.

- `GET /projects/:projectID/image-scan`: `data.apps`, the apps of the project
  as `GET /system/settings/image-scan/apps` answers them, and `data.enabled`,
  whether scanning is on, which a project's members cannot read in System
  settings. Not paged - a project has a screenful of apps - so the switch can
  sit in `data` beside them. Reading the project is enough; a grant on some of
  its envs or apps lists theirs alone, as the project's apps list does.
- While scanning is off, the tab says so, and still lists what was found.
- The project's apps list gets nothing new: the app header's badge already
  says it on each app.

## 2. HivePaaS's own images

The daily scan also scans the images of HivePaaS's own components that run on
the installation: the release's `appImage`, `agentImage`, `dbImage`,
`redisImage`, `traefikImage`, and, when they run, `registryImage`,
`victoriaLogsImage`, `vlagentImage` and `obiImage`. Trivy's own image is left
out: it runs only while it scans.

- Each is scanned by its reference in the release, which is pinned by digest:
  from the node running it when the scan runs there, from its registry
  otherwise - the path phase 1 already has. The platform is the manager's.
- Results are kept as an app's are, under a component key in place of an app
  ID (`hivepaas:traefik`): `app_image_scans` has no foreign key to `apps`.
- **System > Image Scanning > HivePaaS**, a tab beside Apps: each component,
  its image, CRITICAL and HIGH with their fixable part; a row opens the
  findings, as an app's Security tab shows them, with Ignore... for the global
  scope only.
- They are fixed by updating HivePaaS, not by anything the operator rebuilds:
  no Needs attention item, and the daily notification leaves them out. The tab
  says when a newer release is available, from the release info the updates
  page already reads.

## 3. An SBOM to download

**Download SBOM** on each image of the app's Security tab: the packages of the
image as CycloneDX JSON, `<app>-<short id>.cdx.json`.

- `GET /projects/:projectID/:projectEnv/apps/:appID/image-scan/sbom?imageId=…`,
  for whoever reads the app. It runs Trivy as a scan does - same node, same
  volume, same lock - with `--format cyclonedx` and every package, and answers
  the document; it is not kept. With the layers analyzed already, a few
  seconds.
- Only an image the app's last scan saw: an image ID from anywhere else is
  refused.

## 4. Ignores in spec export and import

An ignore is a decision about an app's code, which a project's spec carries
elsewhere. The exported spec gets `vulnerabilityIgnores` at every level a
setting has - global, project, env, app - under the key `specmodel` already
reserved:

```yaml
vulnerabilityIgnores:
  <the key its name derives>:   # a collection, as secrets are
    id: 01J…                    # matches it on a re-import into the same installation
    vulnId: CVE-2025-1234
    package: libxml2
    reason: the parser is never fed outside input
    setBy: alice
    setting:
      name: CVE-2025-1234 (libxml2)
      status: active
      inheritable: true
      expireAt: 2027-01-11T00:00:00Z
```

- Who wrote it goes with it, as every value specific to the installation
  does: an export is also a restore, and it must be exact. Import writes the
  row as it came: an ignore is applied when results are read, and its row is
  all there is.
- An ignore already expired is not exported: it sets nothing aside any more.
- Global ignores go with the global document, as the other global settings
  do.

## 5. MCP

Two tools, as the existing ones are made:

- `get_app_vulnerabilities` (read): the app's images and counts, and its
  fixable CRITICAL and HIGH findings - package, installed and fixed version,
  ID - the most a 64 KB answer holds; `severity`, `includeUnfixed` and
  `includeIgnored` narrow or widen it.
- `plan_scan_app_images` (a plan, as every change the server makes is):
  Scan now, for a key with Write, as the endpoint needs; `apply_plan` answers
  the task, which the existing `get_task` follows. While scanning is off there
  is no plan.

The `debug_app` prompt mentions the first.

## 6. The CLI

- `hivepaas app vulns [APP]`: the images and counts, then the findings as a
  table - fixable only unless `--all`, `--severity critical,high`,
  `--ignored` to list those set aside, `-o json`.
- `--fail-on critical|high|medium|low`: exits with a new code, **11 (vulnerable)**,
  when the app runs a fixable finding of that severity or worse that no ignore
  sets aside. The gate a CI job can run after `hivepaas deploy`, until phase 3
  puts one in the server.
- `hivepaas app scan [APP] [--wait]`: Scan now; with `--wait`, it follows the
  task and then prints what `vulns` prints.
- A server without image scanning - older than the release that brings it -
  answers 404 for these: the CLI says the server's version has no image
  scanning, from `GET /sessions/me`.

The CLI's client is generated from a released server's spec, and is released
only with one. These commands are built on hivepaas-cli's main against the
local spec: the CLI and the HivePaaS release that brings image scanning are
released together, and the release pins `SPEC_REF` to it.

## The API

All new operations; the API level is unchanged.

| Operation | What |
|---|---|
| `GET /projects/:projectID/image-scan` | the project's apps and their counts, and whether scanning is on |
| `GET /system/settings/image-scan/hivepaas` | HivePaaS's components, their images and counts |
| `GET /system/settings/image-scan/hivepaas/:component` | a component's findings and ignored ones, as an app's |
| `GET …/apps/:appID/image-scan/sbom?imageId=` | the image's CycloneDX SBOM |

## Building it, in order

1. Check, on the dind env. Done 2026-10-15, Trivy 0.75.0, a fresh volume:
   - `--format cyclonedx` with no `--scanners` lists the packages only: no
     vulnerability, and **no DB** - nothing was downloaded. `alpine:3.10`: 1 s,
     25 KB, 15 components, `apk-tools` among them. A Node app
     (`mendhak/http-https-echo`, 253 MB): 4 s cold, under 1 s again, 190 KB,
     143 components. The SBOM needs neither the DB nor its download's memory.
   - A release image the node does not have, `ghcr.io/project-zot/zot@sha256:…`
     with `--platform linux/arm64`: read from its registry in 56 s, the DB's
     download and zot's layers included; 5 HIGH, 28 MEDIUM. The release's
     references carry their digest, so a component is scanned by it: its ID for
     the node's Docker (the containerd store's ID is that digest), its
     reference for the registry.
2. Ignores in spec export and import (small, and on its own).
3. The project's Security tab: endpoint, dashboard.
4. HivePaaS's own images: the daily scan, the endpoints, the tab.
5. The SBOM: the run's format, the endpoint, the button.
6. MCP tools.
7. Website: the image scanning page and the CLI's.
8. e2e, level 2: the project tab; HivePaaS's tab lists Traefik; an SBOM
   downloads and names apk-tools; an ignore goes out with a project's spec and
   comes back with it.
9. The CLI, on main, against the local spec: `vulns`, `scan`, `--fail-on`, its
   own e2e.

## Decided (2026-10-15)

1. Phase 2's scope: the six parts above, secret scanning left out.
2. HivePaaS's own images: no Needs attention, no notification, a note when a
   newer release is out.
3. The CLI's exit code 11 for `--fail-on`, a new entry in its contract.
4. The CLI on main, against the local spec: it is released with the HivePaaS
   release that brings image scanning.
