# Docker API Access - Templates Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** The two templates the Docker API was built for, `autobase` and `gitea-runner`, in the `app-templates` repository.

**Architecture:** Each is one YAML file and one icon.
- **What `settings.dockerApi` gives them:** the images their children run, what the children share with the app, and what else they may do. The app reaches its socket through `${HIVEPAAS_DOCKER_HOST}`.
- **Tooling:** `apptemplate lint` checks both against the buildable subset of plan 3, and `apptemplate index` regenerates `index.json`.

**Tech Stack:** YAML templates; the `apptemplate` tool in the hivepaas repository.

**Spec:** `docs/superpowers/specs/2026-09-24-docker-api-access-design.md` (§11).

## Global Constraints

- Templates live in `/Users/tnt/go/src/github.com/hivepaas/app-templates` and follow its README. `index.json` is generated, never edited.
- **Images:** pinned exactly. `autobase/console:2.11.0`, `gitea/act_runner:0.6.1`.
- **Docker API blocks** carry no placeholders, since they are read from the template and shown before anybody deploys.
- **What was measured for this plan:**
  - The console image takes `POSTGRES_PASSWORD` and `PG_CONSOLE_DB_PASSWORD` for its bundled database, which listens on all interfaces with `postgres-pass` unless told otherwise. Both are therefore set to one generated secret.
  - Its API answers `/api/v1/version` with 401 without the token and 200 with it, six seconds after start.
  - act_runner reads its config with no environment expansion, and a job's cache server is not reachable from a job's own network, so the cache is off.
- **Git:** work in `app-templates` on branch `feat/docker-api-templates`. Every commit ends with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. Merge into its `main` locally and delete the branch. Do not push.

---

### Task 1: `autobase`

**Files:**
- Create: `templates/autobase.yaml`
- Create: `icons/autobase.svg` (from `autobase-tech/autobase`, `console/ui/src/shared/assets/AutobaseLogo.svg`, MIT)

- [ ] **Step 1: Create the branch and the icon**

```bash
cd /Users/tnt/go/src/github.com/hivepaas/app-templates
git checkout -b feat/docker-api-templates
gh api "repos/autobase-tech/autobase/contents/console/ui/src/shared/assets/AutobaseLogo.svg" \
  -H "Accept: application/vnd.github.raw" > icons/autobase.svg
```

- [ ] **Step 2: Write the template**

`templates/autobase.yaml`:

