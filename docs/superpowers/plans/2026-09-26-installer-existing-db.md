# Installer: Keep or Reset an Earlier Database Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** When the installer finds the database of an earlier HivePaaS without its services, it asks at once to **keep** or **reset** it, instead of deploying over it silently from `credentials.txt`.

**Architecture:** `detect_install_state` calls `choose_existing_db` (`HIVEPAAS_EXISTING_DB=keep|reset`, or typed, no default, not answered by `--yes`). `keep` takes the password (given, or `credentials.txt`) and needs `hivepaas.toml` (`take_kept_credentials`, in the questions), then `check_kept_db` tries the password in a throwaway `--network none` Postgres of the release's image on the volume before anything changes; any failure stops the install. `reset` deletes the database volumes and `db-volume.env` in the deploy step (`reset_db`). Both wait for the removed stack's containers to let go of the volume (`wait_volume_free`). `reinstall_over_database` goes.

**Tech Stack:** bash (tests under bash 3.2 and 5), Docker swarm, the release's Postgres image.

**Spec:** `docs/superpowers/specs/2026-09-25-release-installer-design.md` - Decision 3, §4 "What a run cannot guess", §10; amended in this plan's commit.

## Global Constraints

- No default for keep or reset; `--yes` never answers it; a silent install without `HIVEPAAS_EXISTING_DB` stops.
- `keep` never falls back to `reset`: a missing password, a missing `hivepaas.toml`, or a password that does not open the database stops the install, before anything changes.
- The password check runs with `--network none`, as `postgres`, with an `hba_file` that checks passwords on loopback (the image's own trusts loopback).
- `reset` deletes only `hivepaas_db` and `hivepaas_db_<major>` volumes, and `<app data>/system/update/db-volume.env`.
- `credentials.txt` is still written every run; it is read only on the `keep` path.
- Branch `feat/installer-existing-db`; merge into `main` locally with `--no-ff`; never push; commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Review Focus

- **`docker stack rm` a moment ago:** the old containers still hold the volume, and a second Postgres on the same data would be the worst outcome. `wait_volume_free` waits up to 120s; the e2e runs straight after the stack's containers stop running, which is the race.
- **A database of another Postgres major** than the release's image: `check_kept_db` uses `db-volume.env`'s major when there is one, else the image's; a mismatch reports that Postgres would not start, and stops.
- **A typo at the question** is asked again; only `keep` and `reset` go on.

## Before you start

The patches were cut from a tree where: the new tests failed before the code (`400 passed, 7 failed`) and passed after (`414 passed`; bash 5: `399 passed`), shellcheck was clean, and the e2e passed every check with the agent built from the checkout - among them a silent run stopping until told, a wrong password stopping `keep`, `keep` bringing the old admin back, and `reset` starting over with a new admin while the old password is refused.

---

### Task 1: Keep or reset an earlier database

**Files:**
- Modify: `deployment/release/install.sh`, `deployment/release/install.env`, `deployment/release/install_e2e.sh`
- Test: `deployment/release/install_test.sh`

**Interfaces:**
- Consumes: `ask`, `die`, `warn`, `info` (Output/Questions); `credentials_file`, `take_credential`, `secret_file`, `write_credentials`, `write_secret_file`, `resolve_images`, `STACK`.
- Produces: `EXISTING_DB` (`''|keep|reset`); `choose_existing_db`; `take_kept_credentials`; `db_volumes`; `DB_CHECK_SCRIPT`; `wait_volume_free VOLUME`; `check_kept_db`; `reset_db`. `resolve_images` now clears `HP_DB_VOLUME` and `HP_DB_MAJOR` first. `print_summary` says whether the earlier database is kept or deleted. `--help` and the template explain `HIVEPAAS_EXISTING_DB`.

- [ ] **Step 1: Branch**

```bash
git checkout main && git checkout -b feat/installer-existing-db
```

- [ ] **Step 2: Write the failing tests** - save as a file and `git apply` it:

```diff
diff --git a/deployment/release/install_test.sh b/deployment/release/install_test.sh
index 72a93439..71b0105c 100755
--- a/deployment/release/install_test.sh
+++ b/deployment/release/install_test.sh
@@ -629,6 +629,47 @@ test_credentials_file() {
   check "nothing else is taken" "$TMP/data" "$HIVEPAAS_DATA_DIR"
 }
 
+test_choose_existing_db() {
+  local out
+  HIVEPAAS_EXISTING_DB=reset
+  choose_existing_db
+  check "given: reset" reset "$EXISTING_DB"
+  out=$(HIVEPAAS_EXISTING_DB=maybe; (choose_existing_db) 2>&1)
+  check "given something else: stops" 1 "$?"
+  unset HIVEPAAS_EXISTING_DB
+  EXISTING_DB=''
+  out=$( (choose_existing_db) 2>&1)
+  check "no terminal and nothing given: stops" 1 "$?"
+  check_contains "saying what to set" "$out" "HIVEPAAS_EXISTING_DB"
+  ASSUME_YES=1
+  answers '' yes keep
+  choose_existing_db 2>/dev/null
+  check "asked until keep or reset, --yes or not" keep "$EXISTING_DB"
+}
+
+test_take_kept_credentials() {
+  local out
+  HIVEPAAS_DATA_DIR=$TMP/kept
+  mkdir -p "$HIVEPAAS_DATA_DIR"
+  out=$( (take_kept_credentials) 2>&1)
+  check "no password: stops" 1 "$?"
+  check_contains "saying where one comes from" "$out" "credentials.txt"
+  HIVEPAAS_DB_PASSWORD=db-pass HIVEPAAS_REDIS_PASSWORD=redis-pass HIVEPAAS_AGENT_TOKEN=agent-token
+  HIVEPAAS_JWT_SECRET=jwt-secret
+  write_credentials "$(credentials_file)"
+  unset HIVEPAAS_DB_PASSWORD HIVEPAAS_REDIS_PASSWORD HIVEPAAS_AGENT_TOKEN HIVEPAAS_JWT_SECRET
+  out=$( (take_kept_credentials) 2>&1)
+  check "no hivepaas.toml: stops" 1 "$?"
+  check_contains "naming it" "$out" "hivepaas.toml"
+  write_secret_file "$(secret_file)" 0123456789abcdef0123456789abcdef
+  take_kept_credentials
+  check "the password from credentials.txt" db-pass "$HIVEPAAS_DB_PASSWORD"
+  check "the JWT secret too, so sessions last" jwt-secret "$HIVEPAAS_JWT_SECRET"
+  HIVEPAAS_DB_PASSWORD=given-pass
+  take_kept_credentials
+  check "a given password wins" given-pass "$HIVEPAAS_DB_PASSWORD"
+}
+
 # --------------------------------------------------------------------- Host
 
 test_read_os_release() {
@@ -783,6 +824,7 @@ test_help() {
   check "--help exits 0" 0 "$?"
   check_contains "usage" "$out" "Usage: install.sh"
   check_contains "the settings file to fill in" "$out" "deployment/release/install.env"
+  check_contains "the choice over an earlier database" "$out" "HIVEPAAS_EXISTING_DB"
   for key in ADMIN_EMAIL ADMIN_PASSWORD APP_DOMAIN ROOT_DOMAIN APP_SECRET DATA_DIR PROJECT_DATA_DIR CHANNEL \
     SWAP SWAP_SIZE_MB EARLYOOM UPGRADE_DOCKER AGENT_IMAGE RELEASE_BRANCH INSTALL_REF; do
     check_contains "--help lists HIVEPAAS_$key" "$out" "HIVEPAAS_$key"
```

- [ ] **Step 3: Run them to see them fail**

Run: `bash deployment/release/install_test.sh`
Expected: FAIL - `400 passed, 7 failed` (`choose_existing_db`, `take_kept_credentials` not found; `--help` not naming `HIVEPAAS_EXISTING_DB`).

- [ ] **Step 4: Write the code** - `git apply`:

```diff
diff --git a/deployment/release/install.env b/deployment/release/install.env
index ce6b6118..d6ec6ff4 100644
--- a/deployment/release/install.env
+++ b/deployment/release/install.env
@@ -62,7 +62,9 @@ HIVEPAAS_APP_DOMAIN=
 # The agent's image. Default: the release's, or the app image's with -agent.
 #HIVEPAAS_AGENT_IMAGE=
 
-# Only to deploy again over a database whose services were removed, and only
-# when <data dir>/credentials.txt, which the installer reads for it, is gone
-# too: the password the database was created with.
+# When this server has the database of an earlier HivePaaS - after
+# `docker stack rm` - keep it (use it again) or reset it (delete everything in
+# it). A silent install stops without it. keep needs the database's password,
+# from <data dir>/credentials.txt or below, and <data dir>/hivepaas.toml.
+#HIVEPAAS_EXISTING_DB=keep
 #HIVEPAAS_DB_PASSWORD=
diff --git a/deployment/release/install.sh b/deployment/release/install.sh
index a552030c..5a6a1281 100755
--- a/deployment/release/install.sh
+++ b/deployment/release/install.sh
@@ -488,6 +488,8 @@ MISSING=''
 RELEASE_FILE=''
 # INSTALLED=1: HivePaaS runs here and has created its admin.
 INSTALLED=0
+# EXISTING_DB: keep or reset, for the database an earlier HivePaaS left here.
+EXISTING_DB=''
 
 # open_tty: questions are read from the terminal, not from stdin, which
 # `curl | bash` makes the script itself. fd 3 reads answers, fd 4 shows the
@@ -760,6 +762,58 @@ write_credentials() {
   ) && mv -f "$1.tmp" "$1"
 }
 
+# choose_existing_db: what to do with the database an earlier HivePaaS left on
+# this server - given in HIVEPAAS_EXISTING_DB, or asked. Nothing is assumed:
+# old data can be what the person wants back, or what broke the old install.
+# --yes does not answer it.
+choose_existing_db() {
+  local reply
+  if [ -n "${HIVEPAAS_EXISTING_DB:-}" ]; then
+    case "$HIVEPAAS_EXISTING_DB" in
+      keep | reset)
+        EXISTING_DB=$HIVEPAAS_EXISTING_DB
+        return 0
+        ;;
+    esac
+    die "HIVEPAAS_EXISTING_DB: '$HIVEPAAS_EXISTING_DB' is neither keep nor reset."
+  fi
+  if [ "$HAVE_TTY" != 1 ]; then
+    die "This server has the database of an earlier HivePaaS. Set HIVEPAAS_EXISTING_DB to keep, to use" \
+      "it again, or to reset, to delete it; then run the installer again."
+  fi
+  warn "This server has the database of an earlier HivePaaS."
+  info "  keep:  use it again, with its password and the hivepaas.toml of its app data"
+  info "  reset: delete everything in it, and install afresh"
+  while :; do
+    ask reply "Keep it or reset it? Type keep or reset: "
+    case "$reply" in
+      keep | reset)
+        EXISTING_DB=$reply
+        return 0
+        ;;
+    esac
+  done
+}
+
+# take_kept_credentials: what a kept database needs - its password, given or
+# from credentials.txt in the app data directory, and hivepaas.toml there, whose
+# secret is the only key to its encrypted data. Either missing stops the install.
+take_kept_credentials() {
+  local creds
+  creds=$(credentials_file)
+  if [ -z "${HIVEPAAS_DB_PASSWORD:-}" ] && [ -f "$creds" ]; then
+    read_kv_file "$creds" take_credential
+  fi
+  if [ -z "${HIVEPAAS_DB_PASSWORD:-}" ]; then
+    die "To keep the database, the installer needs its password: set HIVEPAAS_DB_PASSWORD, or put" \
+      "credentials.txt back in $HIVEPAAS_DATA_DIR - or run again and reset the database."
+  fi
+  if [ ! -f "$(secret_file)" ]; then
+    die "To keep the database, the installer needs $(secret_file): its secret is the only key to" \
+      "the encrypted data. Put the file back - or run again and reset the database."
+  fi
+}
+
 # take_credential KEY VALUE: a line of credentials.txt - the four keys it holds,
 # and nothing else.
 take_credential() {
@@ -808,6 +862,7 @@ ask_questions() {
     "Enter an absolute path outside the system's directories, and not the app data directory or one above it." \
     "$HIVEPAAS_DATA_DIR/project_data"
   normalize_settings
+  if [ "$EXISTING_DB" = keep ]; then take_kept_credentials; fi
   load_secret_file
   answer HIVEPAAS_APP_SECRET "App secret, which encrypts stored secrets" valid_app_secret \
     "The app secret needs 32 characters or more, and no spaces, quotes or backslashes." \
@@ -841,6 +896,10 @@ print_summary() {
   fi
   info "App domain     $HIVEPAAS_APP_DOMAIN"
   info "Root domain    $HIVEPAAS_ROOT_DOMAIN"
+  case "$EXISTING_DB" in
+    keep) info "Database       the earlier one, kept" ;;
+    reset) info "Database       the earlier one is DELETED, and a new one made" ;;
+  esac
   info "App secret     $(mask "$HIVEPAAS_APP_SECRET"), kept in $(secret_file)"
   info "App data       $HIVEPAAS_DATA_DIR"
   info "Project data   $HIVEPAAS_PROJECT_DATA_DIR"
@@ -1370,8 +1429,77 @@ service_exists() {
   docker service inspect "${STACK}_$1" >/dev/null 2>&1
 }
 
+# db_volumes: the volumes a HivePaaS database is on - hivepaas_db, and the
+# hivepaas_db_<major> a Postgres major upgrade moves it to.
+db_volumes() {
+  docker volume ls -q 2>/dev/null | grep -E "^${STACK}_db(_[0-9]+)?\$" || true
+}
+
 db_volume_exists() {
-  [ -n "$(docker volume ls -q --filter name=hivepaas_db 2>/dev/null)" ]
+  [ -n "$(db_volumes)" ]
+}
+
+# DB_CHECK_SCRIPT: run as postgres in a throwaway container of the database's
+# image, on its volume: start it on loopback with a password check there, try
+# the password, stop it. Exit 0: it opens the database; 3: Postgres would not
+# start on the volume; anything else: it does not.
+# shellcheck disable=SC2016 # expanded in the container
+DB_CHECK_SCRIPT='
+printf "host all all 127.0.0.1/32 scram-sha-256\n" >/tmp/hba.conf &&
+  pg_ctl -D "$PGDATA" -o "-c listen_addresses=127.0.0.1 -c hba_file=/tmp/hba.conf" -w -t 60 start >/dev/null 2>&1 ||
+  exit 3
+PGPASSWORD=$HP_DB_PASSWORD psql -h 127.0.0.1 -U hivepaas -d hivepaas -tAc "select 1" >/dev/null 2>&1
+status=$?
+pg_ctl -D "$PGDATA" -m fast -w stop >/dev/null 2>&1
+exit $status'
+
+# wait_volume_free VOLUME: until no container uses the volume. Right after
+# `docker stack rm`, swarm is still stopping and removing the old tasks: a
+# second Postgres on the same data, or a volume still in use, would follow.
+wait_volume_free() {
+  local start=$SECONDS
+  while [ -n "$(docker ps -aq --filter "volume=$1" 2>/dev/null)" ]; do
+    if [ $((SECONDS - start)) -ge 120 ]; then
+      die "The volume $1 is still in use: 'docker ps -a --filter volume=$1' says by what. Stop that," \
+        "then run the installer again."
+    fi
+    sleep 2
+  done
+}
+
+# check_kept_db: the password opens the kept database, tried before anything
+# changes, in a throwaway Postgres with no network. It fails, and the install
+# stops: a run again can keep it with the right password, or reset it.
+check_kept_db() {
+  local status=0
+  resolve_images
+  wait_volume_free "${HP_DB_VOLUME:-${STACK}_db}"
+  HP_DB_PASSWORD=$HIVEPAAS_DB_PASSWORD docker run --rm --network none --user postgres -e HP_DB_PASSWORD \
+    -e PGDATA="/var/lib/postgresql/$HP_DB_MAJOR/docker" -v "${HP_DB_VOLUME:-${STACK}_db}:/var/lib/postgresql" \
+    --entrypoint sh "$HIVEPAAS_IMAGE_DB" -c "$DB_CHECK_SCRIPT" >/dev/null 2>&1 || status=$?
+  case "$status" in
+    0) ok "The database's password opens it: it is kept." ;;
+    3)
+      die "Postgres $HP_DB_MAJOR ($HIVEPAAS_IMAGE_DB) would not start on the volume" \
+        "${HP_DB_VOLUME:-${STACK}_db}. Run the installer again and reset the database."
+      ;;
+    *)
+      die "The password does not open the database. Run the installer again with its password in" \
+        "HIVEPAAS_DB_PASSWORD, or reset the database."
+      ;;
+  esac
+}
+
+# reset_db: the earlier database deleted - every volume it was on, and the
+# record of the one a major upgrade moved it to.
+reset_db() {
+  local volume
+  for volume in $(db_volumes); do
+    wait_volume_free "$volume"
+    docker volume rm "$volume" >/dev/null || die "Could not delete the volume $volume; see above."
+    ok "Deleted the database volume $volume."
+  done
+  rm -f "$HIVEPAAS_DATA_DIR/system/update/db-volume.env"
 }
 
 # service_env SERVICE: the service's environment, one VAR=VALUE a line; empty
@@ -1395,34 +1523,13 @@ detect_install_state() {
   else
     INSTALL_STATE=fresh
     if db_volume_exists; then
-      reinstall_over_database
+      choose_existing_db
+      # A kept database has its admin.
+      if [ "$EXISTING_DB" = keep ]; then INSTALLED=1; fi
     fi
   fi
 }
 
-# reinstall_over_database: a database volume without services - left by
-# `docker stack rm` - is deployed over again, with the password it was created
-# with: given, or from credentials.txt in the app data directory. Its admin
-# exists already, so it is not asked for.
-reinstall_over_database() {
-  local creds
-  if [ -z "${HIVEPAAS_DB_PASSWORD:-}" ]; then
-    creds="$(normalize_dir "${HIVEPAAS_DATA_DIR:-/var/lib/hivepaas}")/credentials.txt"
-    if [ -f "$creds" ]; then
-      read_kv_file "$creds" take_credential
-      HIVEPAAS_DATA_DIR=${HIVEPAAS_DATA_DIR:-${creds%/*}}
-    fi
-  fi
-  if [ -z "${HIVEPAAS_DB_PASSWORD:-}" ]; then
-    die "A HivePaaS database volume is on this server, but no HivePaaS services and no credentials.txt" \
-      "in the app data directory to take its password from. Set HIVEPAAS_DB_PASSWORD (or HIVEPAAS_DATA_DIR," \
-      "where credentials.txt and hivepaas.toml are), or remove the volume to start over ('docker volume ls'," \
-      "'docker volume rm'), then run the installer again."
-  fi
-  INSTALLED=1
-  info "Deploying again over the database already here; its admin is kept."
-}
-
 # port_taken PORT: something on this host takes connections on PORT.
 port_taken() {
   (exec 3<>"/dev/tcp/127.0.0.1/$1") 2>/dev/null
@@ -1464,6 +1571,7 @@ take_db_volume() {
 
 resolve_images() {
   local app agent db_env="$HIVEPAAS_DATA_DIR/system/update/db-volume.env"
+  HP_DB_VOLUME='' HP_DB_MAJOR=
   app=$(release_field "$RELEASE_FILE" "$HIVEPAAS_CHANNEL" appImage)
   agent=${HIVEPAAS_AGENT_IMAGE:-$(release_field "$RELEASE_FILE" "$HIVEPAAS_CHANNEL" agentImage)}
   if [ -z "$agent" ]; then
@@ -1581,6 +1689,7 @@ redeploy_stack() {
 deploy() {
   case "$INSTALL_STATE" in
     fresh)
+      if [ "$EXISTING_DB" = reset ]; then reset_db; fi
       resolve_images
       # Over a database already here, the admin exists: no first boot.
       if [ "$INSTALLED" = 1 ]; then deploy_stack 0; else deploy_stack 1; fi
@@ -1792,8 +1901,10 @@ Settings, from the environment or --config (the environment wins):
   HIVEPAAS_AGENT_IMAGE         the agent's image (default: from the release)
   HIVEPAAS_RELEASE_BRANCH      the branch the release info is read from (default: release)
   HIVEPAAS_INSTALL_REF         the ref the stack files are downloaded from (default: main)
-  HIVEPAAS_DB_PASSWORD         only to redeploy over a database whose services are gone,
-                               when <data dir>/credentials.txt, read by default, is gone too
+  HIVEPAAS_EXISTING_DB         keep or reset, when this server has the database of an
+                               earlier HivePaaS; asked when not set, and never assumed
+  HIVEPAAS_DB_PASSWORD         with keep: the database's password, when
+                               <data dir>/credentials.txt is gone
 
 Silent install - a settings file to fill in, with every setting explained:
   curl -fsSLO https://raw.githubusercontent.com/hivepaas/hivepaas/main/deployment/release/install.env
@@ -1871,6 +1982,7 @@ main() {
   generate_secrets
   gather_addresses
   if [ "$INSTALL_STATE" = fresh ] || [ "$REDEPLOY" = 1 ]; then fetch_release; fi
+  if [ "$EXISTING_DB" = keep ]; then check_kept_db; fi
   if [ "$INSTALL_STATE" = installed ]; then
     ok "HivePaaS is installed here; its settings are read from its services."
     if [ "$REDEPLOY" = 1 ]; then
```

- [ ] **Step 5: Run the tests, under both shells**

Run: `make test-installer`
Expected: no shellcheck output, then `414 passed, 0 failed`.

Run: `docker run --rm -v "$PWD":/repo -w /repo/deployment/release bash:5.2 bash -c 'apk add -q jq openssl >/dev/null 2>&1; bash install_test.sh'`
Expected: three `skip` lines, then `399 passed, 0 failed`.

- [ ] **Step 6: The e2e** - `git apply`:

```diff
diff --git a/deployment/release/install_e2e.sh b/deployment/release/install_e2e.sh
index 0932aa8c..829cf9f0 100755
--- a/deployment/release/install_e2e.sh
+++ b/deployment/release/install_e2e.sh
@@ -6,8 +6,9 @@
 # signing in, the admin's password gone, the app secret in hivepaas.toml and
 # in no service - and the runs after it: one that must change nothing, one
 # with another app secret and one without hivepaas.toml that must stop, one
-# after an address changed, a --redeploy, and one after `docker stack rm`,
-# which must deploy again over the database left behind, from credentials.txt.
+# after an address changed, a --redeploy, and after `docker stack rm` the
+# database left behind: a silent run stops until told to keep or reset it, a
+# wrong password stops keeping it, the right one keeps it, and reset starts over.
 #
 #   make test-installer-e2e     (bash deployment/release/install_e2e.sh)
 #
@@ -196,24 +197,44 @@ expect "and every system service has OOM priority -500 again" "-500 -500 -500 -5
 expect "and the dashboard answers" 200 "$(in_dind "curl -sk -o /dev/null -w '%{http_code}' \
   --resolve $DOMAIN:443:127.0.0.1 https://$DOMAIN/_/ping")"
 
-in_dind 'docker stack rm hivepaas >/dev/null
-  for _ in $(seq 120); do
-    if ! docker network inspect hivepaas_local_net >/dev/null 2>&1 &&
-      [ -z "$(docker ps -q --filter label=com.docker.stack.namespace=hivepaas)" ]; then break; fi
-    sleep 1
-  done'
-in_dind 'mv /var/lib/hivepaas/credentials.txt /root/credentials.txt.bak'
-if in_dind "$INSTALL >/root/install6.log 2>&1"; then fail "a run over a database without its password went on"; fi
-in_dind 'grep -q "no credentials.txt" /root/install6.log' || fail "a run over a database without its password: no reason"
-pass "after docker stack rm, a run without credentials.txt stops"
-in_dind 'mv /root/credentials.txt.bak /var/lib/hivepaas/credentials.txt'
-# The services held the domain too: a run over the database is told it again,
-# as a person would answer it.
-in_dind "HIVEPAAS_APP_DOMAIN=$DOMAIN $INSTALL >/root/install7.log 2>&1" || {
-  in_dind 'tail -n 30 /root/install7.log' >&2
-  fail "a run over the database left behind"
+# stack_rm: the stack removed, as a person would, and waited for; the database
+# volume and the app data stay.
+stack_rm() {
+  in_dind 'docker stack rm hivepaas >/dev/null
+    for _ in $(seq 120); do
+      if ! docker network inspect hivepaas_local_net >/dev/null 2>&1 &&
+        [ -z "$(docker ps -q --filter label=com.docker.stack.namespace=hivepaas)" ]; then break; fi
+      sleep 1
+    done'
 }
-pass "with it, a run deploys again over the database left behind"
+
+stack_rm
+if in_dind "HIVEPAAS_APP_DOMAIN=$DOMAIN $INSTALL >/root/install6.log 2>&1"; then
+  fail "a run over an earlier database, not told what to do with it, went on"
+fi
+in_dind 'grep -q HIVEPAAS_EXISTING_DB /root/install6.log' || fail "a run over an earlier database: no reason given"
+pass "after docker stack rm, a silent run stops until told to keep or reset the database"
+if in_dind "HIVEPAAS_EXISTING_DB=keep HIVEPAAS_DB_PASSWORD=wrong-password HIVEPAAS_APP_DOMAIN=$DOMAIN \
+  $INSTALL >/root/install7.log 2>&1"; then
+  fail "keeping the database with a wrong password went on"
+fi
+in_dind 'grep -q "password does not open the database" /root/install7.log' || fail "a wrong password: no reason given"
+pass "keeping it with a wrong password stops"
+in_dind "HIVEPAAS_EXISTING_DB=keep HIVEPAAS_APP_DOMAIN=$DOMAIN $INSTALL >/root/install8.log 2>&1" || {
+  in_dind 'tail -n 30 /root/install8.log' >&2
+  fail "keeping the database, its password from credentials.txt"
+}
+pass "keeping it, with the password from credentials.txt, deploys again over it"
 expect "where the admin still signs in" 200 "$(login "$DOMAIN" "$PASSWORD")"
 
+stack_rm
+in_dind "HIVEPAAS_EXISTING_DB=reset HIVEPAAS_ADMIN_EMAIL=new@example.com HIVEPAAS_ADMIN_PASSWORD=new-password-456 \
+  HIVEPAAS_APP_DOMAIN=$DOMAIN $INSTALL >/root/install9.log 2>&1" || {
+  in_dind 'tail -n 30 /root/install9.log' >&2
+  fail "resetting the database"
+}
+pass "resetting it installs afresh"
+expect "where the new admin signs in" 200 "$(login "$DOMAIN" new-password-456)"
+expect "and the old one does not" 401 "$(login "$DOMAIN" "$PASSWORD")"
+
 printf 'all end-to-end checks passed\n'
```

Build the agent from the checkout and run the e2e as in `docs/superpowers/plans/2026-09-26-installer-secret-file.md`, Task 3 Steps 2-3. Expected: every check `ok`, the last ones `after docker stack rm, a silent run stops until told to keep or reset the database`, `keeping it with a wrong password stops`, `keeping it, with the password from credentials.txt, deploys again over it`, `where the admin still signs in`, `resetting it installs afresh`, `where the new admin signs in`, `and the old one does not`, then `all end-to-end checks passed`.

- [ ] **Step 7: Commit, review, merge**

```bash
git add deployment/release
git commit -m "feat(installer): keep or reset the database an earlier HivePaaS left" -m "A database volume without HivePaaS services is no longer deployed over silently. The installer asks at once - keep or reset, typed, no default, HIVEPAAS_EXISTING_DB for a silent install - since old data can be what broke the old install. keep takes the password, given or from credentials.txt, needs hivepaas.toml, and tries the password in a throwaway Postgres before anything changes; any failure stops. reset deletes the database volumes and installs afresh. Both wait for the removed stack to let go of the volume." -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
git checkout main
git merge --no-ff feat/installer-existing-db -m "Merge branch 'feat/installer-existing-db'" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
make test-installer
git branch -d feat/installer-existing-db
```

Expected: `414 passed, 0 failed` after the merge. Do not push.
