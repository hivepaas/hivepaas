# Level 2 - deploying on one node

Apps are deployed for real: images pulled, swarm services started. Run it on
a swarm that is there to go - `yarn env:up` gives one, inside dind. Each test
has a project of its own, and deleting it takes the apps and their services.

## Done

| #    | Scenario                                                                                                                              | Test                                  |
| ---- | ------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------- |
| 2.1  | An app deployed from an image (Deployment Settings, Deploy) runs 1/1; its deployment is Done                                          | `tests/level-2/apps.spec.ts`          |
| 2.2  | A runtime variable saved, then the app re-deployed: the container prints it, as Copy logs shows                                       | `tests/level-2/apps.spec.ts`          |
| 2.3  | Replicas set to 2 run 2/2                                                                                                             | `tests/level-2/apps.spec.ts`          |
| 2.4  | Stopped, an app runs 0/0; started, 1/1 again                                                                                          | `tests/level-2/apps.spec.ts`          |
| 2.5  | Deleting an app asks for its name, then removes it                                                                                    | `tests/level-2/apps.spec.ts`          |
| 2.6  | A variable added by name, Literal, multi-line: kept after a reload; removed by its button                                             | `tests/level-2/env-vars.spec.ts`      |
| 2.7  | A volume mounted (Persistent Storage, listed by its name): what one container wrote, the next reads                                   | `tests/level-2/storage.spec.ts`       |
| 2.8  | A config file mounted by a setting mount is read by the container                                                                     | `tests/level-2/app-settings.spec.ts`  |
| 2.9  | A secret referenced from a variable as `${secrets.NAME}` reaches the container                                                        | `tests/level-2/app-settings.spec.ts`  |
| 2.10 | A scheduled job run by hand (Run Now, View Run) runs its command in the container; Done                                               | `tests/level-2/jobs.spec.ts`          |
| 2.11 | An app cloned in its environment runs with its image and variables, and deploys on its own                                            | `tests/level-2/clone.spec.ts`         |
| 2.12 | An app made from a template (IT Tools, found by searching the catalog) is deployed; its health check passes                           | `tests/level-2/templates.spec.ts`     |
| 2.13 | A data backup of an app's volume into a repository on a volume is a snapshot; restored (Replace), the app finds its data as backed up | `tests/level-2/backups.spec.ts`       |
| 2.14 | A function made from the runtime's template builds, and a Run answers 200 with what it was asked                                      | `tests/level-2/functions.spec.ts`     |
| 2.15 | A health check runs on its interval: Done when answered as asked, Failed when not                                                     | `tests/level-2/health-checks.spec.ts` |

The images: `traefik/whoami` for an app that serves, `busybox` for one that
prints what it was given. The logs are drawn on a canvas, so a test reads them
the way a person can take them: Copy logs, then the clipboard. What a test
waits for is in the log, not in a deployment's status: a deployment is done
once swarm has the new spec, before the container it starts is running.

## Next

- **Domains and routing**: an app reached through Traefik by its domain, with
  basic auth and a redirect; a function called at its domain. Needs a real
  Traefik inside dind, its ports published, and a way to reach it from the
  tests. Health checks of an app's own URL wait for it: the backend that runs
  them is on no app network in dind.

## Found while writing them (2026-10-07)

- **130 of 193 PUT and POST endpoints panic on a body of `{}`**: a request DTO
  embeds its fields' struct by pointer (`*AppEnvVarsBaseReq`), left nil when
  the body names none of them, and `ModifyRequest` or `Validate` reads it. The
  recovery middleware answers 500 "Unexpected error" - not a 400 naming the
  field. Fixed: see [level 1](level-1.md), 1.7.
- **Accessibility**: an environment variable's value field had no name, nor
  the button beside it. Fixed: each row is a group named by its key, holding
  Key, Value, Literal (described), Multi-line value and Remove variable; the
  merge view's text is named Env variables. Inherited rows are named the same.
- **Persistent Storage listed a volume by its id** (`01M4AB3Z…`), not the
  name it was picked by: docker knows a volume HivePaaS made by its id, and
  that was all the API returned. A mount now answers `sourceName`, and the
  table shows it.
- **Persistent Storage's address was `…/presistent-storage/`**: it is
  `…/persistent-storage/` now.
- **A scheduled job's command runs without a shell**, yet its placeholder was
  `echo "$CMD_ARG_GROUP_1"`, which prints that text as it is. The placeholder
  is `sh -c '…'` now, and a line under the field says so.

Seen, and left as they are:

- Re-deploying an app whose image and settings did not change starts no new
  container - swarm has nothing to change. Restart is what makes new ones.
- A clone is made from the app's service, not deployed: it lists no
  deployment until it is re-deployed.
- A scheduled job run while a deployment swaps the container fails with "No
  running task found".
- Deleting a volume right after the project that used it answers 409 "volume
  is in use" with docker's message, until the project's containers are gone.

## Found in the second round (2026-10-07), and fixed

- **The backend died when two periodic jobs notified in the same round**:
  "fatal error: concurrent map read and map write" - not a panic, the process
  ends. The round's jobs run concurrently and shared one RefObjects, the
  queue's cache, which a notification writes the settings it loads into. Two
  health checks with the default notification were enough. Each run now has
  a copy of its own.
- **Restoring a volume's snapshot was always refused** (400 "dataBackup.
  sourceVolume is invalid"): the app was read without its project and
  environment, which name its directory in a volume, so it seemed not to
  mount the volume it does.
- **A health check could not be saved as the form left it**: no return code
  and no body check, which the server refuses - any answer would pass, a 500
  too. A new check starts with 200, and the form says what is missing.
- **The template catalog's search dropped what was typed as it appeared**,
  as the tables' did.
- **Templates the server could not give showed as "No templates found... clear
  filters"**: the page now says they are not available, and why.

Seen, and left as they are:

- A template's card is a button with buttons in it (Details, Deploy, the
  version picker), which screen readers flatten.
- Every container of an app has the app's key as its hostname.
