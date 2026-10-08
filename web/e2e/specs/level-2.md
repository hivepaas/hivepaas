# Level 2 - deploying on one node

Apps are deployed for real: images pulled, swarm services started. Run it on
a swarm that is there to go - `yarn env:up` gives one, inside dind. Each test
has a project of its own, and deleting it takes the apps and their services.

## Done

| #    | Scenario                                                                                                                                                        | Test                                    |
| ---- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------- |
| 2.1  | An app deployed from an image (Deployment Settings, Deploy) runs 1/1; its deployment is Done                                                                    | `tests/level-2/apps.spec.ts`            |
| 2.2  | A runtime variable saved, then the app re-deployed: the container prints it, as Copy logs shows                                                                 | `tests/level-2/apps.spec.ts`            |
| 2.3  | Replicas set to 2 run 2/2                                                                                                                                       | `tests/level-2/apps.spec.ts`            |
| 2.4  | Stopped, an app runs 0/0; started, 1/1 again                                                                                                                    | `tests/level-2/apps.spec.ts`            |
| 2.5  | Deleting an app asks for its name, then removes it                                                                                                              | `tests/level-2/apps.spec.ts`            |
| 2.6  | A variable added by name, Literal, multi-line: kept after a reload; removed by its button                                                                       | `tests/level-2/env-vars.spec.ts`        |
| 2.7  | A volume mounted (Persistent Storage, listed by its name): what one container wrote, the next reads                                                             | `tests/level-2/storage.spec.ts`         |
| 2.8  | A config file mounted by a setting mount is read by the container                                                                                               | `tests/level-2/app-settings.spec.ts`    |
| 2.9  | A secret referenced from a variable as `${secrets.NAME}` reaches the container                                                                                  | `tests/level-2/app-settings.spec.ts`    |
| 2.10 | A scheduled job run by hand (Run Now, View Run) runs its command in the container; Done                                                                         | `tests/level-2/jobs.spec.ts`            |
| 2.11 | An app cloned in its environment runs with its image and variables, and deploys on its own                                                                      | `tests/level-2/clone.spec.ts`           |
| 2.12 | An app made from a template (IT Tools, found by searching the catalog) is deployed; its health check passes                                                     | `tests/level-2/templates.spec.ts`       |
| 2.13 | A data backup of an app's volume into a repository on a volume is a snapshot; restored (Replace), the app finds its data as backed up                           | `tests/level-2/backups.spec.ts`         |
| 2.14 | A function made from the runtime's template builds, and a Run answers 200 with what it was asked                                                                | `tests/level-2/functions.spec.ts`       |
| 2.15 | A health check runs on its interval: Done when answered as asked, Failed when not                                                                               | `tests/level-2/health-checks.spec.ts`   |
| 2.16 | An app exposed at a domain from Routing Settings answers there over HTTPS; a new domain forces HTTPS, and HTTP is sent there                                    | `tests/level-2/routing.spec.ts`         |
| 2.17 | Without Force HTTPS, the app answers plain HTTP as well                                                                                                         | `tests/level-2/routing.spec.ts`         |
| 2.18 | Basic auth on a domain: 401 without the credentials, 200 with them                                                                                              | `tests/level-2/routing.spec.ts`         |
| 2.19 | A domain set to Redirect To sends its visitors to the main one, path and query kept                                                                             | `tests/level-2/routing.spec.ts`         |
| 2.20 | Headers added on the way in and on the way out; a path prefix stripped                                                                                          | `tests/level-2/routing.spec.ts`         |
| 2.21 | A client outside the Allowed IPs gets 403; a request over the rate limit gets 429                                                                               | `tests/level-2/routing.spec.ts`         |
| 2.22 | A function created at a domain answers there, sent to HTTPS                                                                                                     | `tests/level-2/functions.spec.ts`       |
| 2.23 | A health check of an app at its own domain passes                                                                                                               | `tests/level-2/health-checks.spec.ts`   |
| 2.24 | A project made from a Compose file: its services run as apps, a variable its .env gives reaches the container, a published port answers at the domain suggested | `tests/level-2/compose.spec.ts`         |
| 2.25 | A project exported from its Operations page, deleted, and imported on the installation's runs as it was: its app, its variable, its domain                      | `tests/level-2/spec.spec.ts`            |
| 2.26 | Exporting secrets is refused while the server does not return them, and the page says why                                                                       | `tests/level-2/spec.spec.ts`            |
| 2.27 | A deployment of an image that does not exist fails, and its card says the image was not found; the app runs on in the container it ran                          | `tests/level-2/deploy-failures.spec.ts` |
| 2.28 | A container that keeps exiting is on Home's Needs attention - keeps restarting, with its exit code - and View logs shows what it printed                        | `tests/level-2/deploy-failures.spec.ts` |
| 2.29 | A project's Slack platform tried with Test Send Msg, then saved, and a notification target through it                                                           | `tests/level-2/notifications.spec.ts`   |
| 2.30 | A health check telling a target: nothing while it is healthy, once when it fails, once when it passes again                                                     | `tests/level-2/notifications.spec.ts`   |
| 2.31 | Repeat while failing tells a check that still fails again                                                                                                       | `tests/level-2/notifications.spec.ts`   |
| 2.32 | A file uploaded from the terminal goes over a websocket, in pieces, and lands whole; to a directory not there, it is refused, and the dialog says so            | `tests/level-2/container-files.spec.ts` |

