# Installer: the Secret in hivepaas.toml Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** The release installer keeps nothing of its own: the app secret goes to `<app data>/hivepaas.toml`, every other setting stays in the services' environment and is read back from there, a commented `deployment/release/install.env` template drives silent installs, and the agent no longer requires the app or JWT secret.

**Architecture:** `install.sh` stops writing `/etc/hivepaas/install.env`. Its settings come from the environment, then `--config`, then the HivePaaS already on the server - the `hivepaas_app` service's environment (`take_installed_env`) and the secret in the app's managed settings file (`load_secret_file`). The installer writes that file once (`write_secret_file`, with the security settings commented and explained); the stack gives no service the app secret, and the agent neither it nor the JWT secret, which the app now lets the agent do without. The stack files are fetched into the run's temporary directory.

**Tech Stack:** bash (tests under bash 3.2 and 5), Docker swarm, Go (the app's `config` and `cmd/internal` packages, testify).

**Spec:** `docs/superpowers/specs/2026-09-25-release-installer-design.md` - Decisions 4-5, §1, §2 steps 3, 6 and 9, §4, §5, §6, §9, §10, §11 and "Not in this design" were amended in this plan's commit.

## Global Constraints

- No file of the installer's own on the server: nothing under `/etc/hivepaas`; the stack files live in the run's `mktemp -d` directory.
- The app secret: in `<app data>/hivepaas.toml`, mode `0600` (the app refuses anything wider), as `secret = "<value>"`; in no service's environment. Given values: 32 characters or more, no spaces, quotes or backslashes (so the TOML needs no escaping).
- Precedence: the environment, then `--config`, then the installation (the `hivepaas_app` environment and `hivepaas.toml`); the fixed settings stop the install when given with another value.
- The agent gets neither `HP_APP_SECRET` nor `HP_SESSION_JWT_SECRET`; app, worker and updater get the JWT secret.
- The template `deployment/release/install.env`: the three required settings uncommented and empty, every other setting of `--help` commented out, each explained with its default; it must read with `read_kv_file` without a warning.
- `golangci-lint run ./...` and `go test ./...` on the whole repo for the Go change (CLAUDE.md).
- The local Docker Desktop is the user's own swarm: nothing deploys to it. An image built for the e2e is tagged only long enough to `docker save` it, then untagged.
- Branch `feat/installer-secret-file`; merge into `main` locally with `--no-ff`; never push; commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Review Focus

- **The stack removed, its database volume left.** The database password was only in the services, so the installer must stop - unless `HIVEPAAS_DB_PASSWORD` is given - rather than create services with a new password the database refuses. Needs Docker; checked by reading `detect_install_state`, not by a test.
- **A secret the app rotated**, written by its TOML encoder, must still be read: pinned by `test_toml_secret` (Task 2), which feeds the form the app writes.
- **A setting changed with `docker service update --env-add`** is what the next run reads, and a `--redeploy` keeps: pinned by `test_take_installed_env` and `test_load_settings_precedence` (Task 2).
- **The released images predate the agent change:** their agent exits without the secrets. The e2e loads an agent built from the checkout (Task 3); the user must release new images before a server install works.
- **The dashboard rewrites `hivepaas.toml`** without its comments when a security setting is saved; nothing must depend on the comments. `toml_secret` reads only the `secret` line.

## Before you start

- Repository `/Users/tnt/go/src/github.com/hivepaas/hivepaas`, on `main`, clean. Every command runs from there.
- Each code step is a patch to apply with `git apply`. The patches were cut from a tree where every step below was run: the Go tests fail before Task 1's code and pass after; the installer's tests go from `317 passed, 45 failed` to `383 passed, 0 failed` (bash 3.2) and `368 passed` (bash 5, three skips); shellcheck is clean; and the e2e passes every check with the agent built from the checkout. Applying all five patches to `main` reproduces that tree byte for byte.

---

### Task 1: The agent needs neither secret

**Files:**
- Modify: `hivepaas_app/config/config.go` (`ensureAppSecret`)
- Modify: `hivepaas_app/cmd/internal/config.go` (`validateConfig`)
- Test: `hivepaas_app/config/managed_test.go`, `hivepaas_app/cmd/internal/config_test.go` (new)

**Interfaces:**
- Consumes: `config.RunModeAgent`, `config.EnvBeta`, `ErrAppSecretUnset`, `ErrInvalidConfig`.
- Produces: in `run_mode = agent`, `ensureAppSecret` accepts an empty secret and `validateConfig` checks neither the JWT nor the app secret. Task 2's stack relies on it.

- [ ] **Step 1: Branch**

```bash
git checkout main && git checkout -b feat/installer-secret-file
```

- [ ] **Step 2: Write the failing tests**

Save as `/tmp/installer-1a.diff` and run `git apply /tmp/installer-1a.diff`:

```diff
diff --git a/hivepaas_app/config/managed_test.go b/hivepaas_app/config/managed_test.go
index da8e03a7..1a4f23c6 100644
--- a/hivepaas_app/config/managed_test.go
+++ b/hivepaas_app/config/managed_test.go
@@ -190,6 +190,14 @@ func TestEnsureAppSecret(t *testing.T) {
 		assert.ErrorIs(t, err, ErrAppSecretUnset)
 		assert.Empty(t, config.Secret)
 	})
+
+	// The agent decrypts nothing, and runs on nodes where the app's volume, and
+	// the managed settings in it, are not.
+	t.Run("the agent needs none", func(t *testing.T) {
+		config := &Config{Env: EnvProd, RunMode: RunModeAgent}
+		assert.NoError(t, ensureAppSecret(config, t.TempDir()))
+		assert.Empty(t, config.Secret)
+	})
 }
 
 func TestSaveManagedSettings(t *testing.T) {
diff --git a/hivepaas_app/cmd/internal/config_test.go b/hivepaas_app/cmd/internal/config_test.go
new file mode 100644
index 00000000..bd35559d
--- /dev/null
+++ b/hivepaas_app/cmd/internal/config_test.go
@@ -0,0 +1,31 @@
+package internal
+
+import (
+	"testing"
+
+	"github.com/stretchr/testify/assert"
+
+	"github.com/hivepaas/hivepaas/hivepaas_app/config"
+	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
+)
+
+func TestValidateConfig(t *testing.T) {
+	logger := logging.GlobalLogger()
+
+	t.Run("outside development the app needs both secrets", func(t *testing.T) {
+		cfg := &config.Config{Env: config.EnvBeta, RunMode: config.RunModeApp}
+		assert.ErrorIs(t, validateConfig(cfg, logger), ErrInvalidConfig)
+
+		cfg.Session.JWTSecret = "0123456789abcdef0123456789abcdef"
+		assert.ErrorIs(t, validateConfig(cfg, logger), ErrInvalidConfig)
+
+		cfg.Secret = "0123456789abcdef"
+		assert.NoError(t, validateConfig(cfg, logger))
+	})
+
+	// The agent issues no sessions and decrypts nothing, so it is given neither.
+	t.Run("the agent needs neither", func(t *testing.T) {
+		cfg := &config.Config{Env: config.EnvBeta, RunMode: config.RunModeAgent}
+		assert.NoError(t, validateConfig(cfg, logger))
+	})
+}
```

- [ ] **Step 3: Run them to see them fail**

Run: `go test ./hivepaas_app/config/ ./hivepaas_app/cmd/internal/`
Expected: FAIL - `TestEnsureAppSecret/the_agent_needs_none` (`app secret is not configured`) and `TestValidateConfig/the_agent_needs_neither` (`JWT secret must be at least 32 characters`).

- [ ] **Step 4: Write the code**

Save as `/tmp/installer-1b.diff` and run `git apply /tmp/installer-1b.diff`:

```diff
diff --git a/hivepaas_app/cmd/internal/config.go b/hivepaas_app/cmd/internal/config.go
index 1fed5c00..796e06aa 100644
--- a/hivepaas_app/cmd/internal/config.go
+++ b/hivepaas_app/cmd/internal/config.go
@@ -83,16 +83,19 @@ func reloadConfigOnSignal(logger logging.Logger) {
 func validateConfig(cfg *config.Config, logger logging.Logger) error {
 	logger.Info("validating app config...")
 	isProdEnv := !cfg.IsDevEnv()
+	// The agent issues no sessions and decrypts nothing, so it is given neither
+	// secret; see config.ensureAppSecret.
+	needsSecrets := isProdEnv && cfg.RunMode != config.RunModeAgent
 
 	// JWT secret must not be empty or short enough to brute force offline
-	if isProdEnv && len(cfg.Session.JWTSecret) < jwtSecretMinLen {
+	if needsSecrets && len(cfg.Session.JWTSecret) < jwtSecretMinLen {
 		return fmt.Errorf("%w: JWT secret must be at least %d characters for production",
 			ErrInvalidConfig, jwtSecretMinLen)
 	}
 
 	// App secret is mandatory: everything stored as a secret is encrypted with it,
 	// and config.ensureAppSecret already refuses to start without one outside dev.
-	if isProdEnv && len(cfg.Secret) < appSecretMinLen {
+	if needsSecrets && len(cfg.Secret) < appSecretMinLen {
 		return fmt.Errorf("%w: app secret must be at least %d characters for production",
 			ErrInvalidConfig, appSecretMinLen)
 	}
diff --git a/hivepaas_app/config/config.go b/hivepaas_app/config/config.go
index 812cbee5..3686c41b 100644
--- a/hivepaas_app/config/config.go
+++ b/hivepaas_app/config/config.go
@@ -243,6 +243,11 @@ func ensureAppSecret(config *Config, appPath string) error {
 	if config.Secret != "" {
 		return nil
 	}
+	// The agent decrypts nothing, and runs on every node - where the app's volume,
+	// and the managed settings in it, are not. It is given no secret.
+	if config.RunMode == RunModeAgent {
+		return nil
+	}
 
 	if !config.IsDevEnv() {
 		return fmt.Errorf("%w: HP_APP_SECRET must be set, e.g. %s",
```

- [ ] **Step 5: Run the checks**

Run: `go build ./... && golangci-lint run ./... && go test ./...`
Expected: `0 issues.`, and every package `ok`.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/config hivepaas_app/cmd/internal
git commit -m "fix(config): the agent needs neither the app secret nor the JWT secret" -m "The agent decrypts nothing and issues no sessions, and runs on every node, where the app volume and hivepaas.toml are not. Outside development it was refused a start without both secrets, so the installer had to hand every node the key to all encrypted data." -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 2: The secret in hivepaas.toml, the settings from the services, and the template

**Files:**
- Modify: `deployment/release/install.sh`
- Modify: `deployment/release/hivepaas.yaml`, `deployment/release/hivepaas.first-boot.yaml` (a comment)
- Create: `deployment/release/install.env`
- Test: `deployment/release/install_test.sh`

**Interfaces:**
- Consumes: Task 1's agent behaviour (the stack gives the agent neither secret).
- Produces, in `install.sh`:
  - removed: `INSTALL_ENV`, `SAVED_KEYS`, `take_saved`, `save_settings`, `CONFIG_DIR`, `HIVEPAAS_INSTALLED`, the test variable `HIVEPAAS_INSTALL_ENV_FILE`.
  - `INSTALLED` (0/1 global); `take_installed KEY VALUE`; `channel_of_app_env ENV`; `take_installed_env ENV_TEXT`; `secret_file` (`$HIVEPAAS_DATA_DIR/hivepaas.toml`); `toml_secret FILE`; `write_secret_file FILE SECRET`; `load_secret_file`; `load_settings INSTALLED_ENV`; `service_env SERVICE`; `silent_hint`.
  - `ask_questions` asks the app secret last, after the data directory, and not when `hivepaas.toml` has one; `valid_app_secret` also refuses quotes and backslashes.
  - `detect_install_state` stops on HivePaaS without `hivepaas.toml`, and on a database volume without services unless `HIVEPAAS_DB_PASSWORD` is given.
  - `prepare_files` writes `hivepaas.toml` when missing and fetches the stack files into `$WORK_DIR`; `deploy_stack` deploys from there and exports no app secret.
- Produces, in `hivepaas.yaml`: `HP_APP_SECRET` gone; `HP_SESSION_JWT_SECRET` on app, worker and updater only.

- [ ] **Step 1: Write the failing tests**

Save as `/tmp/installer-2a.diff` and run `git apply /tmp/installer-2a.diff`:

```diff
diff --git a/deployment/release/install_test.sh b/deployment/release/install_test.sh
index 2e4f8abb..287effb7 100755
--- a/deployment/release/install_test.sh
+++ b/deployment/release/install_test.sh
@@ -9,7 +9,7 @@
 #
 # The tests feed install.sh literal $, `...` and $(...) on purpose, and set
 # the settings it reads:
-# shellcheck disable=SC2016,SC2031,SC2153
+# shellcheck disable=SC2016,SC2030,SC2031,SC2153
 
 set -u
 
@@ -168,6 +168,9 @@ test_valid_app_secret() {
   check_ok "64 hex" valid_app_secret "$(openssl rand -hex 32)"
   check_fails "31 characters" valid_app_secret 0123456789abcdef0123456789abcde
   check_fails "a space" valid_app_secret "0123456789abcdef 0123456789abcdef"
+  check_fails "a double quote" valid_app_secret '0123456789abcdef"0123456789abcdef'
+  check_fails "a single quote" valid_app_secret "0123456789abcdef'0123456789abcdef"
+  check_fails "a backslash" valid_app_secret '0123456789abcdef\0123456789abcdef'
 }
 
 test_valid_ipv4() {
@@ -378,7 +381,7 @@ test_addresses_line() {
 # stack_env: what install.sh exports for docker stack deploy, with sample values.
 stack_env() {
   export HIVEPAAS_APP_ENV=beta HIVEPAAS_ROOT_DOMAIN=mydomain.com HIVEPAAS_APP_DOMAIN=hivepaas.dev.mydomain.com \
-    HIVEPAAS_APP_SECRET=s3cret HIVEPAAS_JWT_SECRET=jwt HIVEPAAS_DATA_DIR=/var/lib/hivepaas \
+    HIVEPAAS_JWT_SECRET=jwt HIVEPAAS_DATA_DIR=/var/lib/hivepaas \
     HIVEPAAS_PROJECT_DATA_DIR=/data/projects HIVEPAAS_DB_PASSWORD=dbpw HIVEPAAS_REDIS_PASSWORD=redispw \
     HIVEPAAS_AGENT_TOKEN=agenttok HIVEPAAS_IP_RULE='Host(`1.2.3.4`) || Host(`10.0.0.5`)' \
     HIVEPAAS_ADMIN_EMAIL=you@example.com HIVEPAAS_ADMIN_PASSWORD='p$ss "w0rd"' HP_DB_MAJOR=18 \
@@ -401,17 +404,19 @@ test_stack_renders() {
     'traefik.http.routers.x-custom-router-ip.rule: Host(`1.2.3.4`) || Host(`10.0.0.5`)'
   check_contains "the Postgres major" "$out" 'PGDATA: /var/lib/postgresql/18/docker'
   check_contains "project data seen from the app" "$out" 'target: /host/data/projects'
-  check "every HivePaaS process gets the secret" 5 "$(printf '%s\n' "$out" | grep -c 'HP_APP_SECRET: s3cret')"
+  check "no service is given the app secret" 0 "$(printf '%s\n' "$out" | grep -c 'HP_APP_SECRET')"
+  check "app, worker and updater get the JWT secret, the agent not" 3 \
+    "$(printf '%s\n' "$out" | grep -c 'HP_SESSION_JWT_SECRET: jwt')"
   check_lacks "no admin without the first-boot file" "$out" HP_USER_ADMIN_PASSWORD
   out=$(cd "$HERE" && docker stack config -c hivepaas.yaml -c hivepaas.first-boot.yaml 2>&1)
   check "renders with the first-boot file" 0 "$?"
   check "the admin, literally, on app and worker" 2 \
     "$(printf '%s\n' "$out" | grep -c 'HP_USER_ADMIN_PASSWORD: p$ss "w0rd"')"
-  unset HIVEPAAS_APP_SECRET
+  unset HIVEPAAS_JWT_SECRET
   out=$(cd "$HERE" && docker stack config -c hivepaas.yaml 2>&1)
   status=$?
   check_fails "a missing setting fails" test "$status" -eq 0
-  check_contains "and is named" "$out" HIVEPAAS_APP_SECRET
+  check_contains "and is named" "$out" HIVEPAAS_JWT_SECRET
 }
 
 # ---------------------------------------------------------------- Questions
@@ -425,7 +430,7 @@ answers() {
 
 test_ask_questions_interactive() {
   answers not-an-email you@example.com short password123 password124 password123 password123 \
-    HivePaaS.Dev.MyDomain.com '' '' /var/lib/hivepaas/ ''
+    HivePaaS.Dev.MyDomain.com '' /var/lib/hivepaas/ '' ''
   ask_questions 2>/dev/null
   check "email, asked again" you@example.com "$HIVEPAAS_ADMIN_EMAIL"
   check "password, typed twice alike" password123 "$HIVEPAAS_ADMIN_PASSWORD"
@@ -470,58 +475,106 @@ test_ask_questions_given_invalid() {
 }
 
 test_ask_questions_installed() {
-  HIVEPAAS_INSTALLED=true HIVEPAAS_APP_DOMAIN=app.example.com HIVEPAAS_APP_SECRET=0123456789abcdef0123456789abcdef
+  INSTALLED=1 HIVEPAAS_APP_DOMAIN=app.example.com HIVEPAAS_APP_SECRET=0123456789abcdef0123456789abcdef
   ask_questions
   check "no admin asked for once installed" "" "${HIVEPAAS_ADMIN_PASSWORD:-}"
 }
 
 test_load_settings_precedence() {
   printf '%s\n' HIVEPAAS_ADMIN_EMAIL=file@example.com HIVEPAAS_ADMIN_PASSWORD=file-password >"$TMP/conf"
-  printf '%s\n' "HIVEPAAS_ADMIN_PASSWORD='saved-password'" "HIVEPAAS_DATA_DIR='/srv/hivepaas'" >"$TMP/install.env"
-  CONFIG_FILE=$TMP/conf INSTALL_ENV=$TMP/install.env HIVEPAAS_ADMIN_EMAIL=env@example.com
-  load_settings
+  CONFIG_FILE=$TMP/conf HIVEPAAS_ADMIN_EMAIL=env@example.com
+  load_settings "HP_USER_ADMIN_PASSWORD=installed-password
+HP_STORAGE_HOST_DIR=/srv/hivepaas"
   check "the environment wins" env@example.com "$HIVEPAAS_ADMIN_EMAIL"
-  check "the file wins over install.env" file-password "$HIVEPAAS_ADMIN_PASSWORD"
-  check "install.env fills the rest" /srv/hivepaas "$HIVEPAAS_DATA_DIR"
+  check "the file wins over the installation" file-password "$HIVEPAAS_ADMIN_PASSWORD"
+  check "the installation fills the rest" /srv/hivepaas "$HIVEPAAS_DATA_DIR"
   check "the channel defaults to beta" beta "$HIVEPAAS_CHANNEL"
 }
 
 test_load_settings_fixed() {
-  local out saved=0123456789abcdef0123456789abcdef
-  printf '%s\n' "HIVEPAAS_APP_SECRET='$saved'" "HIVEPAAS_DATA_DIR='/srv/hivepaas'" >"$TMP/install.env"
-  INSTALL_ENV=$TMP/install.env
-  out=$(HIVEPAAS_APP_SECRET=${saved}X; (load_settings) 2>&1)
-  check "a different app secret stops" 1 "$?"
-  check_contains "naming it" "$out" HIVEPAAS_APP_SECRET
-  check_lacks "without printing it" "$out" "$saved"
-  out=$(HIVEPAAS_APP_SECRET=$saved HIVEPAAS_DATA_DIR=/srv//hivepaas/; (load_settings) 2>&1)
+  local out installed='HP_ENV=production
+HP_STORAGE_HOST_DIR=/srv/hivepaas
+HP_DB_PASSWORD=db-password-1'
+  out=$(HIVEPAAS_DB_PASSWORD=db-password-2; (load_settings "$installed") 2>&1)
+  check "a different database password stops" 1 "$?"
+  check_contains "naming it" "$out" HIVEPAAS_DB_PASSWORD
+  check_lacks "without printing it" "$out" db-password
+  out=$(HIVEPAAS_DATA_DIR=/srv//hivepaas/; (load_settings "$installed") 2>&1)
   check "the same values, written differently, go on" 0 "$?"
-  out=$(HIVEPAAS_CHANNEL=nightly; (load_settings) 2>&1)
+  out=$(HIVEPAAS_CHANNEL=beta; (load_settings "$installed") 2>&1)
+  check "another channel stops" 1 "$?"
+  load_settings "$installed"
+  check "the installation's channel" stable "$HIVEPAAS_CHANNEL"
+  out=$(HIVEPAAS_CHANNEL=nightly; (load_settings '') 2>&1)
   check "an unknown channel stops" 1 "$?"
 }
 
+test_take_installed_env() {
+  take_installed_env 'HP_ENV=beta
+HP_APP_DOMAIN=app.example.com
+HP_ROOT_DOMAIN=example.com
+HP_STORAGE_HOST_DIR=/srv/hivepaas
+HP_STORAGE_PROJECT_DATA_HOST_DIR=/data/projects
+HP_SESSION_JWT_SECRET=jwt=with=equals
+HP_DB_PASSWORD=db-password
+HP_CACHE_URL=redis://default:redis-password@redis:6379/0
+HP_AGENT_SECRET_TOKEN=agent-token
+HP_RUN_MODE=app+worker'
+  check "channel" beta "$HIVEPAAS_CHANNEL"
+  check "app domain" app.example.com "$HIVEPAAS_APP_DOMAIN"
+  check "root domain" example.com "$HIVEPAAS_ROOT_DOMAIN"
+  check "data dir" /srv/hivepaas "$HIVEPAAS_DATA_DIR"
+  check "project data dir" /data/projects "$HIVEPAAS_PROJECT_DATA_DIR"
+  check "a value with =" jwt=with=equals "$HIVEPAAS_JWT_SECRET"
+  check "database password" db-password "$HIVEPAAS_DB_PASSWORD"
+  check "redis password, from the URL" redis-password "$HIVEPAAS_REDIS_PASSWORD"
+  check "agent token" agent-token "$HIVEPAAS_AGENT_TOKEN"
+  check "installed" 1 "$INSTALLED"
+  take_installed_env 'HP_USER_ADMIN_EMAIL=you@example.com
+HP_USER_ADMIN_PASSWORD=password123'
+  check "the admin's password there: not finished" 0 "$INSTALLED"
+  check "the admin's password" password123 "$HIVEPAAS_ADMIN_PASSWORD"
+}
+
+test_toml_secret() {
+  printf '%s\n' '# rotated' 'secret = "from-the-app"' '[security]' 'secret = "not-this"' >"$TMP/a.toml"
+  check "a basic string" from-the-app "$(toml_secret "$TMP/a.toml")"
+  printf '%s\n' "  secret='literal'  # a comment" >"$TMP/b.toml"
+  check "a literal string" literal "$(toml_secret "$TMP/b.toml")"
+  printf '%s\n' '[security]' 'allow_privileged_apps = true' >"$TMP/c.toml"
+  check "none" "" "$(toml_secret "$TMP/c.toml")"
+}
+
+test_secret_file() {
+  local out saved=0123456789abcdef0123456789abcdef
+  HIVEPAAS_DATA_DIR=$TMP/data
+  mkdir -p "$HIVEPAAS_DATA_DIR"
+  write_secret_file "$(secret_file)" "$saved"
+  check "written" 0 "$?"
+  check "readable by root alone" 600 "$(stat -c %a "$(secret_file)" 2>/dev/null || stat -f %Lp "$(secret_file)")"
+  check "the secret reads back" "$saved" "$(toml_secret "$(secret_file)")"
+  check_contains "the security settings, explained" "$(cat "$(secret_file)")" "# allow_privileged_apps = false"
+  load_secret_file
+  check "it is the app secret" "$saved" "$HIVEPAAS_APP_SECRET"
+  out=$(HIVEPAAS_APP_SECRET=${saved}X; (load_secret_file) 2>&1)
+  check "another given secret stops" 1 "$?"
+  check_lacks "without printing it" "$out" "$saved"
+  printf '[security]\n' >"$(secret_file)"
+  out=$( (load_secret_file) 2>&1)
+  check "a file without a secret stops" 1 "$?"
+}
+
 test_load_settings_runs_nothing() {
   printf '%s\n' 'HIVEPAAS_X[$(touch '"$TMP"'/ran)]=1' 'PATH=/nowhere' 'HIVEPAAS_Y=$(touch '"$TMP"'/ran2)' >"$TMP/conf"
-  CONFIG_FILE=$TMP/conf INSTALL_ENV=$TMP/none.env
-  load_settings 2>/dev/null
+  CONFIG_FILE=$TMP/conf
+  load_settings '' 2>/dev/null
   check_fails "a key is not evaluated" test -e "$TMP/ran"
   check_fails "a value is not evaluated" test -e "$TMP/ran2"
   check "a value is taken literally" '$(touch '"$TMP"'/ran2)' "$HIVEPAAS_Y"
   check_fails "only HIVEPAAS_ settings are taken" test "$PATH" = /nowhere
 }
 
-test_save_settings() {
-  INSTALL_ENV=$TMP/etc/hivepaas/install.env
-  HIVEPAAS_ADMIN_PASSWORD="it's \$x \"quoted\"" HIVEPAAS_APP_DOMAIN=app.example.com HIVEPAAS_INSTALLED=''
-  save_settings
-  check "file mode" 600 "$(stat -c %a "$INSTALL_ENV" 2>/dev/null || stat -f %Lp "$INSTALL_ENV")"
-  check "dir mode" 700 "$(stat -c %a "${INSTALL_ENV%/*}" 2>/dev/null || stat -f %Lp "${INSTALL_ENV%/*}")"
-  check_lacks "an empty setting is not written" "$(cat "$INSTALL_ENV")" HIVEPAAS_INSTALLED
-  unset HIVEPAAS_ADMIN_PASSWORD HIVEPAAS_APP_DOMAIN
-  load_settings
-  check "a password reads back" "it's \$x \"quoted\"" "$HIVEPAAS_ADMIN_PASSWORD"
-  check "a domain reads back" app.example.com "$HIVEPAAS_APP_DOMAIN"
-}
+
 
 test_confirm_and_offer() {
   ASSUME_YES=1
@@ -704,6 +757,7 @@ test_help() {
   out=$(bash "$HERE/install.sh" --help)
   check "--help exits 0" 0 "$?"
   check_contains "usage" "$out" "Usage: install.sh"
+  check_contains "the settings file to fill in" "$out" "deployment/release/install.env"
   for key in ADMIN_EMAIL ADMIN_PASSWORD APP_DOMAIN ROOT_DOMAIN APP_SECRET DATA_DIR PROJECT_DATA_DIR CHANNEL \
     SWAP SWAP_SIZE_MB EARLYOOM UPGRADE_DOCKER AGENT_IMAGE RELEASE_BRANCH INSTALL_REF; do
     check_contains "--help lists HIVEPAAS_$key" "$out" "HIVEPAAS_$key"
@@ -721,6 +775,19 @@ test_refuses_without_root() {
   check_contains "saying why" "$out" "as root"
 }
 
+test_install_env_template() {
+  local file="$HERE/install.env" out key
+  COLLECTED=''
+  out=$(read_kv_file "$file" collect 2>&1)
+  check "it reads without a warning" "" "$out"
+  read_kv_file "$file" collect
+  check "the required settings, to fill in" "HIVEPAAS_ADMIN_EMAIL=|HIVEPAAS_ADMIN_PASSWORD=|HIVEPAAS_APP_DOMAIN=|" \
+    "$COLLECTED"
+  for key in $(usage | grep -oE 'HIVEPAAS_[A-Z_]+' | sort -u); do
+    check_contains "it explains $key" "$(cat "$file")" "$key"
+  done
+}
+
 # ------------------------------------------------------------------- Runner
 
 for t in $(declare -F | awk '$3 ~ /^test_/ {print $3}'); do
```

- [ ] **Step 2: Run them to see them fail**

Run: `bash deployment/release/install_test.sh`
Expected: FAIL - `317 passed, 45 failed`: `take_installed_env`, `toml_secret`, `write_secret_file` not found, `install.env` missing, the stack still giving the secret out.

- [ ] **Step 3: Write the code**

Save as `/tmp/installer-2b.diff` and run `git apply /tmp/installer-2b.diff`:

```diff
diff --git a/deployment/release/hivepaas.first-boot.yaml b/deployment/release/hivepaas.first-boot.yaml
index 06ee919b..2855dc2c 100644
--- a/deployment/release/hivepaas.first-boot.yaml
+++ b/deployment/release/hivepaas.first-boot.yaml
@@ -2,9 +2,9 @@
 #
 # The app reads the admin's account on its first boot, when it creates them,
 # and never again. Once the dashboard answers, install.sh removes these from
-# the services (docker service update --env-rm) and the password from
-# install.env, so it does not stay readable in `docker service inspect`. A later
-# deploy leaves this file out, and so does not bring them back.
+# the services (docker service update --env-rm), so the password does not stay
+# readable in `docker service inspect`. A later deploy leaves this file out, and
+# so does not bring them back.
 services:
   app:
     environment:
diff --git a/deployment/release/hivepaas.yaml b/deployment/release/hivepaas.yaml
index 09fcaa32..9d5f1f83 100644
--- a/deployment/release/hivepaas.yaml
+++ b/deployment/release/hivepaas.yaml
@@ -1,23 +1,24 @@
 # The HivePaaS stack as install.sh deploys it.
 #
 # Every value that differs between servers is a ${VAR} the installer exports
-# before `docker stack deploy`: the answers saved in /etc/hivepaas/install.env
-# and the images of the release it installs. `docker stack config -c
-# hivepaas.yaml` shows what a set of them renders to.
+# before `docker stack deploy`: the answers it was given, or read back from the
+# services an earlier run deployed, and the images of the release it installs.
+# `docker stack config -c hivepaas.yaml` shows what a set of them renders to.
 #
 # The healthchecks, update_config and restart policies are dev's, and
 # deployment/dev/hivepaas.yaml says why each number is what it is. Keep the two
 # in step.
 
 # The environment every HivePaaS process loads its configuration from: app,
-# worker, updater and agent all read the database, the cache and the secrets.
+# worker, updater and agent all read the database and the cache. The app
+# secret is in <app data>/hivepaas.toml, which app, worker and updater see at
+# /var/lib/hivepaas; the agent, on every node, needs neither it nor the JWT
+# secret, which the others get below.
 x-hivepaas-env: &hivepaas-env
   HP_ENV: ${HIVEPAAS_APP_ENV:?}
   HP_CONFIG_FILE: config/config.${HIVEPAAS_APP_ENV:?}.toml
   HP_ROOT_DOMAIN: ${HIVEPAAS_ROOT_DOMAIN:?}
   HP_APP_DOMAIN: ${HIVEPAAS_APP_DOMAIN:?}
-  HP_APP_SECRET: ${HIVEPAAS_APP_SECRET:?}
-  HP_SESSION_JWT_SECRET: ${HIVEPAAS_JWT_SECRET:?}
   HP_STORAGE_HOST_DIR: ${HIVEPAAS_DATA_DIR:?}
   HP_STORAGE_PROJECT_DATA_HOST_DIR: ${HIVEPAAS_PROJECT_DATA_DIR:?}
   HP_DB_HOST: db
@@ -176,6 +177,7 @@ services:
       local_net:
     environment:
       <<: *hivepaas-env
+      HP_SESSION_JWT_SECRET: ${HIVEPAAS_JWT_SECRET:?}
     volumes:
       - ${HIVEPAAS_DATA_DIR:?}:/var/lib/hivepaas
       - ${HIVEPAAS_DATA_DIR:?}:/host${HIVEPAAS_DATA_DIR:?}
@@ -249,6 +251,7 @@ services:
     environment:
       <<: *hivepaas-env
       HP_RUN_MODE: worker
+      HP_SESSION_JWT_SECRET: ${HIVEPAAS_JWT_SECRET:?}
     volumes:
       - ${HIVEPAAS_DATA_DIR:?}:/var/lib/hivepaas
       - ${HIVEPAAS_DATA_DIR:?}:/host${HIVEPAAS_DATA_DIR:?}
@@ -282,6 +285,7 @@ services:
     environment:
       <<: *hivepaas-env
       HP_RUN_MODE: updater
+      HP_SESSION_JWT_SECRET: ${HIVEPAAS_JWT_SECRET:?}
     volumes:
       - ${HIVEPAAS_DATA_DIR:?}:/var/lib/hivepaas
       - ${HIVEPAAS_DATA_DIR:?}:/host${HIVEPAAS_DATA_DIR:?}
diff --git a/deployment/release/install.sh b/deployment/release/install.sh
index 958d92cb..9a6c1199 100755
--- a/deployment/release/install.sh
+++ b/deployment/release/install.sh
@@ -5,17 +5,18 @@
 #   curl -fsSL https://raw.githubusercontent.com/hivepaas/hivepaas/main/deployment/release/install.sh | sudo bash
 #
 # It installs Docker when it is missing, makes the server a single-node swarm
-# and deploys the HivePaaS stack of a release channel, beta for now. Every
-# answer is saved in /etc/hivepaas/install.env. Running it again finishes an
-# interrupted install or re-checks the host: it never resets a swarm, removes a
-# service, network or volume, or generates a secret a second time.
-# `install.sh --help` lists the settings of a silent install.
+# and deploys the HivePaaS stack of a release channel, beta for now. The
+# settings end up where HivePaaS reads them: the app secret in
+# <app data>/hivepaas.toml, the rest in the environment of its services, where a
+# second run reads them back. Running it again finishes an interrupted install
+# or re-checks the host: it never resets a swarm, removes a service, network or
+# volume, or generates a secret a second time. For a silent install, fill in
+# deployment/release/install.env and pass it with --config; see --help.
 #
 # For tests only:
 #   HIVEPAAS_INSTALL_LIB=1       define the functions and stop (install_test.sh)
 #   HIVEPAAS_INSTALL_FILES_DIR   copy the stack files from this directory
 #   HIVEPAAS_RELEASE_URL         read the release info from this URL
-#   HIVEPAAS_INSTALL_ENV_FILE    where the answers are saved
 #   HIVEPAAS_TTY                 read answers from this file, not the terminal
 #   HIVEPAAS_WAIT_SECONDS        how long to wait for the dashboard (300)
 
@@ -172,8 +173,10 @@ valid_project_data_dir() {
   return 0
 }
 
+# valid_app_secret: 32 characters or more, with nothing TOML would have to
+# escape in hivepaas.toml - no spaces, quotes or backslashes.
 valid_app_secret() {
-  [ "${#1}" -ge 32 ] && [[ ! $1 =~ [[:space:]] ]]
+  [ "${#1}" -ge 32 ] && [[ ! $1 =~ [[:space:]\"\'\\] ]]
 }
 
 valid_ipv4() {
@@ -212,7 +215,7 @@ version_ge() {
 
 # ----------------------------------------------------------- Settings files
 
-# kv_quote: a value in single quotes, as install.env stores it.
+# kv_quote: a value in single quotes, for a file a shell reads.
 kv_quote() {
   local q="'"
   printf "'%s'" "${1//$q/$q\\$q$q}"
@@ -451,16 +454,8 @@ addresses_line() {
 # ---------------------------------------------------------------- Questions
 
 ADMIN_USERNAME='admin'
-INSTALL_ENV=${HIVEPAAS_INSTALL_ENV_FILE:-/etc/hivepaas/install.env}
 
-# The answers install.env keeps, so that a second run asks nothing twice and
-# generates no secret twice.
-SAVED_KEYS="HIVEPAAS_CHANNEL HIVEPAAS_ADMIN_EMAIL HIVEPAAS_ADMIN_PASSWORD HIVEPAAS_APP_DOMAIN
-HIVEPAAS_ROOT_DOMAIN HIVEPAAS_APP_SECRET HIVEPAAS_DATA_DIR HIVEPAAS_PROJECT_DATA_DIR
-HIVEPAAS_JWT_SECRET HIVEPAAS_DB_PASSWORD HIVEPAAS_REDIS_PASSWORD HIVEPAAS_AGENT_TOKEN
-HIVEPAAS_INSTALLED"
-
-# The saved answers a later run may not change: the secrets open data already
+# The settings a later run may not change: the secrets open data already
 # written with them, and the rest says where that data is and what serves it.
 FIXED_KEYS="HIVEPAAS_CHANNEL HIVEPAAS_APP_DOMAIN HIVEPAAS_ROOT_DOMAIN HIVEPAAS_APP_SECRET
 HIVEPAAS_DATA_DIR HIVEPAAS_PROJECT_DATA_DIR HIVEPAAS_JWT_SECRET HIVEPAAS_DB_PASSWORD
@@ -471,6 +466,8 @@ CONFIG_FILE=''
 HAVE_TTY=0
 MISSING=''
 RELEASE_FILE=''
+# INSTALLED=1: HivePaaS runs here and has created its admin.
+INSTALLED=0
 
 # open_tty: questions are read from the terminal, not from stdin, which
 # `curl | bash` makes the script itself. fd 3 reads answers, fd 4 shows the
@@ -586,14 +583,11 @@ take_setting() {
   printf -v "$1" '%s' "$2"
 }
 
-# take_saved KEY VALUE: a line of install.env, for a setting the environment
-# and --config left unset. One they set to another value stops the install
-# when the setting is fixed; otherwise theirs wins.
-take_saved() {
-  case " $SAVED_KEYS " in
-    *[[:space:]]"$1"[[:space:]]*) ;;
-    *) return 0 ;;
-  esac
+# take_installed KEY VALUE: a setting of the HivePaaS already on this server,
+# for what the environment and --config left unset. One they set to another
+# value stops the install when the setting is fixed; otherwise theirs wins.
+take_installed() {
+  if [ -z "$2" ]; then return 0; fi
   if [ -z "${!1:-}" ]; then
     printf -v "$1" '%s' "$2"
     return 0
@@ -601,13 +595,109 @@ take_saved() {
   if [ "${!1}" != "$2" ]; then
     case " $FIXED_KEYS " in
       *[[:space:]]"$1"[[:space:]]*)
-        die "$1 is saved in $INSTALL_ENV with another value, and it cannot change once HivePaaS" \
-          "has been installed with it. Leave it out of the environment and of --config."
+        die "$1 differs from what the HivePaaS on this server runs with, and it cannot change once" \
+          "installed. Leave it out of the environment and of --config."
         ;;
     esac
   fi
 }
 
+# channel_of_app_env ENV: the channel an app env runs, app_env_of_channel undone.
+channel_of_app_env() {
+  case "$1" in
+    beta) printf 'beta' ;;
+    production) printf 'stable' ;;
+    *) return 1 ;;
+  esac
+}
+
+# take_installed_env ENV: the settings in the environment of the hivepaas_app
+# service, one VAR=VALUE a line - where they live once installed. The admin's
+# password there means the install has not finished.
+take_installed_env() {
+  local line key value
+  INSTALLED=1
+  while IFS= read -r line; do
+    key=${line%%=*}
+    value=${line#*=}
+    case "$key" in
+      HP_ENV) take_installed HIVEPAAS_CHANNEL "$(channel_of_app_env "$value")" ;;
+      HP_APP_DOMAIN) take_installed HIVEPAAS_APP_DOMAIN "$value" ;;
+      HP_ROOT_DOMAIN) take_installed HIVEPAAS_ROOT_DOMAIN "$value" ;;
+      HP_STORAGE_HOST_DIR) take_installed HIVEPAAS_DATA_DIR "$value" ;;
+      HP_STORAGE_PROJECT_DATA_HOST_DIR) take_installed HIVEPAAS_PROJECT_DATA_DIR "$value" ;;
+      HP_SESSION_JWT_SECRET) take_installed HIVEPAAS_JWT_SECRET "$value" ;;
+      HP_DB_PASSWORD) take_installed HIVEPAAS_DB_PASSWORD "$value" ;;
+      HP_CACHE_URL)
+        value=${value#*://*:}
+        take_installed HIVEPAAS_REDIS_PASSWORD "${value%@*}"
+        ;;
+      HP_AGENT_SECRET_TOKEN) take_installed HIVEPAAS_AGENT_TOKEN "$value" ;;
+      HP_USER_ADMIN_EMAIL) take_installed HIVEPAAS_ADMIN_EMAIL "$value" ;;
+      HP_USER_ADMIN_PASSWORD)
+        INSTALLED=0
+        take_installed HIVEPAAS_ADMIN_PASSWORD "$value"
+        ;;
+    esac
+  done <<<"$1"
+}
+
+# secret_file: where the app keeps its secret - its managed settings file, in
+# the app data directory.
+secret_file() {
+  printf '%s/hivepaas.toml' "$HIVEPAAS_DATA_DIR"
+}
+
+# toml_secret FILE: the top-level `secret` of a TOML file, quoted as the app or
+# write_secret_file writes it.
+toml_secret() {
+  sed -n -e '/^[[:space:]]*\[/q' \
+    -e 's/^[[:space:]]*secret[[:space:]]*=[[:space:]]*"\([^"\\]*\)".*/\1/p' \
+    -e "s/^[[:space:]]*secret[[:space:]]*=[[:space:]]*'\\([^']*\\)'.*/\\1/p" "$1" | head -n 1
+}
+
+# write_secret_file FILE SECRET: the app's managed settings file, with the
+# settings it may hold explained. The app rewrites it itself, without the
+# comments, when its secret is rotated or a security setting is saved.
+write_secret_file() {
+  (
+    umask 077
+    cat >"$1.tmp" <<EOF
+# HivePaaS's own settings. The app reads this file whenever it starts, and
+# writes it when its secret is rotated or a security setting is saved in the
+# dashboard. It must stay readable by root alone (0600): the app refuses more.
+#
+# Every other setting is the environment of the HivePaaS services:
+#   docker service inspect hivepaas_app --format '{{range .Spec.TaskTemplate.ContainerSpec.Env}}{{println .}}{{end}}'
+
+# The key every stored secret is encrypted with, and the only one. Back this
+# file up somewhere other than this server: without it the data is unreadable.
+secret = "$2"
+
+[security]
+# Show stored secrets' values through the API and the dashboard. Default: false.
+# return_secrets_via_api = false
+# Secret types shown even while the switch above is off: "swarm-join-token".
+# always_return_secret_types = ["swarm-join-token"]
+# Let apps mount the Docker socket or a host directory, which makes them root
+# on their node. Default: false.
+# allow_privileged_apps = false
+EOF
+  ) && mv -f "$1.tmp" "$1"
+}
+
+# load_secret_file: the app secret already in the app data directory, left by
+# an earlier run or rotated by the app. It wins over one generated here, and
+# one given that differs stops the install.
+load_secret_file() {
+  local file secret
+  file=$(secret_file)
+  if [ ! -f "$file" ]; then return 0; fi
+  secret=$(toml_secret "$file") || secret=''
+  if [ -z "$secret" ]; then die "Cannot read the secret in $file."; fi
+  take_installed HIVEPAAS_APP_SECRET "$secret"
+}
+
 # normalize_settings: settings put in the form they are saved and compared in.
 normalize_settings() {
   if [ -n "${HIVEPAAS_APP_DOMAIN:-}" ]; then HIVEPAAS_APP_DOMAIN=$(lowercase "$HIVEPAAS_APP_DOMAIN"); fi
@@ -618,26 +708,27 @@ normalize_settings() {
   fi
 }
 
-# load_settings: the environment, then --config, then install.env - each only
-# for what the ones before it left unset.
+# load_settings INSTALLED_ENV: the environment, then --config, then the
+# HivePaaS already on this server - each only for what the ones before it left
+# unset. INSTALLED_ENV is the hivepaas_app service's environment, empty when
+# there is no such service.
 load_settings() {
   if [ -n "$CONFIG_FILE" ]; then
     [ -r "$CONFIG_FILE" ] || die "Cannot read $CONFIG_FILE."
     read_kv_file "$CONFIG_FILE" take_setting
   fi
   normalize_settings
-  if [ -f "$INSTALL_ENV" ]; then
-    read_kv_file "$INSTALL_ENV" take_saved
-  fi
+  if [ -n "${1:-}" ]; then take_installed_env "$1"; fi
   HIVEPAAS_CHANNEL=${HIVEPAAS_CHANNEL:-beta}
   app_env_of_channel "$HIVEPAAS_CHANNEL" >/dev/null ||
     die "HIVEPAAS_CHANNEL: '$HIVEPAAS_CHANNEL' is not a channel; it is beta or stable."
 }
 
-# ask_questions: every setting install.env does not hold yet. The admin is
-# asked for only until HivePaaS has created them.
+# ask_questions: every setting still missing. The admin is asked for only until
+# HivePaaS has created them, and the app secret only when the app data
+# directory has none.
 ask_questions() {
-  if [ "${HIVEPAAS_INSTALLED:-}" != true ]; then
+  if [ "$INSTALLED" != 1 ]; then
     answer HIVEPAAS_ADMIN_EMAIL "Admin email" valid_email "Enter an email address, like you@example.com."
     answer_password
   fi
@@ -648,8 +739,6 @@ ask_questions() {
     answer HIVEPAAS_ROOT_DOMAIN "Root domain, which apps get subdomains of" valid_root_domain \
       "Enter the app domain or a domain it is under, like example.com." "$(root_domain_of "$HIVEPAAS_APP_DOMAIN")"
   fi
-  answer HIVEPAAS_APP_SECRET "App secret, which encrypts stored secrets" valid_app_secret \
-    "The app secret needs 32 characters or more, and no spaces." "$(openssl rand -hex 32)" "Enter to generate"
   answer HIVEPAAS_DATA_DIR "App data directory" valid_data_dir \
     "Enter an absolute path of letters, digits, '.', '_' and '-', outside the system's directories." \
     /var/lib/hivepaas
@@ -658,9 +747,13 @@ ask_questions() {
     "Enter an absolute path outside the system's directories, and not the app data directory or one above it." \
     "$HIVEPAAS_DATA_DIR/project_data"
   normalize_settings
+  load_secret_file
+  answer HIVEPAAS_APP_SECRET "App secret, which encrypts stored secrets" valid_app_secret \
+    "The app secret needs 32 characters or more, and no spaces, quotes or backslashes." \
+    "$(openssl rand -hex 32)" "Enter to generate"
   if [ -n "$MISSING" ]; then
     die "No terminal to ask on, and these settings are missing:$MISSING. Set them in the environment" \
-      "or in a --config file; install.sh --help lists them."
+      "or in a --config file (deployment/release/install.env is one to fill in); install.sh --help lists them."
   fi
 }
 
@@ -671,24 +764,7 @@ generate_secrets() {
   HIVEPAAS_AGENT_TOKEN=${HIVEPAAS_AGENT_TOKEN:-$(rand_alnum 32)}
 }
 
-# save_settings: install.env, replaced whole: readable by root alone, and never
-# half written.
-save_settings() {
-  local dir="${INSTALL_ENV%/*}" tmp key
-  mkdir -p "$dir"
-  chmod 700 "$dir"
-  tmp=$(mktemp "$dir/.install.env.XXXXXX")
-  {
-    printf '# HivePaaS install settings, written by install.sh.\n'
-    printf '# HIVEPAAS_APP_SECRET is the only key to the encrypted data: keep a copy of\n'
-    printf '# this file somewhere other than this server.\n'
-    for key in $SAVED_KEYS; do
-      if [ -n "${!key:-}" ]; then printf '%s=%s\n' "$key" "$(kv_quote "${!key}")"; fi
-    done
-  } >"$tmp"
-  chmod 600 "$tmp"
-  mv -f "$tmp" "$INSTALL_ENV"
-}
+
 
 mask() {
   printf '****%s' "${1: -4}"
@@ -699,12 +775,12 @@ print_summary() {
   if [ -n "$RELEASE_FILE" ]; then
     info "Release        HivePaaS $(release_field "$RELEASE_FILE" "$HIVEPAAS_CHANNEL" appVersion) ($HIVEPAAS_CHANNEL)"
   fi
-  if [ "${HIVEPAAS_INSTALLED:-}" != true ]; then
+  if [ "$INSTALLED" != 1 ]; then
     info "Admin          $ADMIN_USERNAME, $HIVEPAAS_ADMIN_EMAIL, a password of ${#HIVEPAAS_ADMIN_PASSWORD} characters"
   fi
   info "App domain     $HIVEPAAS_APP_DOMAIN"
   info "Root domain    $HIVEPAAS_ROOT_DOMAIN"
-  info "App secret     $(mask "$HIVEPAAS_APP_SECRET")"
+  info "App secret     $(mask "$HIVEPAAS_APP_SECRET"), kept in $(secret_file)"
   info "App data       $HIVEPAAS_DATA_DIR"
   info "Project data   $HIVEPAAS_PROJECT_DATA_DIR"
   info "Addresses      $(addresses_line)"
@@ -1092,7 +1168,6 @@ ensure_docker() {
 SUBNET=10.11.0.0/16
 SUBNET_GATEWAY=10.11.0.1
 REPO_RAW=https://raw.githubusercontent.com/hivepaas/hivepaas
-CONFIG_DIR=${INSTALL_ENV%/*}
 
 ensure_swarm() {
   local state node
@@ -1180,14 +1255,21 @@ prepare_files() {
       die "Could not make the self-signed certificate."
     ok "Self-signed certificate for $HIVEPAAS_ROOT_DOMAIN, *.$HIVEPAAS_ROOT_DOMAIN and $HIVEPAAS_APP_DOMAIN."
   fi
-  fetch_install_file hivepaas.yaml "$CONFIG_DIR/hivepaas.yaml"
-  fetch_install_file hivepaas.first-boot.yaml "$CONFIG_DIR/hivepaas.first-boot.yaml"
+  # The app keeps its secret here, and rotates it here; a run after the first
+  # leaves the file alone.
+  if [ -f "$(secret_file)" ]; then
+    ok "App secret: kept in $(secret_file)."
+  else
+    write_secret_file "$(secret_file)" "$HIVEPAAS_APP_SECRET" || die "Could not write $(secret_file)."
+    ok "App secret: in $(secret_file)."
+  fi
+  fetch_install_file hivepaas.yaml "$WORK_DIR/hivepaas.yaml"
+  fetch_install_file hivepaas.first-boot.yaml "$WORK_DIR/hivepaas.first-boot.yaml"
   # The app writes Traefik's configuration from here on; a run after the first
   # leaves it alone.
   if [ ! -f "$data/traefik/etc/dynamic/dynamic_conf.yml" ]; then
     fetch_install_file traefik/dynamic_conf.yml "$data/traefik/etc/dynamic/dynamic_conf.yml"
   fi
-  ok "Stack files in $CONFIG_DIR."
 }
 
 # ------------------------------------------------------------------- Deploy
@@ -1229,22 +1311,31 @@ db_volume_exists() {
   [ -n "$(docker volume ls -q --filter name=hivepaas_db 2>/dev/null)" ]
 }
 
+# service_env SERVICE: the service's environment, one VAR=VALUE a line; empty
+# when there is no such service.
+service_env() {
+  docker service inspect "${STACK}_$1" \
+    --format '{{range .Spec.TaskTemplate.ContainerSpec.Env}}{{println .}}{{end}}' 2>/dev/null || true
+}
+
 # detect_install_state: fresh, unfinished (deployed, the admin's password not
-# yet removed) or installed.
+# yet removed) or installed. It stops where a run could only guess: HivePaaS
+# running without the file its secret is in, or a database volume left
+# without the services that held its password.
 detect_install_state() {
-  if [ ! -f "$INSTALL_ENV" ] && { service_exists app || db_volume_exists; }; then
-    die "HivePaaS is on this server already, but $INSTALL_ENV is not, and the secrets in it are what open" \
-      "its database. Put your copy of the file back, then run the installer again."
-  fi
-  if [ "${HIVEPAAS_INSTALLED:-}" = true ]; then
-    INSTALL_STATE=installed
-    if ! service_exists app && [ "$REDEPLOY" != 1 ]; then
-      die "HivePaaS was installed here, but its services are gone. Run the installer with --redeploy to deploy them again."
+  if service_exists app; then
+    if [ "$INSTALLED" = 1 ]; then INSTALL_STATE=installed; else INSTALL_STATE=unfinished; fi
+    if [ ! -f "$(secret_file)" ]; then
+      die "HivePaaS runs here, but $(secret_file) is not there, and the secret in it is the only key to" \
+        "the encrypted data. Put your copy of the file back, then run the installer again."
     fi
-  elif service_exists app; then
-    INSTALL_STATE=unfinished
   else
     INSTALL_STATE=fresh
+    if db_volume_exists && [ -z "${HIVEPAAS_DB_PASSWORD:-}" ]; then
+      die "A HivePaaS database volume is on this server, but no HivePaaS services, and its password was in" \
+        "them. Set HIVEPAAS_DB_PASSWORD to it (and keep <app data>/hivepaas.toml where it is), or remove the" \
+        "volume to start over ('docker volume ls', 'docker volume rm'), then run the installer again."
+    fi
   fi
 }
 
@@ -1316,11 +1407,11 @@ resolve_images() {
 # environment it interpolates. FIRST_BOOT=1 adds the admin's account, which the
 # app reads on its first boot only.
 deploy_stack() {
-  local -a files=(-c "$CONFIG_DIR/hivepaas.yaml")
-  if [ "$1" = 1 ]; then files+=(-c "$CONFIG_DIR/hivepaas.first-boot.yaml"); fi
+  local -a files=(-c "$WORK_DIR/hivepaas.yaml")
+  if [ "$1" = 1 ]; then files+=(-c "$WORK_DIR/hivepaas.first-boot.yaml"); fi
   (
     HIVEPAAS_APP_ENV=$(app_env_of_channel "$HIVEPAAS_CHANNEL")
-    export HIVEPAAS_APP_ENV HIVEPAAS_APP_DOMAIN HIVEPAAS_ROOT_DOMAIN HIVEPAAS_APP_SECRET \
+    export HIVEPAAS_APP_ENV HIVEPAAS_APP_DOMAIN HIVEPAAS_ROOT_DOMAIN \
       HIVEPAAS_JWT_SECRET HIVEPAAS_DATA_DIR HIVEPAAS_PROJECT_DATA_DIR HIVEPAAS_DB_PASSWORD \
       HIVEPAAS_REDIS_PASSWORD HIVEPAAS_AGENT_TOKEN HIVEPAAS_IP_RULE HIVEPAAS_ADMIN_EMAIL \
       HIVEPAAS_ADMIN_PASSWORD HP_DB_MAJOR HIVEPAAS_IMAGE_APP HIVEPAAS_IMAGE_WORKER \
@@ -1543,10 +1634,16 @@ tune_services() {
 # ------------------------------------------------------------------- Finish
 
 finish_install() {
-  HIVEPAAS_ADMIN_PASSWORD=
-  HIVEPAAS_INSTALLED=true
-  save_settings
-  ok "The admin's password is out of $INSTALL_ENV and of the services' settings."
+  HIVEPAAS_ADMIN_PASSWORD=''
+  INSTALLED=1
+  ok "The admin's password is out of the services' settings."
+}
+
+# silent_hint: how to install without being asked, for the next server.
+silent_hint() {
+  info "To install without questions, fill in the settings file and pass it:"
+  info "  curl -fsSLO $REPO_RAW/${HIVEPAAS_INSTALL_REF:-main}/deployment/release/install.env"
+  info "  sudo bash install.sh --config install.env --yes"
 }
 
 print_done() {
@@ -1566,8 +1663,11 @@ print_done() {
   warn "The certificate is self-signed until you get one in the setup, so the browser warns"
   info "  about it: accept the warning to go on."
   warn "After a restart, HivePaaS can take 30 to 60 seconds to answer."
-  warn "$INSTALL_ENV holds the app secret, the only key to your encrypted data."
+  warn "$(secret_file) holds the app secret, the only key to your encrypted data."
   info "  Keep a copy of it somewhere other than this server."
+  info "The other settings are the environment of the services: docker service inspect hivepaas_app"
+  printf '\n'
+  silent_hint
   printf '\n'
 }
 
@@ -1577,9 +1677,10 @@ usage() {
   cat <<'USAGE'
 Usage: install.sh [--yes] [--config FILE] [--redeploy]
 
-Installs HivePaaS on this server, as root. Every answer is saved in
-/etc/hivepaas/install.env; running it again finishes an interrupted install
-or re-checks the host.
+Installs HivePaaS on this server, as root. The app secret is written to
+<app data>/hivepaas.toml - back that file up - and every other setting is the
+environment of the HivePaaS services, where a second run reads it back.
+Running it again finishes an interrupted install or re-checks the host.
 
 Options:
   -y, --yes        answer yes to installing or upgrading Docker and to the
@@ -1594,8 +1695,8 @@ Settings, from the environment or --config (the environment wins):
   HIVEPAAS_ADMIN_PASSWORD      the admin's password, 10 characters or more
   HIVEPAAS_APP_DOMAIN          the dashboard's domain, e.g. hivepaas.example.com
   HIVEPAAS_ROOT_DOMAIN         the domain apps get subdomains of (default: from the app domain)
-  HIVEPAAS_APP_SECRET          the key stored secrets are encrypted with, 32 characters
-                               or more (default: generated)
+  HIVEPAAS_APP_SECRET          the key stored secrets are encrypted with, 32 characters or
+                               more, no spaces, quotes or backslashes (default: generated)
   HIVEPAAS_DATA_DIR            HivePaaS's data (default: /var/lib/hivepaas)
   HIVEPAAS_PROJECT_DATA_DIR    projects' data (default: <data dir>/project_data)
   HIVEPAAS_CHANNEL             beta or stable (default: beta)
@@ -1606,10 +1707,11 @@ Settings, from the environment or --config (the environment wins):
   HIVEPAAS_AGENT_IMAGE         the agent's image (default: from the release)
   HIVEPAAS_RELEASE_BRANCH      the branch the release info is read from (default: release)
   HIVEPAAS_INSTALL_REF         the ref the stack files are downloaded from (default: main)
+  HIVEPAAS_DB_PASSWORD         only to redeploy over a database whose services are gone
 
-Silent install:
-  curl -fsSL .../install.sh | sudo HIVEPAAS_ADMIN_EMAIL=... HIVEPAAS_ADMIN_PASSWORD=... \
-    HIVEPAAS_APP_DOMAIN=... bash -s -- --yes
+Silent install - a settings file to fill in, with every setting explained:
+  curl -fsSLO https://raw.githubusercontent.com/hivepaas/hivepaas/main/deployment/release/install.env
+  sudo bash install.sh --config install.env --yes
 USAGE
 }
 
@@ -1656,15 +1758,19 @@ main() {
   ensure_docker
 
   step "Settings"
-  load_settings
+  load_settings "$(service_env app)"
   detect_install_state
   if [ "$INSTALL_STATE" = fresh ] && ! service_exists traefik; then check_ports; fi
+  if [ "$INSTALL_STATE" = fresh ] && [ "$HAVE_TTY" = 1 ] && [ -z "$CONFIG_FILE" ]; then
+    silent_hint
+    printf '\n'
+  fi
   ask_questions
   generate_secrets
   gather_addresses
   if [ "$INSTALL_STATE" = fresh ] || [ "$REDEPLOY" = 1 ]; then fetch_release; fi
   if [ "$INSTALL_STATE" = installed ]; then
-    ok "HivePaaS is installed here, with the settings in $INSTALL_ENV."
+    ok "HivePaaS is installed here; its settings are read from its services."
     if [ "$REDEPLOY" = 1 ]; then
       warn "--redeploy puts HivePaaS's services back as the stack file has them: Traefik's settings, the"
       info "  worker and updater replicas and the app's routing labels, if HivePaaS changed them since."
@@ -1674,8 +1780,6 @@ main() {
     print_summary
     confirm "Install HivePaaS with these settings?" || die "Stopped; HivePaaS was not installed."
   fi
-  save_settings
-  ok "Settings saved in $INSTALL_ENV."
 
   step "Host memory"
   setup_swap || true
@@ -1697,7 +1801,7 @@ main() {
   ok "HivePaaS answers."
 
   step "Finish"
-  if [ "${HIVEPAAS_INSTALLED:-}" != true ]; then
+  if [ "$INSTALLED" != 1 ]; then
     finish_install
     just_installed=1
   fi
diff --git a/deployment/release/install.env b/deployment/release/install.env
new file mode 100644
index 00000000..97b61c81
--- /dev/null
+++ b/deployment/release/install.env
@@ -0,0 +1,67 @@
+# Settings for a silent install of HivePaaS. Fill in the three required ones,
+# then run, as root, in the directory this file is in:
+#
+#   curl -fsSLO https://raw.githubusercontent.com/hivepaas/hivepaas/main/deployment/release/install.sh
+#   sudo bash install.sh --config install.env --yes
+#
+# KEY=VALUE lines, read as data and never run. Quote a value with spaces in
+# single or double quotes. An empty value, or a line left commented out, takes
+# the default. Settings in the environment win over this file.
+#
+# This file holds the admin's password: delete it once HivePaaS is installed.
+# The installer keeps nothing of it: the app secret goes to
+# <app data>/hivepaas.toml, and the rest into the environment of the services.
+
+# --- Required
+
+# The admin's email, and their password: 10 characters or more.
+HIVEPAAS_ADMIN_EMAIL=
+HIVEPAAS_ADMIN_PASSWORD=
+
+# The dashboard's domain, e.g. hivepaas.example.com.
+HIVEPAAS_APP_DOMAIN=
+
+# --- Optional
+
+# The domain apps get subdomains of: the app domain or a domain it is under.
+# Default: the app domain's last two labels, or three under a suffix like
+# co.uk - example.com for hivepaas.example.com.
+#HIVEPAAS_ROOT_DOMAIN=example.com
+
+# The key every stored secret is encrypted with: 32 characters or more, no
+# spaces, quotes or backslashes. Default: 64 random hex characters. It ends up
+# in <app data>/hivepaas.toml, the one file to back up.
+#HIVEPAAS_APP_SECRET=
+
+# Where HivePaaS keeps its data. Default: /var/lib/hivepaas
+#HIVEPAAS_DATA_DIR=/var/lib/hivepaas
+
+# Where projects' data goes. Default: project_data in the data directory.
+#HIVEPAAS_PROJECT_DATA_DIR=/data/projects
+
+# The release channel: beta or stable. Default: beta
+#HIVEPAAS_CHANNEL=beta
+
+# A swap file when the server has no swap, and its size in MB. Default: yes, 2048.
+#HIVEPAAS_SWAP=false
+#HIVEPAAS_SWAP_SIZE_MB=4096
+
+# earlyoom, which kills the largest user process before a server out of memory
+# stalls. Default: installed where the distribution packages it.
+#HIVEPAAS_EARLYOOM=false
+
+# Upgrade Docker to its latest release without asking. Default: asked, and not
+# upgraded with --yes.
+#HIVEPAAS_UPGRADE_DOCKER=true
+
+# Where the release comes from: the branch the release info is read from, and
+# the ref the stack files are downloaded from. Default: release, main.
+#HIVEPAAS_RELEASE_BRANCH=main
+#HIVEPAAS_INSTALL_REF=main
+
+# The agent's image. Default: the release's, or the app image's with -agent.
+#HIVEPAAS_AGENT_IMAGE=
+
+# Only to deploy again over a database whose services were removed: the
+# password it was created with.
+#HIVEPAAS_DB_PASSWORD=
```

- [ ] **Step 4: Run the tests, under both shells**

Run: `make test-installer`
Expected: no shellcheck output, then `383 passed, 0 failed`.

Run: `docker run --rm -v "$PWD":/repo -w /repo/deployment/release bash:5.2 bash -c 'apk add -q jq openssl >/dev/null 2>&1; bash install_test.sh'`
Expected: three `skip` lines (no python3, root, no docker CLI), then `368 passed, 0 failed`.

- [ ] **Step 5: Commit**

```bash
git add deployment/release
git commit -m "feat(installer): the secret in hivepaas.toml, and nothing kept besides" -m "install.env is gone. The app secret is written once to hivepaas.toml in the app data directory, where the app keeps and rotates it, and no service gets it in its environment; every other setting stays in the services, and a second run reads it back from them. deployment/release/install.env is a template for silent installs, and --help and the installer point to it." -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 3: The e2e, with the agent built from this checkout, and the merge

**Files:**
- Modify: `deployment/release/install_e2e.sh`

**Interfaces:**
- Consumes: Tasks 1-2; the new optional `HIVEPAAS_E2E_IMAGES` (a `docker save` archive loaded into the container after the pulls).
- Produces: `make test-installer-e2e` checking the install from the template, the secret in `hivepaas.toml` and in no service, the agent running without the JWT secret, no `/etc/hivepaas`, and a lost `hivepaas.toml` stopping a run.

The released `hivepaas-agent-dev` image predates Task 1: its agent exits without the secrets. The e2e therefore runs with an agent image built from the checkout, under the release's name, loaded into the container only.

- [ ] **Step 1: Update the e2e**

Save as `/tmp/installer-3.diff` and run `git apply /tmp/installer-3.diff`:

```diff
diff --git a/deployment/release/install_e2e.sh b/deployment/release/install_e2e.sh
index b47ec311..9077fc44 100755
--- a/deployment/release/install_e2e.sh
+++ b/deployment/release/install_e2e.sh
@@ -1,10 +1,11 @@
 #!/usr/bin/env bash
 #
 # End-to-end test of install.sh, in a docker:dind container: a silent install
-# with this checkout's release info and stack files, then what a person would
-# check - the dashboard by domain and by address, signing in, the admin's
-# password gone - and the runs after it: one that must change nothing, one
-# with another app secret and one without install.env that must stop, one
+# with this checkout's release info, stack files and install.env template,
+# then what a person would check - the dashboard by domain and by address,
+# signing in, the admin's password gone, the app secret in hivepaas.toml and
+# in no service - and the runs after it: one that must change nothing, one
+# with another app secret and one without hivepaas.toml that must stop, one
 # after an address changed, and a --redeploy.
 #
 #   make test-installer-e2e     (bash deployment/release/install_e2e.sh)
@@ -13,6 +14,8 @@
 # privileged and runs a swarm of its own; HIVEPAAS_SWAP=false keeps the install
 # from setting vm.swappiness, which is the kernel's and so the host's too.
 # HIVEPAAS_E2E_KEEP=1 leaves the container running to look around in.
+# HIVEPAAS_E2E_IMAGES names a `docker save` archive loaded after the pulls, to
+# test images built from this checkout under the release's names.
 #
 # The commands in single quotes run in the container, where their $ expand:
 # shellcheck disable=SC2016
@@ -91,9 +94,15 @@ for field in dbImage redisImage traefikImage; do
 done
 EOF
 pass "images pulled"
+if [ -n "${HIVEPAAS_E2E_IMAGES:-}" ]; then
+  docker exec -i "$NAME" docker load -q <"$HIVEPAAS_E2E_IMAGES" >/dev/null
+  pass "images loaded from $HIVEPAAS_E2E_IMAGES"
+fi
 
-in_dind "HIVEPAAS_ADMIN_EMAIL=admin@example.com HIVEPAAS_ADMIN_PASSWORD=$PASSWORD HIVEPAAS_APP_DOMAIN=$DOMAIN \
-  $INSTALL >/root/install.log 2>&1" || {
+in_dind "sed -e 's/^HIVEPAAS_ADMIN_EMAIL=.*/HIVEPAAS_ADMIN_EMAIL=admin@example.com/' \
+  -e 's/^HIVEPAAS_ADMIN_PASSWORD=.*/HIVEPAAS_ADMIN_PASSWORD=$PASSWORD/' \
+  -e 's/^HIVEPAAS_APP_DOMAIN=.*/HIVEPAAS_APP_DOMAIN=$DOMAIN/' /repo/deployment/release/install.env >/root/install.env"
+in_dind "$INSTALL --config /root/install.env >/root/install.log 2>&1" || {
   in_dind 'tail -n 40 /root/install.log' >&2
   fail "the install"
 }
@@ -121,9 +130,16 @@ oom_priorities() {
     docker service inspect hivepaas_$s --format "{{.Spec.TaskTemplate.ContainerSpec.OomScoreAdj}}"; done' | xargs
 }
 expect "every system service has OOM priority -500" "-500 -500 -500 -500 -500 -500 -500" "$(oom_priorities)"
-expect "install.env is root's alone" "700 600" "$(in_dind 'stat -c %a /etc/hivepaas /etc/hivepaas/install.env' | xargs)"
-expect "install.env says installed" 1 "$(in_dind 'grep -c "^HIVEPAAS_INSTALLED=.true.$" /etc/hivepaas/install.env')"
-expect "install.env has no password" 0 "$(in_dind 'grep -c ADMIN_PASSWORD /etc/hivepaas/install.env || true')"
+expect "the app secret is in hivepaas.toml, root's alone" "600 1" "$(in_dind \
+  'stat -c %a /var/lib/hivepaas/hivepaas.toml; grep -c "^secret = \"[0-9a-f]\{64\}\"$" /var/lib/hivepaas/hivepaas.toml' |
+  xargs)"
+expect "and in no service" 0 "$(in_dind 'for s in traefik db redis app worker updater agent; do
+  docker service inspect hivepaas_$s --format "{{range .Spec.TaskTemplate.ContainerSpec.Env}}{{println .}}{{end}}"
+  done | grep -c HP_APP_SECRET || true')"
+expect "the agent runs, without the JWT secret" "1/1 0" "$(in_dind 'docker service ls --filter name=hivepaas_agent \
+  --format "{{.Replicas}}"; docker service inspect hivepaas_agent \
+  --format "{{range .Spec.TaskTemplate.ContainerSpec.Env}}{{println .}}{{end}}" | grep -c JWT || true' | xargs)"
+expect "nothing is kept in /etc/hivepaas" no "$(in_dind 'test -e /etc/hivepaas && echo yes || echo no')"
 
 versions() {
   in_dind 'for s in traefik db redis app worker updater agent; do
@@ -138,11 +154,12 @@ if in_dind "HIVEPAAS_APP_SECRET=0123456789abcdef0123456789abcdefXX $INSTALL >/ro
 fi
 pass "a run with another app secret stops"
 
-in_dind 'mv /etc/hivepaas/install.env /root/install.env.bak'
-if in_dind "$INSTALL >/root/install-lost.log 2>&1"; then fail "a run without install.env went on"; fi
-in_dind 'grep -q "Put your copy of the file back" /root/install-lost.log' || fail "a run without install.env: no reason given"
-in_dind 'mv /root/install.env.bak /etc/hivepaas/install.env'
-pass "a run without install.env stops"
+in_dind 'mv /var/lib/hivepaas/hivepaas.toml /root/hivepaas.toml.bak'
+if in_dind "$INSTALL >/root/install-lost.log 2>&1"; then fail "a run without hivepaas.toml went on"; fi
+in_dind 'grep -q "Put your copy of the file back" /root/install-lost.log' ||
+  fail "a run without hivepaas.toml: no reason given"
+in_dind 'mv /root/hivepaas.toml.bak /var/lib/hivepaas/hivepaas.toml'
+pass "a run without hivepaas.toml stops"
 
 in_dind "docker service update --detach --quiet \
   --label-add 'traefik.http.routers.x-custom-router-ip.rule=Host(\`192.0.2.1\`)' hivepaas_app >/dev/null"
```

Run: `make test-installer`
Expected: no shellcheck output, then `383 passed, 0 failed`.

- [ ] **Step 2: Build the agent from the checkout, for the e2e only**

```bash
mkdir -p /tmp/hivepaas-agent-e2e
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /tmp/hivepaas-agent-e2e/hivepaas-agent ./hivepaas_app/cmd/agent/
AGENT=$(jq -r .payload release.signed.json | base64 -d | jq -r '.beta.agentImage // empty')
AGENT=${AGENT:-hivepaas/hivepaas-agent-dev:0.1.0}
printf 'FROM %s\nCOPY hivepaas-agent /hivepaas/hivepaas-agent\n' "$AGENT" >/tmp/hivepaas-agent-e2e/Dockerfile
docker build -q --platform linux/amd64 -t "$AGENT" /tmp/hivepaas-agent-e2e
docker save "$AGENT" -o /tmp/hivepaas-agent-e2e/agent.tar
docker rmi "$AGENT" >/dev/null
```

Expected: an image ID, then `/tmp/hivepaas-agent-e2e/agent.tar` (about 190 MB). The tag is removed at once: the local swarm runs `:latest` images and must not see this one. (`hivepaas/hivepaas-agent-dev:0.1.0` is what `derive_agent_image` gives for the release's `hivepaas/hivepaas-dev:0.1.0`.)

- [ ] **Step 3: Run it**

Run: `HIVEPAAS_E2E_IMAGES=/tmp/hivepaas-agent-e2e/agent.tar make test-installer-e2e`
Expected: `ok: images pulled`, `ok: images loaded from /tmp/hivepaas-agent-e2e/agent.tar`, then every check `ok`, among them `the app secret is in hivepaas.toml, root's alone`, `and in no service`, `the agent runs, without the JWT secret`, `nothing is kept in /etc/hivepaas`, `a run without hivepaas.toml stops`, and last `all end-to-end checks passed`. Then `docker ps -a --filter name=hivepaas-install-e2e --format '{{.Names}}'` prints nothing, and `rm -rf /tmp/hivepaas-agent-e2e`.

- [ ] **Step 4: Commit**

```bash
git add deployment/release/install_e2e.sh
git commit -m "test(installer): the e2e installs from the template, and checks the secret file" -m "The first install fills in and passes deployment/release/install.env. The checks follow the secret into hivepaas.toml and out of every service, see the agent run without it, and stop a run that finds HivePaaS without its hivepaas.toml. HIVEPAAS_E2E_IMAGES loads images built from the checkout, for an app change the release does not have yet." -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

- [ ] **Step 5: Review, merge, hand over**

Review the branch against the spec and the Review Focus (a self-review: the user declined a reviewer subagent), then:

```bash
git checkout main
git merge --no-ff feat/installer-secret-file -m "Merge branch 'feat/installer-secret-file'" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
make test-installer && go test ./hivepaas_app/config/ ./hivepaas_app/cmd/internal/
git branch -d feat/installer-secret-file
```

Expected: `383 passed, 0 failed`, both packages `ok`. Do not push. Tell the user, in Vietnamese, that a server install needs images released after this merge: the agent in `hivepaas-agent-dev:0.1.0` does not start without the secrets.
