# Level 2 - deploying on one node

Apps are deployed for real: images pulled, swarm services started. Run it on
a swarm that is there to go - `yarn env:up` gives one, inside dind. Each test
has a project of its own, and deleting it takes the apps and their services.

## Done

| #    | Scenario                                                                                                                                                                                                                                      | Test                                        |
| ---- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------- |
| 2.1  | An app deployed from an image (Deployment Settings, Deploy) runs 1/1; its deployment is Done                                                                                                                                                  | `tests/level-2/apps.spec.ts`                |
| 2.2  | A runtime variable saved, then the app re-deployed: the container prints it, as Copy logs shows                                                                                                                                               | `tests/level-2/apps.spec.ts`                |
| 2.3  | Replicas set to 2 run 2/2                                                                                                                                                                                                                     | `tests/level-2/apps.spec.ts`                |
| 2.4  | Stopped, an app runs 0/0; started, 1/1 again                                                                                                                                                                                                  | `tests/level-2/apps.spec.ts`                |
| 2.5  | Deleting an app asks for its name, then removes it                                                                                                                                                                                            | `tests/level-2/apps.spec.ts`                |
| 2.6  | A variable added by name, Literal, multi-line: kept after a reload; removed by its button                                                                                                                                                     | `tests/level-2/env-vars.spec.ts`            |
| 2.7  | A volume mounted (Persistent Storage, listed by its name): what one container wrote, the next reads                                                                                                                                           | `tests/level-2/storage.spec.ts`             |
| 2.8  | A config file mounted by a setting mount is read by the container                                                                                                                                                                             | `tests/level-2/app-settings.spec.ts`        |
| 2.9  | A secret referenced from a variable as `${secrets.NAME}` reaches the container                                                                                                                                                                | `tests/level-2/app-settings.spec.ts`        |
| 2.10 | A scheduled job run by hand (Run Now, View Run) runs its command in the container; Done                                                                                                                                                       | `tests/level-2/jobs.spec.ts`                |
| 2.11 | An app cloned in its environment runs with its image and variables, and deploys on its own                                                                                                                                                    | `tests/level-2/clone.spec.ts`               |
| 2.12 | An app made from a template (IT Tools, found by searching the catalog) is deployed; its health check passes                                                                                                                                   | `tests/level-2/templates.spec.ts`           |
| 2.13 | A data backup of an app's volume into a repository on a volume is a snapshot; restored (Replace), the app finds its data as backed up                                                                                                         | `tests/level-2/backups.spec.ts`             |
| 2.14 | A function made from the runtime's template builds, and a Run answers 200 with what it was asked                                                                                                                                              | `tests/level-2/functions.spec.ts`           |
| 2.15 | A health check runs on its interval: Done when answered as asked, Failed when not                                                                                                                                                             | `tests/level-2/health-checks.spec.ts`       |
| 2.16 | An app exposed at a domain from Routing Settings answers there over HTTPS; a new domain forces HTTPS, and HTTP is sent there                                                                                                                  | `tests/level-2/routing.spec.ts`             |
| 2.17 | Without Force HTTPS, the app answers plain HTTP as well                                                                                                                                                                                       | `tests/level-2/routing.spec.ts`             |
| 2.18 | Basic auth on a domain: 401 without the credentials or with a wrong password, 200 with them                                                                                                                                                   | `tests/level-2/routing.spec.ts`             |
| 2.19 | A domain set to Redirect To sends its visitors to the main one, path and query kept                                                                                                                                                           | `tests/level-2/routing.spec.ts`             |
| 2.20 | Headers added on the way in and on the way out; a path prefix stripped                                                                                                                                                                        | `tests/level-2/routing.spec.ts`             |
| 2.21 | A client outside the Allowed IPs gets 403; a request over the rate limit gets 429                                                                                                                                                             | `tests/level-2/routing.spec.ts`             |
| 2.22 | A function created at a domain answers there, sent to HTTPS                                                                                                                                                                                   | `tests/level-2/functions.spec.ts`           |
| 2.23 | A health check of an app at its own domain passes                                                                                                                                                                                             | `tests/level-2/health-checks.spec.ts`       |
| 2.24 | A project made from a Compose file: its services run as apps, a variable its .env gives reaches the container, a published port answers at the domain suggested                                                                               | `tests/level-2/compose.spec.ts`             |
| 2.25 | A project exported from its Operations page, deleted, and imported on the installation's runs as it was: its app, its variable, its domain                                                                                                    | `tests/level-2/spec.spec.ts`                |
| 2.26 | Exporting secrets is refused while the server does not return them, and the page says why                                                                                                                                                     | `tests/level-2/spec.spec.ts`                |
| 2.27 | A deployment of an image that does not exist fails, and its card says the image was not found; the app runs on in the container it ran                                                                                                        | `tests/level-2/deploy-failures.spec.ts`     |
| 2.28 | A container that keeps exiting is on Home's Needs attention - keeps restarting, with its exit code - and View logs shows what it printed                                                                                                      | `tests/level-2/deploy-failures.spec.ts`     |
| 2.29 | A project's Slack platform tried with Test Send Msg, then saved, and a notification target through it                                                                                                                                         | `tests/level-2/notifications.spec.ts`       |
| 2.30 | A health check telling a target: nothing while it is healthy, once when it fails, once when it passes again                                                                                                                                   | `tests/level-2/notifications.spec.ts`       |
| 2.31 | Repeat while failing tells a check that still fails again                                                                                                                                                                                     | `tests/level-2/notifications.spec.ts`       |
| 2.32 | An upload from the terminal goes over a websocket and lands whole; a file never replaces a directory; refusals say why; an expired session is refreshed                                                                                       | `tests/level-2/container-files.spec.ts`     |
| 2.33 | Autoscale is not turned on while neither of what it scales on can be read: the page says why, sends nothing, and the API refuses it                                                                                                           | `tests/level-2/autoscale.spec.ts`           |
| 2.34 | A Compose file added to a project: a service named as an app of the env uses that app, another is created under another key and reaches the env's app by name; nothing the env had changes                                                    | `tests/level-2/compose.spec.ts`             |
| 2.35 | A Compose folder opened: its secret, its mounted file and the files of its mounted directory reach the services; its Traefik labels make a domain, which serves                                                                               | `tests/level-2/compose.spec.ts`             |
| 2.36 | A template that depends on a database (Atuin, postgres) deploys both; the database is listed under the app; an account made through the app is read back                                                                                      | `tests/level-2/templates.spec.ts`           |
| 2.37 | A template made again where a deleted app left its data asks first: Create anyway keeps the data, Delete that data and create starts without it                                                                                               | `tests/level-2/templates.spec.ts`           |
| 2.38 | A deployment in progress (held by its pre-deployment command) is canceled from its card; the app runs on in the container it ran                                                                                                              | `tests/level-2/deployments.spec.ts`         |
| 2.39 | Deployments are listed newest first; an older one opened from its card has its own log, without what a later one's pre-deployment command printed                                                                                             | `tests/level-2/deployments.spec.ts`         |
| 2.40 | Resource limits saved - CPUs, memory, swap 0, PIDs - are the container's cgroup's; memory asked beyond them is killed (137)                                                                                                                   | `tests/level-2/resources.spec.ts`           |
| 2.41 | Variables of the project, the environment and another app (shared), and HivePaaS's own, reach the container, the nearest winning; the page lists the inherited                                                                                | `tests/level-2/env-vars.spec.ts`            |
| 2.42 | An app built from a repository (main, at a commit); a push of the next commit through the webhook builds and deploys it, marked Webhook; another branch, a bad signature (401) and the push delivered again do not                            | `tests/level-2/webhooks.spec.ts`            |
| 2.43 | A scheduled job on a cron runs on it, minute after minute                                                                                                                                                                                     | `tests/level-2/jobs.spec.ts`                |
| 2.44 | A deployment tells its notification target that it failed, and why, and that it succeeded: the app, the image, a link to the deployment                                                                                                       | `tests/level-2/notifications.spec.ts`       |
| 2.45 | An upload cancelled from its dialog stops at once, says what may be left partly written, and the dialog takes another                                                                                                                         | `tests/level-2/container-files.spec.ts`     |
| 2.46 | An app set to Git Source in Deployment Settings - a repository, branch develop, Deploy on Push off - builds the branch's commit and runs it; Deploy on Push is saved off                                                                      | `tests/level-2/git.spec.ts`                 |
| 2.47 | A repository with no Dockerfile, Auto-Generate picked: built with the Dockerfile HivePaaS writes, a static web server, it serves the page at its domain                                                                                       | `tests/level-2/git.spec.ts`                 |
| 2.48 | Build-time variables: the build's step shows the argument, the container prints it; one that uses an app secret is a BuildKit secret, read from a file, in no log                                                                             | `tests/level-2/git.spec.ts`                 |
| 2.49 | A private repository over HTTPS: refused without credentials (ERR_REPO_NOT_FOUND); built with an access token picked in Git Credentials                                                                                                       | `tests/level-2/git.spec.ts`                 |
| 2.50 | The same repository over SSH, built with a key                                                                                                                                                                                                | `tests/level-2/git.spec.ts`                 |
| 2.51 | Previews and PR comments on in Feature Settings: a stranger's "/hivepaas deploy" on a pull request makes nothing; the owner's makes pr-7, built from the pull request's head, its /data a directory of its own; "/hivepaas cancel" removes it | `tests/level-2/webhooks.spec.ts`            |
| 2.52 | A job sequence, watched as it runs: a step hands its output to the next; a step that fails stops it, the rest skipped                                                                                                                         | `tests/level-2/jobs.spec.ts`                |
| 2.53 | A job triggered after a deploy runs once the deploy is done, told the event and the deployment                                                                                                                                                | `tests/level-2/jobs.spec.ts`                |
| 2.54 | A failing job is retried as many times as asked (Retries: 2), then failed; a slow one stops at its timeout                                                                                                                                    | `tests/level-2/jobs.spec.ts`                |
| 2.55 | A job turned off does not run on its cron; turned on again, it runs at the next minute                                                                                                                                                        | `tests/level-2/jobs.spec.ts`                |
| 2.56 | A database dumped by a data backup job's command is a snapshot; restored, the restore command loads it back                                                                                                                                   | `tests/level-2/backups.spec.ts`             |
| 2.57 | A scheduled job that fails tells the target its form names: Scheduled task failed                                                                                                                                                             | `tests/level-2/notifications.spec.ts`       |
| 2.58 | A health check reads the answer's body: a text by its pattern, a JSON by what it contains; others fail                                                                                                                                        | `tests/level-2/health-checks.spec.ts`       |
| 2.59 | A check left waiting fails at its timeout and holds no other; a check turned off sees no change                                                                                                                                               | `tests/level-2/health-checks.spec.ts`       |
| 2.60 | Compression: an app's first domain compresses from the start; an answer over the size set is compressed (zstd, br or gzip), one under it is not, nor one of a type left out                                                                   | `tests/level-2/routing-middlewares.spec.ts` |
| 2.61 | Circuit breaker: failing answers past its expression open it - it answers in the app's place, with the code set - and once its fallback is over the app answers again                                                                         | `tests/level-2/routing-middlewares.spec.ts` |
| 2.62 | Client: a client within the Allowed IPs is let in; a body over Max Request Body Size is refused (413), one under it is not                                                                                                                    | `tests/level-2/routing-middlewares.spec.ts` |
| 2.63 | Headers: one removed from requests does not reach the app, one removed from answers does not reach the browser; Auto Detect gives an untyped answer its type                                                                                  | `tests/level-2/routing-middlewares.spec.ts` |
| 2.64 | Path rewrite: a prefix added, a prefix stripped by a pattern, a path replaced by a pattern with its groups                                                                                                                                    | `tests/level-2/routing-middlewares.spec.ts` |
| 2.65 | Path rewrite: a path written out is replaced, and what is under it (/old/page to /new/page); /older and other paths are not                                                                                                                   | `tests/level-2/routing-middlewares.spec.ts` |
| 2.66 | Rate limit: over Max In-Flight Requests, a request made while another is answered is refused (429)                                                                                                                                            | `tests/level-2/routing-middlewares.spec.ts` |
| 2.67 | A setting turned off is saved, kept off, and not applied: an allowlist off turns no one away                                                                                                                                                  | `tests/level-2/routing-middlewares.spec.ts` |
| 2.68 | An app cloned with its volumes' data, the source stopped for it: the clone runs on a copy - the source's files and its own - and the source, started again, on its own data                                                                   | `tests/level-2/clone.spec.ts`               |
| 2.69 | An app cloned with its volumes but not their data starts on empty directories of its own; the source's data stays its own                                                                                                                     | `tests/level-2/clone.spec.ts`               |
| 2.70 | An app cloned without its volumes has none - Persistent Storage lists nothing - and the source reads nothing of the clone's                                                                                                                   | `tests/level-2/clone.spec.ts`               |
| 2.71 | Subpaths of a volume: two are directories of their own, one mounted twice is the same directory; read only, it refuses writes                                                                                                                 | `tests/level-2/storage.spec.ts`             |
| 2.72 | Data of another app: it reads that app's files and cannot write them; Allow writing, it can; the owner's Persistent Storage names it, reading or changing                                                                                     | `tests/level-2/storage.spec.ts`             |
| 2.73 | A mount moved to another target keeps its data; removed, the container has none; mounted again, "This storage already has data" keeps it (Save anyway) or deletes it                                                                          | `tests/level-2/storage.spec.ts`             |
| 2.74 | Reset permissions: given to a user (1000:1000) the files are that user's, modes kept; opened to every user, 666                                                                                                                               | `tests/level-2/storage.spec.ts`             |
| 2.75 | A volume bound to a directory of the node is mounted as a bind, and keeps what the app wrote for the next container                                                                                                                           | `tests/level-2/storage.spec.ts`             |
| 2.76 | A file mounted from a config file follows it: changed, the container reads the new content without a deploy; the mount turned off, the file is gone; on again, it is back                                                                     | `tests/level-2/setting-mounts.spec.ts`      |
| 2.77 | A certificate and its key mounted from an SSL certificate, the key 0400; renewed with Renew Now, the container has the new pair without a deploy                                                                                              | `tests/level-2/setting-mounts.spec.ts`      |
| 2.78 | A basic auth's htpasswd mounted: Apache's htpasswd accepts its password, and no other; the env's own basic auth not inheritable, which the app cannot use, is not offered                                                                     | `tests/level-2/setting-mounts.spec.ts`      |
| 2.79 | A path one setting mount has is refused to another; a config file a mount reads is not deleted, and the dialog lists the mount, a link to it                                                                                                  | `tests/level-2/setting-mounts.spec.ts`      |
| 2.80 | A deployment running is in the app's header - Deploying, which opens it - and marked on its Deployments tab; once it ends, neither                                                                                                            | `tests/level-2/deployments.spec.ts`         |
| 2.81 | A variable saved and a secret changed reach the running container without a deploy                                                                                                                                                            | `tests/level-2/variables-secrets.spec.ts`   |
| 2.82 | A preview takes its app's variables as they change; a secret kept from previews is empty there, and the preview is made                                                                                                                       | `tests/level-2/variables-secrets.spec.ts`   |
| 2.83 | Variables refused: a reference to a secret that does not exist, a circular one; a secret a variable uses is not deleted                                                                                                                       | `tests/level-2/variables-secrets.spec.ts`   |
| 2.84 | A project's secret and an env's config file made for its apps reach them; an env's secret not made for them is refused, not available here                                                                                                    | `tests/level-2/variables-secrets.spec.ts`   |
| 2.85 | A secret's value is never shown; Reveal Secret and Download File are refused while the server returns no secrets                                                                                                                              | `tests/level-2/variables-secrets.spec.ts`   |
| 2.86 | A literal variable reaches the container as written, a reference worked out; Show Final Values shows both, a secret masked                                                                                                                    | `tests/level-2/variables-secrets.spec.ts`   |
| 2.87 | Link App: an app reaches another at the address it adds; a shared variable changed reaches the app that uses it without a deploy                                                                                                              | `tests/level-2/variables-secrets.spec.ts`   |
| 2.88 | A binary secret and a binary config file, uploaded, are mounted byte for byte; a config file downloads as it was                                                                                                                              | `tests/level-2/variables-secrets.spec.ts`   |
| 2.89 | A member without Can Reveal Secrets finds a basic auth's password and htpasswd locked, the API refusing them, and mounts its username; given it, signed in again, the htpasswd                                                                | `tests/level-2/setting-mounts.spec.ts`      |

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

