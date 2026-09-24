# Docker API Access - Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** The dashboard shows and edits what plans 3, 4 and 7 built:
- the Docker API a template gives, in the store, the template's page and the deploy dialog;
- an app's Docker API screen, host mode included;
- the privileged-apps switch as it now is, with the apps in host mode.

**Architecture:**
- **Templates.** The Docker API follows the capabilities pattern exactly:
  - a util describes the block in words;
  - the template page and the deploy dialog each show a notice;
  - the dialog is blocked without Write on the Cluster module;
  - the card is marked from a new index flag.
- **The app screen** is a new configuration route, `docker-api`. It follows the resources slice file for file: contracts, validator, API, context, hook, query key, query, command, route, form, schema. It sits under "Runtime & Scale".
- **Security.** The switch's description is rewritten, and a list of the apps in host mode links to each app's screen.

**Tech Stack:** React 19, TanStack Query, React Hook Form with Zod, shadcn UI (dashboard); Go (the index flag).

**Spec:** `docs/superpowers/specs/2026-09-24-docker-api-access-design.md` §9, §10, §15.

## Global Constraints

- **Wire format** (backend `appsettingsdto`, `apptemplatedto` and `hpappsettingsdto`):
  - `GET/PUT .../apps/{appID}/docker-api-settings` carry `enabled`, `mode`, `images`, `sharedDirs`, `networks`, `allow` and `limits {containers, memory (string such as "2gb"), cpus}`;
  - `GET` adds `defaultLimits`, `hostMode {available, blockedBy: "" | "switch" | "admin"}` and `updateVer`;
  - the template detail and its dependencies have `dockerApi {images, sharedDirs, networks, allow, containers, memory, cpus}`;
  - the template list has `requiresDockerApi`;
  - the security settings have `privilegedApps [{appId, appName, projectId, projectName, projectEnvKey, projectEnvName}]`.
- **Gates shown before the server refuses:**
  - Cluster Write, from `useConditionalModule({id: MODULE_IDS.Cluster})`, for proxy access;
  - `hostMode.available` for host mode.