The images: `traefik/whoami` for an app that serves - it answers with the
request it got, as the proxy passed it on - and `busybox` for one that prints
what it was given. The logs are drawn on a canvas, so a test reads them
the way a person can take them: Copy logs, then the clipboard. What a test
waits for is in the log, not in a deployment's status: a deployment is done
once swarm has the new spec, before the container it starts is running.

An app's domain is `<e2e name>.localhost`, reached the way a person reaches a
domain: the browser keeps the address and connects where the env's proxy
answers (see the README). Chrome's own upgrade of `http://` to HTTPS is off, so
what a test sees of plain HTTP is the proxy's answer.

What stands in for Slack is an app of the test's own,
`mendhak/http-https-echo`, which logs each request it gets, body and all, on a
line of its own: the backend posts to it at its domain, as it would to Slack's
webhook. The app a health check calls is `whoami`, whose `/health` answers with
the code last POSTed to it: it fails, and passes again, with no restart.

## Next

- **Routing, the rest**: a TCP domain (a database through the proxy, by SNI),
  settings of one path, compression, websockets, a certificate of one's own.
- **Secrets in a spec**: exported encrypted and imported with the passphrase.
  The server refuses to return secrets unless its Security settings allow it,
  and turning that on for one test would unmask them for the tests running
  beside it.

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

## Found reviewing the periodic jobs (2026-10-07), and fixed

- **A periodic job on an interval longer than five minutes never ran, or ran
  every five minutes**: every reload of the workers' cache - one every five
  minutes, one for every setting that changes - gave every job a new slot in
  the schedule, as if new; a job due later than the next reload was pushed on
  before it was ever due. The schedule is kept across reloads now; a job is
  slotted when it is new, and again when its interval changes.
- **A health check that never got an answer held every other periodic job**:
  a tick waited for all its runs, and a check with no timeout - the HTTP
  client has none - waited for the run's ceiling of fifteen minutes. The runs
  go in the background now, and a check with no timeout gives an attempt 30
  seconds, or its interval when that is shorter.
- **A health check notified on every run** - every 10 to 30 seconds for a
  healthy app, when a default notification was configured - and the state it
  kept was its first result ever, so every run after the first change read
  as a change. It notifies when the result changes now: failing, and healthy
  again after failing; a check healthy from its first run says nothing. "Min
  Send Interval" is "Repeat while failing": set, a check still failing is
  told again that often.