- **Git, for real**: a repository on GitHub, a GitHub App, a webhook GitHub
  calls itself, the comments HivePaaS posts on a pull request. The env is
  offline: these need GitHub, and an account of the tests' own there.
- **A build pushed to a registry**: waits for a registry in the env.
- **Metrics and autoscale on load**: the HTTP and CPU charts, and an app and a
  function scaling out under load and back in. They read the stored logs, which
  the env does not run yet: `env/up.sh` is to set up Logging (VictoriaLogs and
  its agent) first. Until then autoscale can be checked only refusing (2.33).
- **Routing, the rest**: a TCP domain (a database through the proxy, by SNI),
  settings of one path (Paths), load balancing, websockets, a certificate of
  one's own; Mem Request Body Size, which nothing outside the proxy sees.
- **Secrets in a spec**: exported encrypted and imported with the passphrase.
  The server refuses to return secrets unless its Security settings allow it,
  and turning that on for one test would unmask them for the tests running
  beside it.

The Git repositories apps are built from are served inside dind (see the
README): their commits are made at fixed dates, so their hashes are fixed too.
A test plays GitHub's webhook - a push of a commit, a comment on pull request
7 - signing it with the webhook's secret as GitHub does. GitHub itself is not
called: what HivePaaS posts back on a pull request fails, and is let go.