- **Dashboard gates:** `npm run lint:ci`, `npm run build`.
- **Backend gates:** `go build ./...`, `golangci-lint run ./...`, `go test ./...`, `make gen-swag`.
- **Git:**
  - branch `feat/docker-api` in each repo;
  - every commit ends with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`;
  - merge locally and delete the branch. Do not push.

---

### Task 1: The store flag, and the env key of a privileged app (backend, app-templates)

**Files:**
- `hivepaas_app/service/apptemplateservice/templatemodel/index.go`: `RequiresDockerAPI bool` (`json:"requiresDockerApi,omitempty"`)
- `hivepaas_app/service/apptemplateservice/templaterepo/index.go`: `requiresCapabilities` becomes `requires(repo, tmpl, asks)`, used for both flags (+ `repo_test.go`)
- `hivepaas_app/usecase/apptemplateuc/apptemplatedto/template_list.go`: `RequiresDockerAPI bool` (`json:"requiresDockerApi"`)
- `hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto/security_settings_get.go`: `ProjectEnvKey` in place of `ProjectEnvID`, because dashboard routes take the env key (+ test)
- `app-templates/index.json`: regenerated with `go run ./tools/apptemplate index ../app-templates`

- [ ] **Test first:**
  - a template with the block, and one depending on it, are marked;
  - the demo is not;
  - the privileged app carries its env key.
- [ ] **Implement, regenerate** swagger and the index, run the gates.
- [ ] **Commit in each repo, and merge.**

### Task 2: Templates in the dashboard

**Files (dashboard, `src/application/modules/app-templates`):**
- `api/app-templates.api.ts`:
  - `AppTemplateDockerApi`;
  - `dockerApi` on the detail and on each dependency;
  - `requiresDockerApi` on the summary.
- `utils/docker-api.ts` (+ export):
  - `grantedDockerApiByTemplate(template)`, which returns `{app, access}[]`;
  - `describeDockerApi(access)`, which returns one short phrase per part: the images, the shared directories, the env network, the groups, and the limits with their defaults.
- `components/app-templates-details-view.com.tsx`: `DockerApiSection` beside `CapabilitiesSection`.
- `dialogs/deploy-template/form/deploy-template.form.com.tsx`:
  - `DockerApiNotice`;
  - the submit is blocked when capabilities **or** the Docker API are asked without Cluster Write.
- `components/app-template-card.com.tsx`: a "Starts containers" badge when `requiresDockerApi`.

- [ ] **Implement:** reuse the capabilities markup, with a sky tone rather than orange, because nothing reaches the host. The shared sentence: the app starts containers on its node through HivePaaS's Docker API proxy, only these images, on its own network, within these limits, and never gets the Docker socket.
- [ ] **Check:** `npm run lint:ci`.
- [ ] **Commit:** `feat(templates): show the Docker API a template gives, and block deploying it without Cluster Write`.

### Task 3: The app's Docker API screen

**Files (dashboard, `src/application/modules/projects`):**
- `domain/apps/docker-api-settings/app-docker-api-settings.entity.ts` (+ index, and `domain/apps/index.ts`).
- `api/services/project-apps-services/docker-api-settings/`: `*.api.contracts.ts`, `*.api.validator.ts`, `*.api.ts`, `index.ts`.
- Registration:
  - `api/services/project-apps-services/index.ts`;
  - `api/api-context/projects.api.context.ts` (`dockerApiSettings`);
  - `api/hooks/project-apps/use-app-docker-api-settings.api.ts` (+ index);
  - `data/constants/projects.query-keys.ts` (`projects.apps.docker-api-settings.$.find-one`);
  - `data/queries/project-apps/app-docker-api-settings.queries.ts` (+ index);
  - `data/commands/project-apps/app-docker-api-settings.commands.ts` (+ index), and the query key added to `invalidateSingleAppConfigurationQueries`.
- `src/application/shared/constants/route.constants.ts`: `configuration.dockerApi`, at `projects/:id/:env/apps/:appId/docker-api`.
- `projects.router.tsx`, `projects.module.ts` and `routes/index.ts`: `AppConfigDockerApiRoute`.
- `layouts/single-app/configuration-layout/single-app-configuration-layout.com.tsx`: a "Docker API" item under "Runtime & Scale", with the `Container` icon.
- `layouts/single-app/header/single-app-header.com.tsx`: the active path prefix.
- `routes/single-project/single-app/configuration/docker-api/`: `route`, `form`, `schemas`, `types`, `building-blocks`, `index.ts`.

**The screen:**
- **Enable.** A switch for access.
- **Mode.** A radio between:
  - "Through the HivePaaS proxy" (recommended);
  - "The node's Docker socket", locked with the reason from `hostMode.blockedBy` when unavailable.
- **Proxy fields,** shown only in proxy mode:
  - images and shared directories, as `SingleValueList`s;
  - the env network, as a checkbox;
  - the groups, as checkboxes with a line each;
  - the limits, with the defaults as placeholders.
- **Warnings:**
  - the proxy gate: widening needs Write on the Cluster module;
  - host mode: a red warning that the app controls its node, and on a manager the whole cluster.
- **Submit** sends the whole state with `updateVer`. The server's `meta.warning` is shown by the global interceptor.

- [ ] **Implement, following the resources slice.**
- [ ] **Check:** `npm run lint:ci`, `npm run build`.
- [ ] **Commit:** `feat(apps): a Docker API screen, with host mode behind the switch and an administrator`.

### Task 4: The security settings

**Files (dashboard, `src/application/modules/system-settings`):**
- The security settings validator and domain: `privilegedApps`.
- `routes/hivepaas/security/building-blocks/allow-privileged-apps-section.com.tsx`:
  - the new description: templates and the proxy do not need the switch; host mode and raw host mounts in an import do, together with an administrator; turning it off takes nothing away;
  - `PrivilegedAppsList`, which links each app to its Docker API screen.

- [ ] **Implement.**
- [ ] **Check:** `npm run lint:ci`, `npm run build`.
- [ ] **Commit, and merge** the dashboard branch.