```yaml
apiVersion: hivepaas.com/v1
kind: AppTemplate

metadata:
  name: autobase
  title: Autobase
  tagline: Highly available PostgreSQL clusters on your own servers, from a web console
  description: |
    Autobase deploys and runs PostgreSQL clusters - Patroni for failover, etcd, HAProxy and
    PgBouncer, pgBackRest for backups - on servers you give it, or on cloud machines it creates
    for you. This template is its console: the web page, its API, and the small database the
    console keeps its inventory in.

    **This app is given the Docker API, and that is how it works**

    Every operation - creating a cluster, adding a replica, upgrading PostgreSQL - is an Ansible
    playbook the console runs in a container of the `autobase/automation` image. Upstream installs
    the console with the Docker socket mounted, which makes it root on its node. Here it reaches
    the Docker API through HivePaaS instead, limited to what that needs:

    - it may run the `autobase/automation` image and nothing else;
    - the only part of this app's storage those containers see is the directory the playbook
      writes its log to;
    - they get no privileges, no path of the host and no host network, and reach your servers
      over SSH the way anything on the internet does;
    - at most five exist at once, with 2 GB of memory and two CPUs each.

    Creating this app needs Write permission on the Cluster module.

    **Signing in**

    The console asks for a token rather than a password: the **Console token** below, generated
    when left empty. Read it in this app's secrets.

    **What to expect**

    - The first operation downloads the automation image, about 3 GB, before it starts.
    - The servers a cluster goes on have to accept SSH from this node's address, with a key you
      give the console.
    - The console writes its own logs to files inside the container, so the app's log view shows
      little. The log of every operation is in the console.
  categories: [databases/sql]
  tags: [sql, relational, automation, open-source]
  aliases: [postgresql cluster, postgres ha, patroni, high availability, dbaas]
  icon: icons/autobase.svg
  links:
    website: https://autobase.tech
    documentation: https://autobase.tech/docs
    source: https://github.com/autobase-tech/autobase
  license: MIT
  requires:
    versionCode: v000001

parameters:
  - name: domain
    title: Domain
    description: >-
      The address the console answers at. Leave it empty to add one later, in the app's routing
      settings.
    type: domain
    optional: true
  - name: consoleToken
    title: Console token
    description: What the console asks for to sign in. Leave it empty and one is generated.
    type: secret
    generate: {length: 32}
  - name: encryptionKey
    title: Encryption key
    description: >-
      Encrypts the SSH keys and cloud credentials the console stores. Leave it empty and one is
      generated; it has to stay the same for as long as the console keeps them.
    type: secret
    generate: {length: 48}
  - name: dbPassword
    title: Console database password
    description: The password of the console's own database. Leave it empty and one is generated.
    type: secret
    generate: {length: 32}
  - name: dataVolume
    title: Data volume
    description: The console's database, its database viewer's settings, and the operations' logs.
    type: volume
  - name: memoryLimit
    title: Memory limit
    description: What the console gets. Its database holds an inventory, not your data.
    type: size
    default: 1GB

versions:
  - name: "2.11"
    release: "2.11.0"
    default: true
    image: autobase/console:2.11.0

app:
  deployment:
    source:
      activeMethod: image
      imageSource: {image: "${{ image }}"}
    storage:
      mounts:
        /var/lib/postgresql: {type: volume, source: "${{ params.dataVolume }}", volumeOptions: {subpath: postgresql}}
        /opt/dbdesk-studio/data: {type: volume, source: "${{ params.dataVolume }}", volumeOptions: {subpath: dbdesk}}
        # Shared with the automation containers: the playbook writes its log here, and the
        # console reads it to follow the operation.
        /var/lib/autobase/ansible: {type: volume, source: "${{ params.dataVolume }}", volumeOptions: {subpath: ansible}}
    container:
      healthcheck:
        enabled: true
        mode: CMD-SHELL
        # The API refuses a request without the token, so a healthy answer needs it.
        command: >-
          curl -fsS -o /dev/null -H "Authorization: Bearer $PG_CONSOLE_AUTHORIZATION_TOKEN"
          http://127.0.0.1/api/v1/version
        interval: 30s
        timeout: 10s
        startPeriod: 120s
        retries: 5
    resources:
      limits: {memory: "${{ params.memoryLimit }}"}
  settings:
    kind:
      category: webapp
      engine: autobase
      version: "${{ version.name }}"
      webapp: {}
    envVars:
      data:
        - {k: PG_CONSOLE_AUTHORIZATION_TOKEN, v: "${secrets.CONSOLE_TOKEN}"}
        - {k: PG_CONSOLE_ENCRYPTIONKEY, v: "${secrets.ENCRYPTION_KEY}"}
        # The bundled database listens on every interface; without these it would take the
        # image's default password from anything on the project's network.
        - {k: POSTGRES_PASSWORD, v: "${secrets.CONSOLE_DB_PASSWORD}"}
        - {k: PG_CONSOLE_DB_PASSWORD, v: "${secrets.CONSOLE_DB_PASSWORD}"}
        - {k: PG_CONSOLE_DOCKER_HOST, v: "${HIVEPAAS_DOCKER_HOST}"}
        - {k: PG_CONSOLE_DOCKER_IMAGE, v: "autobase/automation:${{ version.release }}"}
        - {k: PG_CONSOLE_DOCKER_LOGDIR, v: /var/lib/autobase/ansible}
        - {k: PG_CONSOLE_LOGGER_LEVEL, v: INFO}
    secrets:
      CONSOLE_TOKEN:
        value: "${{ params.consoleToken }}"
      ENCRYPTION_KEY:
        value: "${{ params.encryptionKey }}"
      CONSOLE_DB_PASSWORD:
        value: "${{ params.dbPassword }}"
    routing:
      port: 80
      exposePublicly: true
      domains:
        - {domain: "${{ params.domain }}", enabled: true, forceHttps: true}
    dockerApi:
      images: [autobase/automation]
      sharedDirs: [/var/lib/autobase/ansible]
      limits: {containers: 5, memory: 2gb, cpus: 2}
```

- [ ] **Step 3: Lint and render**

Run, from the hivepaas repository:

```bash
cd /Users/tnt/go/src/github.com/hivepaas/hivepaas
go run ./tools/apptemplate lint ../app-templates
go run ./tools/apptemplate render -param dataVolume=vol-1 ../app-templates autobase
```

Expected: lint reports no problem. The render shows the `dockerApi` block unchanged, and `PG_CONSOLE_DOCKER_IMAGE` as `autobase/automation:2.11.0`.

---

### Task 2: `gitea-runner`

**Files:**
- Create: `templates/gitea-runner.yaml`
- Create: `icons/gitea-runner.svg` (Gitea's mark, as `icons/gitea.svg` has it)

- [ ] **Step 1: The icon**

```bash
cp icons/gitea.svg icons/gitea-runner.svg
```

- [ ] **Step 2: Write the template**

`templates/gitea-runner.yaml`:

```yaml
apiVersion: hivepaas.com/v1
kind: AppTemplate

metadata:
  name: gitea-runner
  title: Gitea Actions Runner
  tagline: Runs the Actions workflows of a Gitea in the same environment
  description: |
    Gitea Actions are GitHub Actions-compatible workflows, and a Gitea runs none of them by
    itself: a runner picks jobs up and runs each in containers of its own. This is Gitea's
    runner, act_runner, for a Gitea app in the same environment.

    **This app is given the Docker API, and that is how it works**

    Every job is a container, and so is every service a job asks for. Upstream installs the
    runner with the Docker socket mounted, which makes it root on its node. Here it reaches the
    Docker API through HivePaaS instead:

    - jobs run any image, but with no privileges, no path of the host and no host network;
    - a job may use Docker itself - `docker build`, `docker run` - and what it starts goes
      through the same limits;
    - at most 20 containers exist at once, with 4 GB of memory and two CPUs each.

    A workflow that needs `--privileged` does not run here. Creating this app needs Write
    permission on the Cluster module.

    **Registering**

    In Gitea, open Site Administration, then Actions, then Runners, and create a runner: the
    token it shows is the **Registration token** below. The runner registers once, on its first
    start, and keeps what it was given on its data volume.

    **What jobs can reach**

    - Jobs run on networks of their own, not the project's. They clone from the address Gitea
      tells them, its ROOT_URL, so that Gitea needs a domain the jobs can reach over the internet.
    - `actions/cache` finds no cache server: the runner's own cannot be reached from the jobs'
      networks, so it is off.
  categories: [webapps/dev-tools]
  tags: [automation, testing, git, open-source]
  aliases: [act_runner, act runner, ci, actions, gitea actions]
  icon: icons/gitea-runner.svg
  links:
    website: https://docs.gitea.com/usage/actions/overview
    documentation: https://docs.gitea.com/usage/actions/act-runner
    source: https://gitea.com/gitea/act_runner
  license: MIT
  requires:
    versionCode: v000001

parameters:
  - name: giteaApp
    title: Gitea
    description: The Gitea app in this environment whose workflows this runner runs.
    type: app
    engine: gitea
  - name: registrationToken
    title: Registration token
    description: >-
      The token Gitea shows when you create a runner, under Site Administration, Actions,
      Runners - or in a repository's or organization's settings, for a runner of its own.
    type: secret
  - name: labels
    title: Labels
    description: >-
      What jobs ask for in runs-on, and the image each runs in, comma separated. The default
      runs ubuntu-latest, ubuntu-24.04 and ubuntu-22.04 in Gitea's runner images.
    type: string
    default: >-
      ubuntu-latest:docker://docker.gitea.com/runner-images:ubuntu-latest,ubuntu-24.04:docker://docker.gitea.com/runner-images:ubuntu-24.04,ubuntu-22.04:docker://docker.gitea.com/runner-images:ubuntu-22.04
  - name: capacity
    title: Jobs at once
    description: How many jobs this runner runs at the same time.
    type: int
    default: 1
    min: 1
    max: 4
  - name: dataVolume
    title: Data volume
    description: The runner's registration.
    type: volume
  - name: memoryLimit
    title: Memory limit
    description: What the runner itself gets. Its jobs are containers of their own, with limits of their own.
    type: size
    default: 512MB

versions:
  - name: "0.6"
    release: "0.6.1"
    default: true
    image: gitea/act_runner:0.6.1

app:
  deployment:
    source:
      activeMethod: image
      imageSource: {image: "${{ image }}"}
    storage:
      mounts:
        /data: {type: volume, source: "${{ params.dataVolume }}"}
    container:
      healthcheck:
        enabled: true
        mode: CMD-SHELL
        # The runner serves nothing to ask; what it keeps once registered is what says it
        # got past its start.
        command: test -s /data/.runner
        interval: 30s
        timeout: 5s
        startPeriod: 120s
        retries: 5
    resources:
      limits: {memory: "${{ params.memoryLimit }}"}
  settings:
    kind:
      category: webapp
      engine: gitea-runner
      version: "${{ version.name }}"
      webapp: {}
    envVars:
      data:
        # The runner reaches Gitea on the project's network; the jobs reach it by its ROOT_URL.
        - {k: GITEA_INSTANCE_URL, v: "http://${${{ params.giteaApp }}.HIVEPAAS_HOST}:${${{ params.giteaApp }}.HIVEPAAS_PORT}"}
        - {k: GITEA_RUNNER_REGISTRATION_TOKEN, v: "${secrets.REGISTRATION_TOKEN}"}
        - {k: GITEA_RUNNER_NAME, v: "${HIVEPAAS_APP_NAME}"}
        - {k: GITEA_RUNNER_LABELS, v: "${{ params.labels }}"}
        - {k: CONFIG_FILE, v: /etc/act_runner/config.yaml}
        - {k: DOCKER_HOST, v: "${HIVEPAAS_DOCKER_HOST}"}
    secrets:
      REGISTRATION_TOKEN:
        value: "${{ params.registrationToken }}"
    configFiles:
      config.yaml:
        content: |
          log:
            level: info
          runner:
            file: .runner
            capacity: ${{ params.capacity }}
            timeout: 3h
          cache:
            # A job runs on a network of its own and cannot reach the runner's cache server.
            enabled: false
          container:
            # Empty: every job gets a network of its own, where its services answer by name.
            network: ""
            privileged: false
            # Empty: jobs are given the runner's Docker API, and use Docker through it.
            docker_host: ""
        swarmRef:
          file: {name: /etc/act_runner/config.yaml, mode: 444}
    dockerApi:
      images: ["*"]
      allow: [exec, files, volumes, networks, nestedSocket]
      limits: {containers: 20, memory: 4gb, cpus: 2}
```

- [ ] **Step 3: Lint and render**

```bash
cd /Users/tnt/go/src/github.com/hivepaas/hivepaas
go run ./tools/apptemplate lint ../app-templates
go run ./tools/apptemplate render -param dataVolume=vol-1 -param giteaApp=git -param registrationToken=t ../app-templates gitea-runner
```

Expected: no problem; the instance URL renders as `http://${git.HIVEPAAS_HOST}:${git.HIVEPAAS_PORT}`; the config file carries `capacity: 1`.

---

### Task 3: The index, the README, and the merge

**Files:**
- Modify: `README.md` (the supported blocks name `settings.dockerApi`)
- Regenerate: `index.json`

- [ ] **Step 1: README**

In the list of what `app` supports, after `settings.routing (...)`, add:

```markdown
  `settings.dockerApi` gives the app the Docker API through HivePaaS, without the Docker
  socket: `images` its containers may run (`"*"` for any), `sharedDirs` of its own storage
  they may bind, `networks` besides their own (`env`), `allow` for more than running
  containers (`exec`, `files`, `volumes`, `networks`, `nestedSocket`), and `limits`
  (`containers`, `memory`, `cpus`). The app reaches it at `${HIVEPAAS_DOCKER_HOST}`, and
  creating it needs Write on the Cluster module. The block is read from the template, so it
  takes no placeholders and a version cannot override it.
```

- [ ] **Step 2: Index, lint, commit, merge**

```bash
cd /Users/tnt/go/src/github.com/hivepaas/hivepaas
go run ./tools/apptemplate index ../app-templates
go run ./tools/apptemplate lint ../app-templates
cd ../app-templates
git add templates/autobase.yaml templates/gitea-runner.yaml icons/autobase.svg icons/gitea-runner.svg README.md index.json
git commit -m "feat: autobase and a Gitea Actions runner, the first two given the Docker API

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
git checkout main
git merge --no-ff feat/docker-api-templates -m "Merge branch 'feat/docker-api-templates'

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
git branch -d feat/docker-api-templates
```

Expected: `index.json` lists both templates with their hashes; lint is clean on `main`.