## Found with variables and secrets (2026-10-10), and fixed

- **Download File handed an app's secret to anyone who could read the app**:
  it took neither the operator's switch nor the Can Reveal Secrets
  capability, and left no record, though Reveal Secret takes both and records
  every attempt. It goes through the same gate now; an inherited secret is not
  downloaded from below.
- **A preview of an app with a variable was not made**: the task panicked,
  "invalid memory address or nil pointer dereference". Its variables were
  rewritten for the databases cloned with it, with a replacer that exists only
  when there are some.
- **A preview kept the variables it was made with**: saving the app's, or a
  secret they use, applied the app's and none of its previews'. The check
  before the previews was the wrong way round.
- **A preview of an app with a secret kept from previews was not made**: the
  variable built from it, empty in the preview as the pull request's comment
  says, failed the preview's task instead. It is empty now, and no error.

Seen, and left as they are:

- A runtime variable named like HivePaaS's own - `HIVEPAAS_HOST` - is dropped
  without a word: the save says it is updated. The request is cleaned of such
  names before it is checked, so the check that refuses them never sees one.
- A reference to a project's secret not made for its apps says the secret is
  missing, not that it is "not available here": the names held back are those
  of the scope just above.
- Making a preview answers the id of the task that makes it, not the
  preview's: a test finds the preview in the list.