- Smaller: reload events queued while a tick was busy were one reload each;
  a job's next run was counted from when it was seen, a tick late each time;
  the sources a notification would be sent through were loaded before it was
  known whether it would be.

Seen, and left as they are:

- A template's card is a button with buttons in it (Details, Deploy, the
  version picker), which screen readers flatten.
- Every container of an app has the app's key as its hostname.

## Found with the proxy (2026-10-07), and fixed

- **Without Force HTTPS, HTTP answered 404**: a domain had a router on HTTPS
  alone, and one on HTTP only to send it to HTTPS - so turning Force HTTPS off
  left HTTP with no router at all, while the function dialog said to turn it
  off for plain HTTP. A domain not forced to HTTPS now has a router of its own
  on HTTP, with the same rule, service and middlewares; so has each of its
  paths.
- **The function dialog's domain field had no name** but its placeholder: it
  is named Domain.

Seen, and left as they are:

- A health check trusts only certificates someone signed: an app on a domain
  with no certificate - a `.localhost` one, a local one - is checked over HTTP.
- In Routing Settings, every number field is named "Number input", other
  fields only by their placeholder ("192.168.1.0/24", "/foo", "1s"), and each
  section's "Enabled" and "Remove section" carry no section's name.

## Found with specs and Compose (2026-10-07), and fixed

- **A project deleted and made again at once could not create its first app**:
  "network ... not found", a 500 - restored from its spec, a Compose file
  applied again. Its env's network lingers until the old project's containers
  have stopped - seconds, or the grace an app asks for - and while it does, it
  can neither be used nor made again under its name; HivePaaS found it by name
  and took it as the new project's. A network the project neither records nor
  labeled is now waited out (30 seconds at most) and made anew, labeled with
  the project's id.

Seen, and left as they are:

- Every app exposed at a domain is attached to the proxy's network, and the
  import flags that attachment as one its project cannot use - fixable,
  dropped - though routing attaches it again: noise in each plan.
- On the Compose page a published port's domain field is named by the domain
  it suggests, and the project's and the env's "Own settings" checkboxes share
  their name.

## Found with notifications and failing deployments (2026-10-07), and fixed

- **A health check telling one target of both results could not be saved**: a
  500, "ON CONFLICT DO UPDATE command cannot affect row a second time". A
  setting's links to what it uses are written in one statement, and the target
  was in it twice. The same held for anything that tells one target of success
  and of failure: a scheduled job, an app's deployment settings. A link wanted
  twice is written once now.
- **The env wrote to real people**: the backend's seed holds a mail account
  (Gmail's SMTP), a Slack, a Discord and a Telegram, all defaults, and every
  project starts with a target that sends through the defaults - so each
  deployment and each daily cleanup in the env sent mail and messages through
  them. `env/up.sh` removes the seed's channels and targets; a test makes its
  own, pointed at an app of the env.
- **A failed deployment said why in codes**: its reason, its log and the
  comment on a preview's pull request were the error's chain of codes - and
  for an error of HivePaaS's own, that was all of it: "ERR_PRECONDITION_FAILED
  ERR_MISSING" where "Registry auth to pull image is missing" was meant. They
  are the error's code and what it means now, as a task's output is. An error
  no code names - a task's too - is its message alone, no longer under an
  empty line.
- **Home named an app's env by its key**: "project / dev", and linked the
  app's screens at `/dev/`, where they are `/development/` - the server took
  both, one screen at two addresses. It is the env's name now.

Seen, and left as they are:

- In a full run, an app made after other projects were deleted had its first
  tasks fail before the container ran: "invalid pool request: Pool overlaps
  with other one on this address space" - the env's new network was given a
  subnet the node had not let go of yet. Docker's own; Home told it as the
  app's latest failure, which is what it was.
- An app's placeholder container starts twice when the app is made: the
  service is created, then the env's variables are applied by an update, which
  replaces its task.