- A secret made through the API is not for the scopes below unless it says
  so; the dashboard's form ticks it, and the CLI's `secret set` gives it to
  previews unless told `--no-previews`.

## Found with setting mounts (2026-10-10), and fixed

- **A certificate renewed by hand was not saved**: Renew Now obtained a new
  one and wrote Traefik's files from it, but the setting kept the old one. An
  app mounting the certificate kept the old pair, a restart that rewrote the
  files went back to it, and the renewal job saw it still about to expire. It
  is saved now as an update is, and the apps that mount it follow it.
- **What reads a setting could not be listed** below the installation: a
  config file a mount reads was not deleted, but the dialog said "The list of
  what uses it could not be loaded". The usages route was under `/settings`
  only. Every settings group of a project, an environment and an app has it
  now, a setting of another one not found through it; and a mount using a
  setting is a link to the app's Setting Mounts.
- **A picker offered settings the app could not use**: a mount's source, and
  Routing's basic auth and certificate, listed the environment's own ones not
  inheritable too, and saving one answered "Setting '...' is not found". They
  list those the app can use.

Seen, and left as they are:

- A certificate is named by its domain: a name given when it is made is not
  kept.

## Found with Persistent Storage (2026-10-10), and fixed

- **Reset permissions took a second click**: its user and group ids were
  number fields, which take their number only once left; the button, disabled
  until both were there, was still disabled when the second was typed, and
  the click that left the field did nothing. The dialog follows what is typed
  now.
- **A number field was "Number input" to a screen reader** whatever its label
  said: the label's `for` named an id the field never had, and its fallback
  name won anyway. A field given an id now has it, and its label names it.

## Found cloning apps with volumes (2026-10-10), and fixed

- **A clone without its volumes used the source app's**: its service was
  copied from the source's, mounts and all, and nothing took them away. The
  clone read and wrote the source app's directories as its own - two databases
  on one data directory. It has no storage now.
- **A pull request's preview wrote to the app's data**: a preview is a clone,
  and was made without its volumes - so with the app's. Its storage is its
  own now, empty: directories named after it, which go with it.
- **A clone with its volumes but not their data never started**: its
  directories were made only by the copy of the data, and a volume's subpath
  that is not there is not mounted. They are made either way now.
- **The env could not copy a volume**: rsync, which the release's agent image
  has, was not in dind. `env/up.sh` installs it.

## Found with the routing settings (2026-10-10), and fixed

- **A path replaced, not by a pattern, sent every request to it**: "Replace
  Path /old With /new" became Traefik's ReplacePath of /old - which replaces
  every path the domain takes with /old - and /new was never read. It replaces
  /old and what is under it with /new now, /old/page with /new/page, the path
  taken as written: /older and the other paths are left alone.

Seen, and left as they are:

- An app's first domain compresses its answers from the start (past 1 KB):
  the form adds the section to it.
- An answer the app writes in pieces - whoami's /data - is compressed whatever
  its size: the proxy cannot wait to know it. The test sizes whoami's answer by
  the request's headers, which it echoes in one piece.
- A Content-Type removed from an answer comes back on a small one: the proxy's
  server names the type again. 2.63 removes a header of the app's own.
- After hours of runs, dind's swarm stopped placing tasks: every new one stayed
  New, and every domain answered 404. dockerd logged "could not find network
  allocator state for network ..." for networks of projects long deleted -
  swarm's own. `yarn env:down && yarn env:up` gives a swarm that places them.

## Found with Git sources (2026-10-09), and fixed

- **A Dockerfile generated for a static site never built**: its image,
  `joseluisq/static-web-server:2-debian`, runs as a user of its own (`sws`)
  since 2.44, and the Dockerfile's `apt-get` failed as that user ("Permission
  denied"). Every repository with no Dockerfile of its own that is served
  statically - plain HTML, an Astro or a Nuxt site built to files - failed to
  build. The packages are installed as root now, and the server runs as the
  image's user again (dockerfile-generator).

## Found with the jobs and the health checks (2026-10-09), and fixed

- **A job sequence's log stopped after its first step** for anyone following
  it - the run's page, `hivepaas logs -f` - though the sequence went on. Each
  step runs as one more run of its task, and every run ended the log: it wrote
  "Job execution finished" and told those following that the log was closed.
  A run that goes on now hands its lines on and starts the log's list again,
  telling those following to read the new one from its start; the last run
  ends the log. A log grown past its size limit starts its list again the same
  way: those following it lost what came after.
- **A job turned on again stayed silent for up to ten minutes**: it was
  scheduled as it was before the change - off - so its runs were made only at
  the next scan; and its coming runs, canceled when it was turned off, stood
  in the way of the same runs made anew. A status change schedules the job as
  it is now, and a coming run that never started is removed.

Seen, and left as they are:

- A health check's run is kept when its result changes, not every run: a
  check that stays healthy shows one run. A test that a check goes on running
  makes its result change.
- A run's page asks for the run's status every 5 seconds once its log has
  ended: "Done" can come that late.
- A run that failed and waits for its retry reads Failed on its page, which
  then asks no more: the retries show once the page is opened again.

## Found in the third round (2026-10-09), and fixed

- **A push never deployed anything**: the webhook answered 200, and no
  deployment was made, nor anything logged. The push found its apps, then
  checked each was active by reading its `id` alone - so its status read
  empty, and every app inactive - and the error was dropped. The check reads
  the status now, and an app a push does not deploy is logged with why.
- **A build that failed on a build node panicked the deployment** ("index out
  of range [0] with length 0"), its reason lost. An error of no kind HivePaaS
  knows - here the Dockerfile generator's - became gRPC's code 0, OK, and
  `status.Error(OK)` is no error: the agent answered as if the build had
  succeeded, with no image, and updating the service read the first of none.
  Such an error is Unknown now, its message kept; and a build answering no
  image is a failed one.
- **A push not signed with the webhook's secret answered 500**, logged as the
  server's own failure: it answers 401, `ERR_WEBHOOK_UNVERIFIED`, "The
  delivery is not signed with the webhook's secret".
- **A pre-deployment command right after a deployment failed** "Running task
  of service not found": it runs in a container that has run 15 seconds, and
  the deployment looked three times, five seconds apart, for one - the last
  wait never followed by a look. A container running, not that long yet, is
  waited for now until it has, 15 seconds more at most; with none running the
  looks are as many as before, the last one after the last wait; a cancel
  ends the wait. With none found, the deployment says so: "No active container
  found for app 'web'". The same holds for a scheduled job, a command pipe and
  a preview's commands.
- **A failed deployment's notification did not say why**: it has a Reason on
  every channel now, as the deployment's details say it, cut to 1000
  characters.
- **A notification in JSON - Slack, Discord, Lark - was not sent when a name had
  a quote in it**: a project's or an app's name, a commit message, a health
  check's answer was written into the body as it was, and quotes, new lines or
  control characters broke it. Every value goes through `json` now; a test
  renders each of the 18 templates with such text.
- **Swappiness set without swap read back as unset**: the settings page took it
  only when swap was set too.
- **Accessibility**: the Compose page's choice for a service an app of the env
  is named ("What to do with web") and so is its key's field; a template's
  parameters are groups named by their titles.

Found, and left for review:

- **Swap Memory is swap beyond the memory limit** - swarm's `SwapBytes`: 96 MB
  with 64 MB of memory is 96 MB of swap - while its help said "total memory
  plus swap". The help says what it is now; unset, it is as much as the
  memory, and 0 is none.

Seen, and left as they are:

- A checkout that fails says "exit status 128"; git's words are in the log.
- A notification target's own platform is looked up as the app sees it: a
  target the project's apps may use, sending through a platform only the
  project may, cannot be chosen for an app ("Setting ... is not found").
- Dockerfile "auto" generates one: for a repository it finds no language in,
  it says "A Dockerfile was not detected", whether there is one or not.
- A repository served over `git://` has its commits linked at
  `https://<host>/.../commit/<hash>`.
- A domain just set answers in 10 to 40 seconds, more on a busy node: the
  proxy reads the swarm every 15 seconds, and the app's container is made anew
  to join its network.

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
