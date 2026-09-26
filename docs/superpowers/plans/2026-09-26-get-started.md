# Get Started Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A fresh install asks for the dashboard's certificate at its first boot, and shows each admin a **Get started** card on the home page - secure the dashboard, turn on two-factor authentication (recommended), connect a GitHub App (optional) - until all three are done or an admin closes it.

**Architecture:** A new `getstartedservice` works each item's state out from what exists (the dashboard domain's attached certificate, an obtain task on it, the admin's TOTP secret, an active GitHub App setting) and asks for the dashboard's certificate through the existing `domainservice.EnsureCertsForDomains` and `tasksslobtain`. The first boot calls it after the data is created; `GetMe` gives admins `setupChecklist` while the installation step is `hivepaas/get-started`; two admin endpoints ask again and close the card. The dashboard reads the checklist from the profile query it already has and polls it while a certificate is being obtained.

**Tech Stack:** Go (gin, fx, bun), React 19 + TanStack Query + zod (`../hivepaas-dashboard`).

**Spec:** `docs/superpowers/specs/2026-09-26-get-started-design.md`

## Global Constraints

- The installation step `hivepaas/get-started` (`base.InstallationStepGetStarted`) replaces `hivepaas/obtain-ssl`. No backward compatibility: an installation left at `hivepaas/obtain-ssl` shows no card.
- The first boot's request for the certificate never stops the boot: a failure is logged (`Errorf`), a request that asked for nothing is logged with why (`Warnf`).
- Nothing is forced: two-factor authentication is **recommended**, the first admin keeps `password-only`, the card never blocks the dashboard.
- One card for every admin: closing it (`POST /system/get-started/dismiss`), or `GetMe` finding all three items `done`, clears `system_statuses.next_step` for everyone.
- A certificate that arrives while the dashboard is open is not forced on the browser: the card says "Certificate installed. Quit and reopen your browser to see this site as secure". Traefik is not restarted.
- The wire: `GetMe.data.setupChecklist` = `{dashboardCert, twoFactor, githubApp}`, each `{status: todo|obtaining|failed|done, domain?, error?}`; absent for a non-admin and once the step is cleared. `POST /system/get-started/dashboard-cert` (admin; past `RetryAfter`; `ERR_CONFLICT` while obtaining) answers with the new `dashboardCert`. `POST /system/get-started/dismiss` (admin).
- The button never does nothing silently: when nothing was asked for, the answer is `409 ERR_CONFLICT` whose detail says why.
- Backend done means: `go build ./...`, `golangci-lint run ./...` on the **whole** repo (120-char lines, US spelling), `go test ./...`, `make gen-swag` with `docs/openapi/swagger.json` committed. Dashboard done means: `npx tsc --noEmit`, `npm run lint`, `npx prettier --check src`.
- Branch `feat/get-started` in both repos; merge into `main` locally with `--no-ff`; never push; commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
- Local runs: the backend on port **10099** (10000 is the user's), local-only settings as environment variables, never in `config/config.local.toml`. Nothing here writes to the user's dev database: the browser check answers the two POSTs itself.

## Decisions beyond the spec

- **`obtaining` is an obtain task that is not finished** (`task.type = ssl-obtain`, `object_id` = the certificate setting named for the domain, status `not-started` or `in-progress`), not "a setting with no content and no `lastError`": a setting whose task died would otherwise say `obtaining` forever.
- **`EnsureCertsReq.IgnoreSelfSigned`.** The certificate a fresh install signs itself is made for the root domain, which the installer lets be the dashboard's own domain; matching would then find it and ask Let's Encrypt for nothing. The dashboard's request leaves self-signed matches out.
- **`CertRequest.NotAsked`.** When asking went nowhere - a certificate already attached to the domain, one covering it, automatic certificates off, a name no authority issues for (`localhost`) - the service says so; the endpoint answers `409` with that reason and the first boot logs it.

## Review Focus

- **The authority refusing** (the domain does not point here, port 80 closed): the task tries 3 times, 2 minutes apart, so `obtaining` lasts about 4 minutes and the card reads the profile every 5 seconds all along; then `failed` with the authority's error, and **Try again** asks at once, past the 6-hour `RetryAfter`.
- **The certificate arriving while an admin watches:** the row says to quit and reopen the browser, and a toast says it too - the toast is what is left when the certificate was the last item, since `GetMe` then clears the step and the card goes.
- **A dashboard domain that is the root domain** (the self-signed default covers it): `IgnoreSelfSigned` must keep that match out, or the first boot asks for nothing; pinned by `TestWithoutSelfSignedKeepsOnlyWhatABrowserTrusts` (Task 1) and `TestNotAskedReasonSaysWhyNothingWasAskedFor` (Task 2).
- **Two admins pressing at once:** the second is refused while the first's task is live (`refuseWhileObtaining`, Task 5); two requests in the same instant are guarded only by `EnsureCertsForDomains` finding the setting the other wrote.
- **Existing installations** sit at `hivepaas/obtain-ssl` - the user's dev database among them - and show no card, as the spec says.

## Before you start

The patches were cut from trees where, measured on 2026-09-26:

- each task's new tests failed to compile before its code, and passed after it; `go build ./...` passed at every task's commit; after the last one `golangci-lint run ./...` said `0 issues.` and `go test ./...` passed (174 packages `ok`);
- `make gen-swag` wrote OpenAPI 3 (`"openapi" : "3.0.1"` on the second line). Once it wrote Swagger 2.0 instead (the converter's container had failed; the output was sent to `/dev/null`): check that line before committing, and run it again if it says `"swagger": "2.0"`;
- against the user's dev database (dashboard domain `localhost`, no certificate attached), the backend on 10099 answered an admin's `POST /system/get-started/dashboard-cert` with `409 ERR_CONFLICT`, detail `The dashboard's certificate conflicts with the existing data` + `no certificate was asked for localhost: a name with no domain part`, and the `settings` and `tasks` row counts did not change; a member got `401`;
- the dashboard passed `tsc`, `eslint` and `prettier`, and the browser check in Task 7 printed every expected line with `errors: none`.

Navigating the dashboard to `/` logs React's "Maximum update depth exceeded" a few times, on `main` as well as with this change: it is not this plan's, and the browser check goes straight to `/home/`.

---

### Task 1: Ask again past the wait, and leave self-signed matches out

**Files:**
- Modify: `hivepaas_app/service/domainservice/domainserviceimpl/cert_ensure.go`
- Modify: `hivepaas_app/service/domainservice/domainserviceimpl/cert_plan_test.go`
- Modify: `hivepaas_app/service/domainservice/service.go`

**Interfaces:**
- Consumes: nothing new.
- Produces: `domainservice.EnsureCertsReq.IgnoreRetryAfter bool` (retry a failed attempt's setting at once, never one holding a certificate); `domainservice.EnsureCertsReq.IgnoreSelfSigned bool` (matching leaves self-signed certificates out); `retryable(setting, timeNow, ignoreWait bool)`; `withoutSelfSigned(map[string]*entity.Setting) map[string]*entity.Setting`.

- [ ] **Step 0: Branch** (in `hivepaas`)

```bash
git checkout main && git checkout -b feat/get-started
```

- [ ] **Step 1: Write the failing tests** - save as `/tmp/gs-t1-tests.diff` and `git apply` it:

```diff
diff --git a/hivepaas_app/service/domainservice/domainserviceimpl/cert_plan_test.go b/hivepaas_app/service/domainservice/domainserviceimpl/cert_plan_test.go
index ca9a21bd..86aea2d1 100644
--- a/hivepaas_app/service/domainservice/domainserviceimpl/cert_plan_test.go
+++ b/hivepaas_app/service/domainservice/domainserviceimpl/cert_plan_test.go
@@ -153,9 +153,32 @@ func TestRetryableWaitsOutAFailureAndAsksAgainAfterIt(t *testing.T) {
 	}
 	for name, tc := range cases {
 		t.Run(name, func(t *testing.T) {
-			retry, reason := retryable(settingWith(tc.cert), timeNow)
+			retry, reason := retryable(settingWith(tc.cert), timeNow, false)
 			assert.Equal(t, tc.wantRetry, retry)
 			assert.Equal(t, tc.wantRetry, reason == "")
 		})
 	}
+
+	// A person asking skips the wait, but never replaces a certificate there is.
+	t.Run("asked by a person", func(t *testing.T) {
+		waiting := &entity.SSLCert{Domain: "app.example.com", LastError: "dns", RetryAfter: timeNow.Add(time.Hour)}
+		retry, _ := retryable(settingWith(waiting), timeNow, true)
+		assert.True(t, retry)
+		held := &entity.SSLCert{Domain: "app.example.com", Certificate: "-----BEGIN CERTIFICATE-----"}
+		retry, _ = retryable(settingWith(held), timeNow, true)
+		assert.False(t, retry)
+	})
+}
+
+func TestWithoutSelfSignedKeepsOnlyWhatABrowserTrusts(t *testing.T) {
+	selfSigned := &entity.Setting{Name: "example.com", Kind: string(base.SSLCertTypeSelfSigned)}
+	letsEncrypt := &entity.Setting{Name: "*.example.com", Kind: string(base.SSLCertTypeLetsEncrypt)}
+
+	trusted := withoutSelfSigned(map[string]*entity.Setting{
+		"example.com":     selfSigned,
+		"app.example.com": letsEncrypt,
+	})
+
+	assert.Equal(t, map[string]*entity.Setting{"app.example.com": letsEncrypt}, trusted)
+	assert.Empty(t, withoutSelfSigned(nil))
 }
```

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./hivepaas_app/service/domainservice/...`
Expected: FAIL to compile - `too many arguments in call to retryable` and `undefined: withoutSelfSigned`

- [ ] **Step 3: Write the code** - save as `/tmp/gs-t1-code.diff` and `git apply` it:

```diff
diff --git a/hivepaas_app/service/domainservice/domainserviceimpl/cert_ensure.go b/hivepaas_app/service/domainservice/domainserviceimpl/cert_ensure.go
index 171e2505..2bdef8e8 100644
--- a/hivepaas_app/service/domainservice/domainserviceimpl/cert_ensure.go
+++ b/hivepaas_app/service/domainservice/domainserviceimpl/cert_ensure.go
@@ -54,6 +54,9 @@ func (s *service) EnsureCertsForDomains(
 	if err != nil {
 		return nil, hperrors.Wrap(err)
 	}
+	if req.IgnoreSelfSigned {
+		matched = withoutSelfSigned(matched)
+	}
 	resp.Matched = matched
 
 	remaining := gofn.Filter(req.Domains, func(domain string) bool { return matched[domain] == nil })
@@ -95,6 +98,17 @@ func (s *service) EnsureCertsForDomains(
 	return resp, nil
 }
 
+// withoutSelfSigned is the matches a browser trusts.
+func withoutSelfSigned(matched map[string]*entity.Setting) map[string]*entity.Setting {
+	trusted := make(map[string]*entity.Setting, len(matched))
+	for domain, setting := range matched {
+		if setting.Kind != string(base.SSLCertTypeSelfSigned) {
+			trusted[domain] = setting
+		}
+	}
+	return trusted
+}
+
 // certProviders is what obtaining goes through: the account a certificate is
 // bought or requested with, and the DNS credentials a wildcard needs.
 type certProviders struct {
@@ -139,7 +153,7 @@ func (s *service) ensurePlan(
 		// Something is already responsible for this name, so a second setting
 		// would only mean a second request to the authority. What is left to
 		// decide is whether to ask again through the setting there is.
-		retry, reason := retryable(existing, timeNow)
+		retry, reason := retryable(existing, timeNow, req.IgnoreRetryAfter)
 		if !retry {
 			for _, domain := range plan.Domains {
 				resp.Skipped[domain] = reason
@@ -282,7 +296,9 @@ func (s *service) certSettingNamed(
 // a DNS record that had not propagated yet, an app that was not reachable when
 // the challenge came. An attempt still in flight is left to finish, and a
 // setting that holds a certificate is renewal's business rather than this one's.
-func retryable(setting *entity.Setting, timeNow time.Time) (retry bool, reason string) {
+// ignoreWait is a person asking: it tries again at once, whatever the last
+// attempt left, and the caller has made sure none is in flight.
+func retryable(setting *entity.Setting, timeNow time.Time, ignoreWait bool) (retry bool, reason string) {
 	cert, err := setting.AsSSLCert()
 	if err != nil {
 		return false, "a certificate named " + setting.Name + " already exists"
@@ -290,6 +306,9 @@ func retryable(setting *entity.Setting, timeNow time.Time) (retry bool, reason s
 	if cert.Certificate != "" {
 		return false, "a certificate for " + setting.Name + " exists but cannot serve this domain"
 	}
+	if ignoreWait {
+		return true, ""
+	}
 	if cert.LastError == "" {
 		return false, "a certificate for " + setting.Name + " is already being obtained"
 	}
diff --git a/hivepaas_app/service/domainservice/service.go b/hivepaas_app/service/domainservice/service.go
index af685e64..6a907dad 100644
--- a/hivepaas_app/service/domainservice/service.go
+++ b/hivepaas_app/service/domainservice/service.go
@@ -38,6 +38,15 @@ type EnsureCertsReq struct {
 	// AppID is whose routing gets applied again once a certificate arrives.
 	AppID   string
 	Domains []string
+	// IgnoreRetryAfter asks again for a certificate a failed attempt left waiting:
+	// a person asking is reason enough. One being obtained is still not asked for
+	// twice by the caller, which checks first.
+	IgnoreRetryAfter bool
+	// IgnoreSelfSigned leaves a self-signed certificate that covers a domain out
+	// of the matching, so one a browser trusts is asked for instead. The one an
+	// installation signs itself is made for its root domain, which the dashboard's
+	// domain can be.
+	IgnoreSelfSigned bool
 }
 
 type EnsureCertsResp struct {
```

- [ ] **Step 4: Run the tests to see them pass, then the whole backend**

Run: `go test ./hivepaas_app/service/domainservice/...`
Expected: `ok` for `domainserviceimpl`

```bash
go build ./... && golangci-lint run ./... && go test ./...
```

Expected: the build passes, `0 issues.`, and every package `ok`.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app
git commit -m "feat(domains): a person can ask for a certificate again, and past a self-signed match" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 2: The Get started service

**Files:**
- Create: `hivepaas_app/service/getstartedservice/getstartedserviceimpl/checklist.go`
- Create: `hivepaas_app/service/getstartedservice/getstartedserviceimpl/dashboard_cert.go`
- Create: `hivepaas_app/service/getstartedservice/getstartedserviceimpl/dashboard_cert_test.go`
- Create: `hivepaas_app/service/getstartedservice/getstartedserviceimpl/finish.go`
- Create: `hivepaas_app/service/getstartedservice/getstartedserviceimpl/service.go`
- Create: `hivepaas_app/service/getstartedservice/service.go`
- Create: `hivepaas_app/service/getstartedservice/service_test.go`
- Modify: `hivepaas_app/base/installation.go`
- Modify: `hivepaas_app/cmd/internal/system_installation.go`
- Modify: `hivepaas_app/registry/provides.go`

**Interfaces:**
- Consumes: `EnsureCertsReq.IgnoreRetryAfter`, `EnsureCertsReq.IgnoreSelfSigned` (Task 1); `hpAppService.LoadAppByKey(ctx, db, base.HivepaasAppKey)`; `settingRepo`, `systemStatusRepo`, `taskRepo`.
- Produces: `base.InstallationStepGetStarted = "hivepaas/get-started"`; package `getstartedservice` with `ItemStatus` (`ItemStatusTodo|Obtaining|Failed|Done`), `Item{Status, Domain, Error}`, `Checklist{DashboardCert, TwoFactor, GithubApp}` + `AllDone()`, `CertRequest{Tasks []*entity.Task; NotAsked string}`, and `Service`:
  - `Checklist(ctx, db database.IDB, hasTwoFactor bool) (*Checklist, error)`
  - `DashboardCert(ctx, db database.IDB) (*Item, error)`
  - `RequestDashboardCert(ctx, db database.IDB, ignoreRetryAfter bool) (*CertRequest, error)`
  - `Finish(ctx, db database.IDB) error`
- `getstartedserviceimpl.New(settingRepo, systemStatusRepo, taskRepo, domainService, hpAppService)`, provided in `registry/provides.go`.

The code step renames the step constant, so `sysInstallationInitData` sets `hivepaas/get-started` from this task on.

- [ ] **Step 1: Write the failing tests** - save as `/tmp/gs-t2-tests.diff` and `git apply` it:

```diff
diff --git a/hivepaas_app/service/getstartedservice/getstartedserviceimpl/dashboard_cert_test.go b/hivepaas_app/service/getstartedservice/getstartedserviceimpl/dashboard_cert_test.go
new file mode 100644
index 00000000..3aebf1bc
--- /dev/null
+++ b/hivepaas_app/service/getstartedservice/getstartedserviceimpl/dashboard_cert_test.go
@@ -0,0 +1,142 @@
+package getstartedserviceimpl
+
+import (
+	"testing"
+	"time"
+
+	"github.com/stretchr/testify/assert"
+
+	"github.com/hivepaas/hivepaas/hivepaas_app/base"
+	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
+	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
+	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
+)
+
+func TestDashboardDomainIsTheFirstEnabledHTTPOne(t *testing.T) {
+	routing := &entity.AppRoutingSettings{
+		ExposePublicly: true,
+		Domains: []*entity.AppDomain{
+			{Enabled: false, Domain: "off.example.com"},
+			{Enabled: true, Domain: "tcp.example.com", Protocol: base.NetworkProtocolTCP},
+			{Enabled: true, Domain: "pass.example.com", TLSPassthrough: true},
+			{Enabled: true, Domain: "dash.example.com"},
+			{Enabled: true, Domain: "second.example.com"},
+		},
+	}
+
+	domain := dashboardDomain(routing)
+
+	if assert.NotNil(t, domain) {
+		assert.Equal(t, "dash.example.com", domain.Domain)
+	}
+}
+
+func TestDashboardDomainIsNoneWithoutRouting(t *testing.T) {
+	assert.Nil(t, dashboardDomain(nil))
+	assert.Nil(t, dashboardDomain(&entity.AppRoutingSettings{
+		ExposePublicly: false,
+		Domains:        []*entity.AppDomain{{Enabled: true, Domain: "dash.example.com"}},
+	}))
+}
+
+func TestDashboardCertItem(t *testing.T) {
+	timeNow := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
+	letsEncrypt := func(expireAt time.Time) *entity.SSLCert {
+		return &entity.SSLCert{
+			CertType:    base.SSLCertTypeLetsEncrypt,
+			Certificate: "-----BEGIN CERTIFICATE-----",
+			ExpireAt:    expireAt,
+		}
+	}
+	selfSigned := &entity.SSLCert{
+		CertType:    base.SSLCertTypeSelfSigned,
+		Certificate: "-----BEGIN CERTIFICATE-----",
+		ExpireAt:    timeNow.AddDate(10, 0, 0),
+	}
+	failed := &entity.SSLCert{CertType: base.SSLCertTypeLetsEncrypt, LastError: "acme: NXDOMAIN"}
+
+	tests := []struct {
+		name      string
+		attached  *entity.SSLCert
+		pending   *entity.SSLCert
+		obtaining bool
+		want      getstartedservice.Item
+	}{
+		{
+			name: "nothing yet",
+			want: getstartedservice.Item{Status: getstartedservice.ItemStatusTodo, Domain: "dash.example.com"},
+		},
+		{
+			name:     "a trusted certificate attached",
+			attached: letsEncrypt(timeNow.AddDate(0, 2, 0)),
+			want:     getstartedservice.Item{Status: getstartedservice.ItemStatusDone, Domain: "dash.example.com"},
+		},
+		{
+			name:     "a trusted certificate with no expiry recorded",
+			attached: letsEncrypt(time.Time{}),
+			want:     getstartedservice.Item{Status: getstartedservice.ItemStatusDone, Domain: "dash.example.com"},
+		},
+		{
+			name:     "the self-signed certificate is not done",
+			attached: selfSigned,
+			want:     getstartedservice.Item{Status: getstartedservice.ItemStatusTodo, Domain: "dash.example.com"},
+		},
+		{
+			name:     "an expired certificate is not done",
+			attached: letsEncrypt(timeNow.Add(-time.Hour)),
+			want:     getstartedservice.Item{Status: getstartedservice.ItemStatusTodo, Domain: "dash.example.com"},
+		},
+		{
+			name:     "an attached setting with no content is not done",
+			attached: &entity.SSLCert{CertType: base.SSLCertTypeLetsEncrypt},
+			want:     getstartedservice.Item{Status: getstartedservice.ItemStatusTodo, Domain: "dash.example.com"},
+		},
+		{
+			name:      "a task on it",
+			pending:   &entity.SSLCert{CertType: base.SSLCertTypeLetsEncrypt},
+			obtaining: true,
+			want:      getstartedservice.Item{Status: getstartedservice.ItemStatusObtaining, Domain: "dash.example.com"},
+		},
+		{
+			name:      "a task on it again after a failure",
+			pending:   failed,
+			obtaining: true,
+			want:      getstartedservice.Item{Status: getstartedservice.ItemStatusObtaining, Domain: "dash.example.com"},
+		},
+		{
+			name:    "the last attempt failed",
+			pending: failed,
+			want: getstartedservice.Item{Status: getstartedservice.ItemStatusFailed, Domain: "dash.example.com",
+				Error: "acme: NXDOMAIN"},
+		},
+		{
+			name:     "an expired certificate being renewed",
+			attached: letsEncrypt(timeNow.Add(-time.Hour)), pending: failed, obtaining: true,
+			want: getstartedservice.Item{Status: getstartedservice.ItemStatusObtaining, Domain: "dash.example.com"},
+		},
+	}
+	for _, tt := range tests {
+		t.Run(tt.name, func(t *testing.T) {
+			got := dashboardCertItem("dash.example.com", tt.attached, tt.pending, tt.obtaining, timeNow)
+			assert.Equal(t, tt.want, got)
+		})
+	}
+}
+
+func TestNotAskedReasonSaysWhyNothingWasAskedFor(t *testing.T) {
+	skipped := &domainservice.EnsureCertsResp{
+		Skipped: map[string]string{"dash.example.com": "automatic certificates are off for this project"},
+	}
+	assert.Equal(t, "no certificate was asked for dash.example.com: automatic certificates are off for this project",
+		notAskedReason("dash.example.com", skipped))
+
+	matched := &domainservice.EnsureCertsResp{
+		Skipped: map[string]string{},
+		Matched: map[string]*entity.Setting{"dash.example.com": {Name: "*.example.com"}},
+	}
+	assert.Equal(t, `the certificate "*.example.com" covers dash.example.com already: attach it to the domain `+
+		"in the HivePaaS routing settings", notAskedReason("dash.example.com", matched))
+
+	assert.Equal(t, "no certificate was asked for dash.example.com",
+		notAskedReason("dash.example.com", &domainservice.EnsureCertsResp{}))
+}
diff --git a/hivepaas_app/service/getstartedservice/service_test.go b/hivepaas_app/service/getstartedservice/service_test.go
new file mode 100644
index 00000000..a48f4472
--- /dev/null
+++ b/hivepaas_app/service/getstartedservice/service_test.go
@@ -0,0 +1,25 @@
+package getstartedservice
+
+import (
+	"testing"
+
+	"github.com/stretchr/testify/assert"
+)
+
+func TestChecklistIsAllDoneOnlyWhenEveryItemIsDone(t *testing.T) {
+	done := Item{Status: ItemStatusDone}
+	checklist := Checklist{DashboardCert: done, TwoFactor: done, GithubApp: done}
+	assert.True(t, checklist.AllDone())
+
+	for _, status := range []ItemStatus{ItemStatusTodo, ItemStatusObtaining, ItemStatusFailed} {
+		left := checklist
+		left.DashboardCert = Item{Status: status}
+		assert.False(t, left.AllDone(), "dashboard certificate %s", status)
+	}
+	left := checklist
+	left.TwoFactor = Item{Status: ItemStatusTodo}
+	assert.False(t, left.AllDone(), "two-factor left")
+	left = checklist
+	left.GithubApp = Item{Status: ItemStatusTodo}
+	assert.False(t, left.AllDone(), "GitHub App left")
+}
```

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./hivepaas_app/service/getstartedservice/...`
Expected: FAIL to compile - `undefined: Item`, `undefined: dashboardDomain`, `undefined: dashboardCertItem`, `undefined: notAskedReason`

- [ ] **Step 3: Write the code** - save as `/tmp/gs-t2-code.diff` and `git apply` it:

```diff
diff --git a/hivepaas_app/base/installation.go b/hivepaas_app/base/installation.go
index 66a1217d..42810ffe 100644
--- a/hivepaas_app/base/installation.go
+++ b/hivepaas_app/base/installation.go
@@ -3,7 +3,7 @@ package base
 type InstallationStep string
 
 const (
-	InstallationStepNone         = ""
-	InstallationStepInitData     = "hivepaas/init-data"
-	InstallationStepObtainAppSSL = "hivepaas/obtain-ssl"
+	InstallationStepNone       = ""
+	InstallationStepInitData   = "hivepaas/init-data"
+	InstallationStepGetStarted = "hivepaas/get-started"
 )
diff --git a/hivepaas_app/cmd/internal/system_installation.go b/hivepaas_app/cmd/internal/system_installation.go
index c7255aa6..2c72544b 100644
--- a/hivepaas_app/cmd/internal/system_installation.go
+++ b/hivepaas_app/cmd/internal/system_installation.go
@@ -105,7 +105,7 @@ func sysInstallationInitData(
 			return fmt.Errorf("failed to initialize dev projects: %w", err)
 		}
 
-		sysStatus.NextStep = base.InstallationStepObtainAppSSL
+		sysStatus.NextStep = base.InstallationStepGetStarted
 		sysStatus.UpdateVer++
 		sysStatus.UpdatedAt = timeutil.NowUTC()
 		err = sysStatusRepo.Upsert(ctx, db, sysStatus,
diff --git a/hivepaas_app/registry/provides.go b/hivepaas_app/registry/provides.go
index c8a3abf4..0552ee3a 100644
--- a/hivepaas_app/registry/provides.go
+++ b/hivepaas_app/registry/provides.go
@@ -72,6 +72,7 @@ import (
 	"github.com/hivepaas/hivepaas/hivepaas_app/service/emailservice/emailserviceimpl"
 	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice/envvarserviceimpl"
 	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice/fileserviceimpl"
+	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice/getstartedserviceimpl"
 	"github.com/hivepaas/hivepaas/hivepaas_app/service/healthcheckservice/healthcheckserviceimpl"
 	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice/hpappserviceimpl"
 	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice/imagebuildserviceimpl"
@@ -376,6 +377,7 @@ var Provides = []any{
 	dbserviceimpl.New,
 	dockerapiserviceimpl.New,
 	domainserviceimpl.New,
+	getstartedserviceimpl.New,
 	emailserviceimpl.New,
 	envvarserviceimpl.New,
 	fileserviceimpl.New,
diff --git a/hivepaas_app/service/getstartedservice/getstartedserviceimpl/checklist.go b/hivepaas_app/service/getstartedservice/getstartedserviceimpl/checklist.go
new file mode 100644
index 00000000..9dcdd4d0
--- /dev/null
+++ b/hivepaas_app/service/getstartedservice/getstartedserviceimpl/checklist.go
@@ -0,0 +1,51 @@
+package getstartedserviceimpl
+
+import (
+	"context"
+
+	"github.com/hivepaas/hivepaas/hivepaas_app/base"
+	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
+	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
+	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
+	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
+	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
+)
+
+func (s *service) Checklist(
+	ctx context.Context,
+	db database.IDB,
+	hasTwoFactor bool,
+) (*getstartedservice.Checklist, error) {
+	dashboardCert, err := s.DashboardCert(ctx, db)
+	if err != nil {
+		return nil, hperrors.Wrap(err)
+	}
+	githubApp, err := s.hasGithubApp(ctx, db)
+	if err != nil {
+		return nil, hperrors.Wrap(err)
+	}
+	return &getstartedservice.Checklist{
+		DashboardCert: *dashboardCert,
+		TwoFactor:     doneIf(hasTwoFactor),
+		GithubApp:     doneIf(githubApp),
+	}, nil
+}
+
+func doneIf(done bool) getstartedservice.Item {
+	if done {
+		return getstartedservice.Item{Status: getstartedservice.ItemStatusDone}
+	}
+	return getstartedservice.Item{Status: getstartedservice.ItemStatusTodo}
+}
+
+func (s *service) hasGithubApp(ctx context.Context, db database.IDB) (bool, error) {
+	settings, _, err := s.settingRepo.List(ctx, db, entity.NewObjectScopeGlobal(), nil,
+		bunex.SelectWhere("setting.type = ?", base.SettingTypeGithubApp),
+		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
+		bunex.SelectLimit(1),
+	)
+	if err != nil {
+		return false, hperrors.Wrap(err)
+	}
+	return len(settings) > 0, nil
+}
diff --git a/hivepaas_app/service/getstartedservice/getstartedserviceimpl/dashboard_cert.go b/hivepaas_app/service/getstartedservice/getstartedserviceimpl/dashboard_cert.go
new file mode 100644
index 00000000..b909c43c
--- /dev/null
+++ b/hivepaas_app/service/getstartedservice/getstartedserviceimpl/dashboard_cert.go
@@ -0,0 +1,222 @@
+package getstartedserviceimpl
+
+import (
+	"context"
+	"errors"
+	"fmt"
+	"time"
+
+	"github.com/uptrace/bun"
+
+	"github.com/hivepaas/hivepaas/hivepaas_app/base"
+	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
+	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
+	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
+	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
+	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
+	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
+	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
+)
+
+// dashboardRouting is the HivePaaS app and its routing settings: where the
+// dashboard is served from.
+type dashboardRouting struct {
+	app     *entity.App
+	routing *entity.AppRoutingSettings
+}
+
+func (s *service) loadDashboardRouting(ctx context.Context, db database.IDB) (*dashboardRouting, error) {
+	app, err := s.hpAppService.LoadAppByKey(ctx, db, base.HivepaasAppKey)
+	if err != nil {
+		return nil, hperrors.Wrap(err)
+	}
+	setting, err := s.settingRepo.GetSingle(ctx, db, app.GetObjectScope(), base.SettingTypeAppRouting, true)
+	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
+		return nil, hperrors.Wrap(err)
+	}
+	var routing *entity.AppRoutingSettings
+	if setting != nil {
+		if routing, err = setting.AsAppRoutingSettings(); err != nil {
+			return nil, hperrors.Wrap(err)
+		}
+	}
+	return &dashboardRouting{app: app, routing: routing}, nil
+}
+
+// dashboardDomain is the domain the dashboard is known by: the first enabled
+// one it is served over HTTP at. TCP routes and TLS passed through are not
+// what a browser opens the dashboard with.
+func dashboardDomain(routing *entity.AppRoutingSettings) *entity.AppDomain {
+	for _, domain := range routing.GetActiveDomains() {
+		if domain.Protocol == base.NetworkProtocolTCP || domain.TLSPassthrough {
+			continue
+		}
+		return domain
+	}
+	return nil
+}
+
+func (s *service) DashboardCert(ctx context.Context, db database.IDB) (*getstartedservice.Item, error) {
+	dashboard, err := s.loadDashboardRouting(ctx, db)
+	if err != nil {
+		return nil, hperrors.Wrap(err)
+	}
+	domain := dashboardDomain(dashboard.routing)
+	if domain == nil {
+		return &getstartedservice.Item{Status: getstartedservice.ItemStatusTodo}, nil
+	}
+
+	var attached *entity.SSLCert
+	if domain.SSLCert.ID != "" {
+		setting, err := s.settingRepo.GetByID(ctx, db, nil, base.SettingTypeSSLCert, domain.SSLCert.ID, true)
+		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
+			return nil, hperrors.Wrap(err)
+		}
+		if setting != nil {
+			if attached, err = setting.AsSSLCert(); err != nil {
+				return nil, hperrors.Wrap(err)
+			}
+		}
+	}
+
+	pendingSetting, err := s.pendingCertSetting(ctx, db, dashboard.app, domain.Domain)
+	if err != nil {
+		return nil, hperrors.Wrap(err)
+	}
+	var pending *entity.SSLCert
+	obtaining := false
+	if pendingSetting != nil {
+		if pending, err = pendingSetting.AsSSLCert(); err != nil {
+			return nil, hperrors.Wrap(err)
+		}
+		if obtaining, err = s.hasActiveObtainTask(ctx, db, pendingSetting.ID); err != nil {
+			return nil, hperrors.Wrap(err)
+		}
+	}
+
+	item := dashboardCertItem(domain.Domain, attached, pending, obtaining, timeutil.NowUTC())
+	return &item, nil
+}
+
+// dashboardCertItem decides the certificate item from what exists: a
+// certificate attached to the domain, one being obtained for it, and whether a
+// task is on that now.
+//
+// Only an attached certificate a browser trusts is done: the self-signed one a
+// fresh install serves has the browser warn on every visit.
+func dashboardCertItem(
+	domain string,
+	attached, pending *entity.SSLCert,
+	obtaining bool,
+	timeNow time.Time,
+) getstartedservice.Item {
+	item := getstartedservice.Item{Status: getstartedservice.ItemStatusTodo, Domain: domain}
+	switch {
+	case attached != nil && attached.Certificate != "" && attached.CertType != base.SSLCertTypeSelfSigned &&
+		(attached.ExpireAt.IsZero() || attached.ExpireAt.After(timeNow)):
+		item.Status = getstartedservice.ItemStatusDone
+	case obtaining:
+		item.Status = getstartedservice.ItemStatusObtaining
+	case pending != nil && pending.LastError != "":
+		item.Status = getstartedservice.ItemStatusFailed
+		item.Error = pending.LastError
+	}
+	return item
+}
+
+// pendingCertSetting is the certificate setting obtaining is going through for
+// the domain, named for it, as domainservice names what it plans.
+func (s *service) pendingCertSetting(
+	ctx context.Context,
+	db database.IDB,
+	app *entity.App,
+	domain string,
+) (*entity.Setting, error) {
+	settings, _, err := s.settingRepo.List(ctx, db, app.GetObjectScope(), nil,
+		bunex.SelectWhere("setting.type = ?", base.SettingTypeSSLCert),
+		bunex.SelectWhere("setting.name = ?", domain),
+	)
+	if err != nil {
+		return nil, hperrors.Wrap(err)
+	}
+	if len(settings) == 0 {
+		return nil, nil //nolint:nilnil // none is not an error
+	}
+	return settings[0], nil
+}
+
+func (s *service) hasActiveObtainTask(ctx context.Context, db database.IDB, settingID string) (bool, error) {
+	tasks, _, err := s.taskRepo.List(ctx, db, nil, nil,
+		bunex.SelectWhere("task.type = ?", base.TaskTypeSSLObtain),
+		bunex.SelectWhere("task.object_id = ?", settingID),
+		bunex.SelectWhere("task.status IN (?)", bun.List([]base.TaskStatus{
+			base.TaskStatusNotStarted, base.TaskStatusInProgress})),
+		bunex.SelectLimit(1),
+	)
+	if err != nil {
+		return false, hperrors.Wrap(err)
+	}
+	return len(tasks) > 0, nil
+}
+
+func (s *service) RequestDashboardCert(
+	ctx context.Context,
+	db database.IDB,
+	ignoreRetryAfter bool,
+) (*getstartedservice.CertRequest, error) {
+	dashboard, err := s.loadDashboardRouting(ctx, db)
+	if err != nil {
+		return nil, hperrors.Wrap(err)
+	}
+	domain := dashboardDomain(dashboard.routing)
+	if domain == nil {
+		return &getstartedservice.CertRequest{
+			NotAsked: "the dashboard is served at no domain over HTTP: add one in the HivePaaS routing settings",
+		}, nil
+	}
+	if domain.SSLCert.ID != "" {
+		// The domain has a certificate someone chose: replacing it is theirs to do.
+		return &getstartedservice.CertRequest{
+			NotAsked: fmt.Sprintf("%s has a certificate attached already: change it in the HivePaaS "+
+				"routing settings", domain.Domain),
+		}, nil
+	}
+	// The dashboard's other names are asked for with it: one request, and every
+	// address it answers at is covered.
+	var wanted []string
+	for _, active := range dashboard.routing.GetActiveDomains() {
+		if active.SSLCert.ID == "" && active.Protocol != base.NetworkProtocolTCP && !active.TLSPassthrough {
+			wanted = append(wanted, active.Domain)
+		}
+	}
+	app := dashboard.app
+	resp, err := s.domainService.EnsureCertsForDomains(ctx, db, &domainservice.EnsureCertsReq{
+		Scope:            app.GetObjectScope(),
+		ProjectID:        app.ProjectID,
+		AppID:            app.ID,
+		Domains:          wanted,
+		IgnoreRetryAfter: ignoreRetryAfter,
+		IgnoreSelfSigned: true,
+	})
+	if err != nil {
+		return nil, hperrors.Wrap(err)
+	}
+	result := &getstartedservice.CertRequest{Tasks: resp.Tasks}
+	if len(resp.Tasks) == 0 && len(resp.Obtaining) == 0 {
+		result.NotAsked = notAskedReason(domain.Domain, resp)
+	}
+	return result, nil
+}
+
+// notAskedReason says why obtaining asked for nothing for the domain: what it
+// skipped the name for, or the certificate it found covering it.
+func notAskedReason(domain string, resp *domainservice.EnsureCertsResp) string {
+	if reason := resp.Skipped[domain]; reason != "" {
+		return fmt.Sprintf("no certificate was asked for %s: %s", domain, reason)
+	}
+	if cert := resp.Matched[domain]; cert != nil {
+		return fmt.Sprintf("the certificate %q covers %s already: attach it to the domain in the HivePaaS "+
+			"routing settings", cert.Name, domain)
+	}
+	return fmt.Sprintf("no certificate was asked for %s", domain)
+}
diff --git a/hivepaas_app/service/getstartedservice/getstartedserviceimpl/finish.go b/hivepaas_app/service/getstartedservice/getstartedserviceimpl/finish.go
new file mode 100644
index 00000000..5da07c47
--- /dev/null
+++ b/hivepaas_app/service/getstartedservice/getstartedserviceimpl/finish.go
@@ -0,0 +1,33 @@
+package getstartedserviceimpl
+
+import (
+	"context"
+
+	"github.com/hivepaas/hivepaas/hivepaas_app/base"
+	"github.com/hivepaas/hivepaas/hivepaas_app/config"
+	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
+	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
+	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
+	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
+)
+
+func (s *service) Finish(ctx context.Context, db database.IDB) error {
+	sysStatus, err := s.systemStatusRepo.Get(ctx, db)
+	if err != nil {
+		return hperrors.Wrap(err)
+	}
+	if sysStatus.NextStep == base.InstallationStepNone {
+		config.SetInstallationStep(base.InstallationStepNone)
+		return nil
+	}
+	sysStatus.NextStep = base.InstallationStepNone
+	sysStatus.UpdateVer++
+	sysStatus.UpdatedAt = timeutil.NowUTC()
+	err = s.systemStatusRepo.Upsert(ctx, db, sysStatus,
+		entity.SystemStatusUpsertingConflictCols, entity.SystemStatusUpsertingUpdateCols)
+	if err != nil {
+		return hperrors.Wrap(err)
+	}
+	config.SetInstallationStep(base.InstallationStepNone)
+	return nil
+}
diff --git a/hivepaas_app/service/getstartedservice/getstartedserviceimpl/service.go b/hivepaas_app/service/getstartedservice/getstartedserviceimpl/service.go
new file mode 100644
index 00000000..f7d008b6
--- /dev/null
+++ b/hivepaas_app/service/getstartedservice/getstartedserviceimpl/service.go
@@ -0,0 +1,32 @@
+package getstartedserviceimpl
+
+import (
+	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
+	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
+	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
+	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
+)
+
+type service struct {
+	settingRepo      repository.SettingRepo
+	systemStatusRepo repository.SystemStatusRepo
+	taskRepo         repository.TaskRepo
+	domainService    domainservice.Service
+	hpAppService     hpappservice.Service
+}
+
+func New(
+	settingRepo repository.SettingRepo,
+	systemStatusRepo repository.SystemStatusRepo,
+	taskRepo repository.TaskRepo,
+	domainService domainservice.Service,
+	hpAppService hpappservice.Service,
+) getstartedservice.Service {
+	return &service{
+		settingRepo:      settingRepo,
+		systemStatusRepo: systemStatusRepo,
+		taskRepo:         taskRepo,
+		domainService:    domainService,
+		hpAppService:     hpAppService,
+	}
+}
diff --git a/hivepaas_app/service/getstartedservice/service.go b/hivepaas_app/service/getstartedservice/service.go
new file mode 100644
index 00000000..5ef9fe88
--- /dev/null
+++ b/hivepaas_app/service/getstartedservice/service.go
@@ -0,0 +1,69 @@
+// Package getstartedservice is what a new installation still has to do, as the
+// dashboard's Get started card shows it: the dashboard's certificate, two-factor
+// authentication, a GitHub App.
+package getstartedservice
+
+import (
+	"context"
+
+	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
+	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
+)
+
+type ItemStatus string
+
+const (
+	ItemStatusTodo      ItemStatus = "todo"
+	ItemStatusObtaining ItemStatus = "obtaining"
+	ItemStatusFailed    ItemStatus = "failed"
+	ItemStatusDone      ItemStatus = "done"
+)
+
+// Item is where one thing to do stands. Domain and Error are the dashboard
+// certificate's: the name it is for, and why the last attempt failed.
+type Item struct {
+	Status ItemStatus
+	Domain string
+	Error  string
+}
+
+type Checklist struct {
+	DashboardCert Item
+	TwoFactor     Item
+	GithubApp     Item
+}
+
+// AllDone reports whether nothing is left to do.
+func (c *Checklist) AllDone() bool {
+	return c.DashboardCert.Status == ItemStatusDone && c.TwoFactor.Status == ItemStatusDone &&
+		c.GithubApp.Status == ItemStatusDone
+}
+
+// CertRequest is what asking for the dashboard's certificate did: the tasks to
+// schedule, or, when it asked for nothing, why not - a certificate already
+// attached or covering the domain, automatic certificates turned off, a name no
+// authority issues for. A button that does nothing without saying so is what
+// NotAsked is there to prevent.
+type CertRequest struct {
+	Tasks    []*entity.Task
+	NotAsked string
+}
+
+type Service interface {
+	// Checklist is each item's state, worked out from what exists. hasTwoFactor
+	// is the asking admin's: two-factor authentication is theirs, not the
+	// installation's.
+	Checklist(ctx context.Context, db database.IDB, hasTwoFactor bool) (*Checklist, error)
+
+	// DashboardCert is the state of the dashboard's certificate alone.
+	DashboardCert(ctx context.Context, db database.IDB) (*Item, error)
+
+	// RequestDashboardCert asks for a certificate for the dashboard's domains that
+	// have none. The tasks it returns have to be scheduled once the transaction
+	// they were written in has committed. ignoreRetryAfter skips the wait a
+	// failed attempt leaves: a person asking is reason enough to try again.
+	RequestDashboardCert(ctx context.Context, db database.IDB, ignoreRetryAfter bool) (*CertRequest, error)
+
+	// Finish clears the installation step, which hides the card for every admin.
+	Finish(ctx context.Context, db database.IDB) error
+}
```

- [ ] **Step 4: Run the tests to see them pass, then the whole backend**

Run: `go test ./hivepaas_app/service/getstartedservice/...`
Expected: `ok` for `getstartedservice` and `getstartedserviceimpl`

```bash
go build ./... && golangci-lint run ./... && go test ./...
```

Expected: the build passes, `0 issues.`, and every package `ok`.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app
git commit -m "feat(get-started): what a new installation still has to do, worked out from what exists" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 3: The first boot asks for the dashboard's certificate

**Files:**
- Create: `hivepaas_app/cmd/internal/system_installation_test.go`
- Modify: `hivepaas_app/cmd/internal/system_installation.go`

**Interfaces:**
- Consumes: `getstartedservice.Service.RequestDashboardCert` (Task 2); `queue.TaskQueue.ScheduleTask`.
- Produces: `SystemInstallation` takes `getStartedService getstartedservice.Service, taskQueue queue.TaskQueue`; `requestDashboardCert(ctx, db *database.DB, getStartedService, taskQueue, logger)`, called after `sysInstallationInitData` succeeds, with `ignoreRetryAfter = false`.

- [ ] **Step 1: Write the failing tests** - save as `/tmp/gs-t3-tests.diff` and `git apply` it:

```diff
diff --git a/hivepaas_app/cmd/internal/system_installation_test.go b/hivepaas_app/cmd/internal/system_installation_test.go
new file mode 100644
index 00000000..70a9a93b
--- /dev/null
+++ b/hivepaas_app/cmd/internal/system_installation_test.go
@@ -0,0 +1,121 @@
+package internal
+
+import (
+	"context"
+	"errors"
+	"fmt"
+	"testing"
+
+	"github.com/stretchr/testify/assert"
+
+	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
+	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
+	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
+	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
+	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
+)
+
+type certRequestingService struct {
+	getstartedservice.Service
+	tasks            []*entity.Task
+	notAsked         string
+	err              error
+	ignoreRetryAfter bool
+}
+
+func (s *certRequestingService) RequestDashboardCert(
+	_ context.Context, _ database.IDB, ignoreRetryAfter bool,
+) (*getstartedservice.CertRequest, error) {
+	s.ignoreRetryAfter = ignoreRetryAfter
+	if s.err != nil {
+		return nil, s.err
+	}
+	return &getstartedservice.CertRequest{Tasks: s.tasks, NotAsked: s.notAsked}, nil
+}
+
+type schedulingQueue struct {
+	queue.TaskQueue
+	scheduled []*entity.Task
+	err       error
+}
+
+func (q *schedulingQueue) ScheduleTask(_ context.Context, tasks ...*entity.Task) error {
+	q.scheduled = append(q.scheduled, tasks...)
+	return q.err
+}
+
+type errorLogger struct {
+	logging.Logger
+	errors   []string
+	warnings []string
+}
+
+func (l *errorLogger) Errorf(template string, args ...any) {
+	l.errors = append(l.errors, fmt.Sprintf(template, args...))
+}
+
+func (l *errorLogger) Warnf(template string, args ...any) {
+	l.warnings = append(l.warnings, fmt.Sprintf(template, args...))
+}
+
+func TestFirstBootSchedulesTheDashboardCertificate(t *testing.T) {
+	task := &entity.Task{ID: "task-1"}
+	service := &certRequestingService{tasks: []*entity.Task{task}}
+	taskQueue := &schedulingQueue{}
+	logger := &errorLogger{}
+
+	requestDashboardCert(context.Background(), nil, service, taskQueue, logger)
+
+	assert.False(t, service.ignoreRetryAfter, "the first boot keeps the wait a failure leaves")
+	assert.Equal(t, []*entity.Task{task}, taskQueue.scheduled)
+	assert.Empty(t, logger.errors)
+}
+
+func TestFirstBootGoesOnWhenTheCertificateCannotBeAskedFor(t *testing.T) {
+	service := &certRequestingService{err: errors.New("no routing")}
+	taskQueue := &schedulingQueue{}
+	logger := &errorLogger{}
+
+	requestDashboardCert(context.Background(), nil, service, taskQueue, logger)
+
+	assert.Empty(t, taskQueue.scheduled)
+	if assert.Len(t, logger.errors, 1) {
+		assert.Contains(t, logger.errors[0], "no routing")
+	}
+}
+
+func TestFirstBootLogsATaskItCannotSchedule(t *testing.T) {
+	service := &certRequestingService{tasks: []*entity.Task{{ID: "task-1"}}}
+	taskQueue := &schedulingQueue{err: errors.New("queue down")}
+	logger := &errorLogger{}
+
+	requestDashboardCert(context.Background(), nil, service, taskQueue, logger)
+
+	if assert.Len(t, logger.errors, 1) {
+		assert.Contains(t, logger.errors[0], "queue down")
+	}
+}
+
+func TestFirstBootSchedulesNothingWhenNothingIsAskedFor(t *testing.T) {
+	taskQueue := &schedulingQueue{}
+	logger := &errorLogger{}
+
+	requestDashboardCert(context.Background(), nil, &certRequestingService{}, taskQueue, logger)
+
+	assert.Empty(t, taskQueue.scheduled)
+	assert.Empty(t, logger.errors)
+}
+
+func TestFirstBootSaysWhyItAskedForNothing(t *testing.T) {
+	service := &certRequestingService{notAsked: "no certificate was asked for localhost: not a public name"}
+	taskQueue := &schedulingQueue{}
+	logger := &errorLogger{}
+
+	requestDashboardCert(context.Background(), nil, service, taskQueue, logger)
+
+	assert.Empty(t, taskQueue.scheduled)
+	assert.Empty(t, logger.errors)
+	if assert.Len(t, logger.warnings, 1) {
+		assert.Contains(t, logger.warnings[0], "not a public name")
+	}
+}
```

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./hivepaas_app/cmd/internal/`
Expected: FAIL to compile - `undefined: requestDashboardCert`

- [ ] **Step 3: Write the code** - save as `/tmp/gs-t3-code.diff` and `git apply` it:

```diff
diff --git a/hivepaas_app/cmd/internal/system_installation.go b/hivepaas_app/cmd/internal/system_installation.go
index 2c72544b..b4501a64 100644
--- a/hivepaas_app/cmd/internal/system_installation.go
+++ b/hivepaas_app/cmd/internal/system_installation.go
@@ -17,9 +17,11 @@ import (
 	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
 	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
 	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
+	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
 	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
 	"github.com/hivepaas/hivepaas/hivepaas_app/service/settinginitservice"
 	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
+	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
 )
 
 func SystemInstallation(
@@ -31,6 +33,8 @@ func SystemInstallation(
 	userService userservice.Service,
 	settingInitService settinginitservice.Service,
 	projectService projectservice.Service,
+	getStartedService getstartedservice.Service,
+	taskQueue queue.TaskQueue,
 	logger logging.Logger,
 ) {
 	stepEnabled := cfg.RunMode != config.RunModeUpdater
@@ -51,6 +55,7 @@ func SystemInstallation(
 				if err != nil {
 					return fmt.Errorf("failed to initialize system data: %w", err)
 				}
+				requestDashboardCert(ctx, db, getStartedService, taskQueue, logger)
 			}
 
 			return nil
@@ -155,3 +160,30 @@ func sysInstallationInitDevProjects(
 
 	return nil
 }
+
+// requestDashboardCert asks for the dashboard's certificate once the first
+// boot has created the data: where the domain already points here and port 80
+// is open, it is there before anyone opens the dashboard. A failure is only
+// logged - the Get started card asks again - and never stops the boot.
+func requestDashboardCert(
+	ctx context.Context,
+	db *database.DB,
+	getStartedService getstartedservice.Service,
+	taskQueue queue.TaskQueue,
+	logger logging.Logger,
+) {
+	certRequest, err := getStartedService.RequestDashboardCert(ctx, db, false)
+	if err != nil {
+		logger.Errorf("failed to request the dashboard's certificate: %v", err)
+		return
+	}
+	if certRequest.NotAsked != "" {
+		logger.Warnf("the dashboard's certificate was not requested: %s", certRequest.NotAsked)
+	}
+	if len(certRequest.Tasks) == 0 {
+		return
+	}
+	if err = taskQueue.ScheduleTask(ctx, certRequest.Tasks...); err != nil {
+		logger.Errorf("failed to schedule obtaining the dashboard's certificate: %v", err)
+	}
+}
```

- [ ] **Step 4: Run the tests to see them pass, then the whole backend**

Run: `go test ./hivepaas_app/cmd/internal/`
Expected: `ok` for `cmd/internal` - its fx wiring tests included, which prove the new parameters resolve

```bash
go build ./... && golangci-lint run ./... && go test ./...
```

Expected: the build passes, `0 issues.`, and every package `ok`.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app
git commit -m "feat(install): the first boot asks for the dashboard's certificate" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 4: `GetMe` gives admins the checklist

**Files:**
- Create: `hivepaas_app/usecase/sessionuc/get_me_setup_checklist_test.go`
- Create: `hivepaas_app/usecase/system/getstarteduc/getstarteddto/checklist.go`
- Modify: `docs/openapi/swagger.json`
- Modify: `hivepaas_app/usecase/sessionuc/get_me.go`
- Modify: `hivepaas_app/usecase/sessionuc/sessiondto/get_me.go`
- Modify: `hivepaas_app/usecase/sessionuc/uc.go`

**Interfaces:**
- Consumes: `getstartedservice.Service.Checklist`, `.Finish`, `Checklist.AllDone()` (Task 2); `base.InstallationStepGetStarted`.
- Produces: `getstarteddto.ChecklistResp{DashboardCert, TwoFactor, GithubApp *ChecklistItemResp}`, `ChecklistItemResp{Status string; Domain, Error string omitempty}`, `TransformChecklist`, `TransformChecklistItem`; `sessiondto.GetMeDataResp.SetupChecklist *getstarteddto.ChecklistResp` (`json:"setupChecklist,omitempty"`); `sessionuc.New` gains `getStartedService` (between `emailService` and `userService`); `(*sessionuc.UC).addSetupChecklist(ctx, user, respData) error`.

- [ ] **Step 1: Write the failing tests** - save as `/tmp/gs-t4-tests.diff` and `git apply` it:

```diff
diff --git a/hivepaas_app/usecase/sessionuc/get_me_setup_checklist_test.go b/hivepaas_app/usecase/sessionuc/get_me_setup_checklist_test.go
new file mode 100644
index 00000000..0e4d147e
--- /dev/null
+++ b/hivepaas_app/usecase/sessionuc/get_me_setup_checklist_test.go
@@ -0,0 +1,95 @@
+package sessionuc
+
+import (
+	"context"
+	"testing"
+
+	"github.com/stretchr/testify/assert"
+
+	"github.com/hivepaas/hivepaas/hivepaas_app/base"
+	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
+	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
+	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
+	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
+	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
+)
+
+// checklistService answers Checklist with what the test sets and counts the
+// calls to Finish. The embedded interface leaves the rest nil: GetMe reaches
+// nothing else.
+type checklistService struct {
+	getstartedservice.Service
+	checklist    getstartedservice.Checklist
+	hasTwoFactor bool
+	finished     int
+}
+
+func (s *checklistService) Checklist(
+	_ context.Context, _ database.IDB, hasTwoFactor bool,
+) (*getstartedservice.Checklist, error) {
+	s.hasTwoFactor = hasTwoFactor
+	checklist := s.checklist
+	return &checklist, nil
+}
+
+func (s *checklistService) Finish(context.Context, database.IDB) error {
+	s.finished++
+	return nil
+}
+
+func adminWithTotp(secret string) *basedto.User {
+	return &basedto.User{User: &entity.User{Role: base.UserRoleAdmin, TotpSecret: secret}}
+}
+
+func TestSetupChecklistIsGivenWhileTheStepIsGetStarted(t *testing.T) {
+	service := &checklistService{checklist: getstartedservice.Checklist{
+		DashboardCert: getstartedservice.Item{Status: getstartedservice.ItemStatusFailed,
+			Domain: "dash.example.com", Error: "acme: NXDOMAIN"},
+		TwoFactor: getstartedservice.Item{Status: getstartedservice.ItemStatusDone},
+		GithubApp: getstartedservice.Item{Status: getstartedservice.ItemStatusTodo},
+	}}
+	uc := &UC{getStartedService: service}
+	respData := &sessiondto.GetMeDataResp{NextStep: base.InstallationStepGetStarted}
+
+	err := uc.addSetupChecklist(context.Background(), adminWithTotp("secret"), respData)
+
+	assert.NoError(t, err)
+	assert.True(t, service.hasTwoFactor)
+	assert.Zero(t, service.finished)
+	assert.Equal(t, base.InstallationStepGetStarted, respData.NextStep)
+	if assert.NotNil(t, respData.SetupChecklist) {
+		assert.Equal(t, "failed", respData.SetupChecklist.DashboardCert.Status)
+		assert.Equal(t, "dash.example.com", respData.SetupChecklist.DashboardCert.Domain)
+		assert.Equal(t, "acme: NXDOMAIN", respData.SetupChecklist.DashboardCert.Error)
+		assert.Equal(t, "done", respData.SetupChecklist.TwoFactor.Status)
+		assert.Equal(t, "todo", respData.SetupChecklist.GithubApp.Status)
+	}
+}
+
+func TestSetupChecklistAllDoneClearsTheStep(t *testing.T) {
+	done := getstartedservice.Item{Status: getstartedservice.ItemStatusDone}
+	service := &checklistService{checklist: getstartedservice.Checklist{
+		DashboardCert: done, TwoFactor: done, GithubApp: done,
+	}}
+	uc := &UC{getStartedService: service}
+	respData := &sessiondto.GetMeDataResp{NextStep: base.InstallationStepGetStarted}
+
+	err := uc.addSetupChecklist(context.Background(), adminWithTotp("secret"), respData)
+
+	assert.NoError(t, err)
+	assert.Equal(t, 1, service.finished)
+	assert.Empty(t, respData.NextStep)
+	assert.Nil(t, respData.SetupChecklist)
+}
+
+func TestSetupChecklistIsAbsentForAnotherStep(t *testing.T) {
+	service := &checklistService{}
+	uc := &UC{getStartedService: service}
+	respData := &sessiondto.GetMeDataResp{NextStep: base.InstallationStepInitData}
+
+	err := uc.addSetupChecklist(context.Background(), adminWithTotp(""), respData)
+
+	assert.NoError(t, err)
+	assert.Nil(t, respData.SetupChecklist)
+	assert.Zero(t, service.finished)
+}
```

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./hivepaas_app/usecase/sessionuc/`
Expected: FAIL to compile - `unknown field getStartedService in struct literal of type UC`, `uc.addSetupChecklist undefined`, `respData.SetupChecklist undefined`

- [ ] **Step 3: Write the code** - save as `/tmp/gs-t4-code.diff` and `git apply` it:

```diff
diff --git a/hivepaas_app/usecase/sessionuc/get_me.go b/hivepaas_app/usecase/sessionuc/get_me.go
index 9f43502b..16ff7f47 100644
--- a/hivepaas_app/usecase/sessionuc/get_me.go
+++ b/hivepaas_app/usecase/sessionuc/get_me.go
@@ -10,6 +10,7 @@ import (
 	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
 	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
 	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
+	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/getstarteduc/getstarteddto"
 )
 
 func (uc *UC) GetMe(
@@ -48,6 +49,9 @@ func (uc *UC) GetMe(
 		}
 		config.SetInstallationStep(sysStatus.NextStep)
 		respData.NextStep = string(sysStatus.NextStep)
+		if err = uc.addSetupChecklist(ctx, user, respData); err != nil {
+			return nil, hperrors.Wrap(err)
+		}
 	}
 
 	if user.Status == base.UserStatusPending && user.TotpSecret == "" {
@@ -58,3 +62,29 @@ func (uc *UC) GetMe(
 		Data: respData,
 	}, nil
 }
+
+// addSetupChecklist gives an admin what the installation still has to do, while
+// the step says so. Finding all of it done clears the step, so the card goes
+// for every admin without anyone closing it.
+func (uc *UC) addSetupChecklist(
+	ctx context.Context,
+	user *basedto.User,
+	respData *sessiondto.GetMeDataResp,
+) error {
+	if respData.NextStep != base.InstallationStepGetStarted {
+		return nil
+	}
+	checklist, err := uc.getStartedService.Checklist(ctx, uc.db, user.TotpSecret != "")
+	if err != nil {
+		return hperrors.Wrap(err)
+	}
+	if checklist.AllDone() {
+		if err = uc.getStartedService.Finish(ctx, uc.db); err != nil {
+			return hperrors.Wrap(err)
+		}
+		respData.NextStep = ""
+		return nil
+	}
+	respData.SetupChecklist = getstarteddto.TransformChecklist(checklist)
+	return nil
+}
diff --git a/hivepaas_app/usecase/sessionuc/sessiondto/get_me.go b/hivepaas_app/usecase/sessionuc/sessiondto/get_me.go
index 772fcc61..e314157f 100644
--- a/hivepaas_app/usecase/sessionuc/sessiondto/get_me.go
+++ b/hivepaas_app/usecase/sessionuc/sessiondto/get_me.go
@@ -4,6 +4,7 @@ import (
 	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
 	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
 	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
+	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/getstarteduc/getstarteddto"
 	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
 )
 
@@ -23,6 +24,9 @@ type GetMeResp struct {
 type GetMeDataResp struct {
 	NextStep string                   `json:"nextStep,omitempty"`
 	User     *userdto.UserDetailsResp `json:"user"`
+	// SetupChecklist is what the installation still has to do, for an admin
+	// while nextStep is hivepaas/get-started.
+	SetupChecklist *getstarteddto.ChecklistResp `json:"setupChecklist,omitempty"`
 }
 
 func TransformUserDetails(user *entity.User) (resp *userdto.UserDetailsResp, err error) {
diff --git a/hivepaas_app/usecase/sessionuc/uc.go b/hivepaas_app/usecase/sessionuc/uc.go
index 33bda6e1..de5c9a31 100644
--- a/hivepaas_app/usecase/sessionuc/uc.go
+++ b/hivepaas_app/usecase/sessionuc/uc.go
@@ -7,6 +7,7 @@ import (
 	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
 	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
 	"github.com/hivepaas/hivepaas/hivepaas_app/service/emailservice"
+	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
 	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
 )
 
@@ -21,9 +22,10 @@ type UC struct {
 	userRepo               repository.UserRepo
 	userTokenRepo          cacherepository.UserTokenRepo
 
-	auditService auditservice.Service
-	emailService emailservice.Service
-	userService  userservice.Service
+	auditService      auditservice.Service
+	emailService      emailservice.Service
+	getStartedService getstartedservice.Service
+	userService       userservice.Service
 
 	permissionManager permission.Manager
 }
@@ -41,6 +43,7 @@ func New(
 
 	auditService auditservice.Service,
 	emailService emailservice.Service,
+	getStartedService getstartedservice.Service,
 	userService userservice.Service,
 
 	permissionManager permission.Manager,
@@ -56,9 +59,10 @@ func New(
 		userRepo:               userRepo,
 		userTokenRepo:          userTokenRepo,
 
-		auditService: auditService,
-		emailService: emailService,
-		userService:  userService,
+		auditService:      auditService,
+		emailService:      emailService,
+		getStartedService: getStartedService,
+		userService:       userService,
 
 		permissionManager: permissionManager,
 	}
diff --git a/hivepaas_app/usecase/system/getstarteduc/getstarteddto/checklist.go b/hivepaas_app/usecase/system/getstarteduc/getstarteddto/checklist.go
new file mode 100644
index 00000000..a56171b7
--- /dev/null
+++ b/hivepaas_app/usecase/system/getstarteduc/getstarteddto/checklist.go
@@ -0,0 +1,41 @@
+package getstarteddto
+
+import (
+	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
+)
+
+// ChecklistResp is what a new installation still has to do, for the dashboard's
+// Get started card. Each item's status is todo, obtaining (the certificate
+// only), failed (the certificate only) or done.
+type ChecklistResp struct {
+	DashboardCert *ChecklistItemResp `json:"dashboardCert"`
+	TwoFactor     *ChecklistItemResp `json:"twoFactor"`
+	GithubApp     *ChecklistItemResp `json:"githubApp"`
+}
+
+type ChecklistItemResp struct {
+	Status string `json:"status"`
+	// Domain is the dashboard certificate's: the name it is for.
+	Domain string `json:"domain,omitempty"`
+	// Error is why the last attempt at the dashboard's certificate failed.
+	Error string `json:"error,omitempty"`
+}
+
+func TransformChecklist(checklist *getstartedservice.Checklist) *ChecklistResp {
+	if checklist == nil {
+		return nil
+	}
+	return &ChecklistResp{
+		DashboardCert: TransformChecklistItem(&checklist.DashboardCert),
+		TwoFactor:     TransformChecklistItem(&checklist.TwoFactor),
+		GithubApp:     TransformChecklistItem(&checklist.GithubApp),
+	}
+}
+
+func TransformChecklistItem(item *getstartedservice.Item) *ChecklistItemResp {
+	return &ChecklistItemResp{
+		Status: string(item.Status),
+		Domain: item.Domain,
+		Error:  item.Error,
+	}
+}
```

Then regenerate the API document:

```bash
make gen-swag && sed -n 2p docs/openapi/swagger.json
```

Expected: `  "openapi" : "3.0.1",`, and `git diff --stat docs/openapi/swagger.json` shows about 40 lines added: `getstarteddto.ChecklistResp`, `getstarteddto.ChecklistItemResp`, and `setupChecklist` in `sessiondto.GetMeDataResp`.

- [ ] **Step 4: Run the tests to see them pass, then the whole backend**

Run: `go test ./hivepaas_app/usecase/sessionuc/`
Expected: `ok` for `sessionuc`

```bash
go build ./... && golangci-lint run ./... && go test ./...
```

Expected: the build passes, `0 issues.`, and every package `ok`.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app docs/openapi/swagger.json
git commit -m "feat(session): GetMe gives an admin what the installation still has to do" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 5: The two endpoints

**Files:**
- Create: `hivepaas_app/interface/api/handler/systemhandler/get_started.go`
- Create: `hivepaas_app/usecase/system/getstarteduc/dashboard_cert_request.go`
- Create: `hivepaas_app/usecase/system/getstarteduc/dashboard_cert_request_test.go`
- Create: `hivepaas_app/usecase/system/getstarteduc/dismiss.go`
- Create: `hivepaas_app/usecase/system/getstarteduc/getstarteddto/dashboard_cert_request.go`
- Create: `hivepaas_app/usecase/system/getstarteduc/getstarteddto/dismiss.go`
- Create: `hivepaas_app/usecase/system/getstarteduc/uc.go`
- Modify: `docs/openapi/swagger.json`
- Modify: `hivepaas_app/interface/api/handler/systemhandler/handler.go`
- Modify: `hivepaas_app/interface/api/server/router_system.go`
- Modify: `hivepaas_app/registry/provides.go`

**Interfaces:**
- Consumes: `getstartedservice.Service.DashboardCert`, `.RequestDashboardCert`, `.Finish`, `CertRequest.NotAsked` (Task 2); `getstarteddto.TransformChecklistItem`, `ChecklistItemResp` (Task 4).
- Produces: `getstarteduc.New(db *database.DB, getStartedService, taskQueue queue.TaskQueue) *UC`; `(*UC).RequestDashboardCert(ctx, auth, *getstarteddto.RequestDashboardCertReq) (*getstarteddto.RequestDashboardCertResp, error)` with `Data *ChecklistItemResp`; `(*UC).Dismiss(ctx, auth, *getstarteddto.DismissReq) (*getstarteddto.DismissResp, error)`; `refuseWhileObtaining`, `refuseNotAsked`, `requireAdmin`; `systemhandler.New` gains `getStartedUC *getstarteduc.UC` (last); routes `POST /system/get-started/dashboard-cert` (`requestDashboardCert`) and `POST /system/get-started/dismiss` (`dismissGetStarted`), both behind `System` write access.

- [ ] **Step 1: Write the failing tests** - save as `/tmp/gs-t5-tests.diff` and `git apply` it:

```diff
diff --git a/hivepaas_app/usecase/system/getstarteduc/dashboard_cert_request_test.go b/hivepaas_app/usecase/system/getstarteduc/dashboard_cert_request_test.go
new file mode 100644
index 00000000..083388c9
--- /dev/null
+++ b/hivepaas_app/usecase/system/getstarteduc/dashboard_cert_request_test.go
@@ -0,0 +1,48 @@
+package getstarteduc
+
+import (
+	"errors"
+	"testing"
+
+	"github.com/stretchr/testify/assert"
+
+	"github.com/hivepaas/hivepaas/hivepaas_app/base"
+	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
+	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
+	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
+	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
+	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
+)
+
+func TestRefuseWhileObtainingRefusesOnlyAnAttemptUnderWay(t *testing.T) {
+	err := refuseWhileObtaining(&getstartedservice.Item{Status: getstartedservice.ItemStatusObtaining})
+	assert.True(t, errors.Is(err, hperrors.ErrConflict), "got %v", err)
+
+	for _, status := range []getstartedservice.ItemStatus{
+		getstartedservice.ItemStatusTodo, getstartedservice.ItemStatusFailed, getstartedservice.ItemStatusDone,
+	} {
+		assert.NoError(t, refuseWhileObtaining(&getstartedservice.Item{Status: status}), string(status))
+	}
+}
+
+func TestRefuseNotAskedSaysWhy(t *testing.T) {
+	assert.NoError(t, refuseNotAsked(&getstartedservice.CertRequest{Tasks: []*entity.Task{{ID: "task-1"}}}))
+
+	err := refuseNotAsked(&getstartedservice.CertRequest{NotAsked: "dash.example.com has a certificate attached"})
+	assert.True(t, errors.Is(err, hperrors.ErrConflict), "got %v", err)
+	var hpErr hperrors.HPError
+	if assert.True(t, errors.As(err, &hpErr)) {
+		assert.Contains(t, hpErr.Build(translation.LangEn).Detail, "dash.example.com has a certificate attached")
+	}
+}
+
+func TestRequireAdmin(t *testing.T) {
+	asRole := func(role base.UserRole) *basedto.Auth {
+		return &basedto.Auth{User: &basedto.User{User: &entity.User{Role: role}}}
+	}
+
+	assert.NoError(t, requireAdmin(asRole(base.UserRoleAdmin)))
+	assert.True(t, errors.Is(requireAdmin(asRole(base.UserRoleMember)), hperrors.ErrForbidden))
+	assert.True(t, errors.Is(requireAdmin(nil), hperrors.ErrForbidden))
+	assert.True(t, errors.Is(requireAdmin(&basedto.Auth{}), hperrors.ErrForbidden))
+}
```

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./hivepaas_app/usecase/system/getstarteduc/`
Expected: FAIL to compile - `undefined: refuseWhileObtaining`, `undefined: refuseNotAsked`, `undefined: requireAdmin`

- [ ] **Step 3: Write the code** - save as `/tmp/gs-t5-code.diff` and `git apply` it:

```diff
diff --git a/hivepaas_app/interface/api/handler/systemhandler/get_started.go b/hivepaas_app/interface/api/handler/systemhandler/get_started.go
new file mode 100644
index 00000000..719af7ea
--- /dev/null
+++ b/hivepaas_app/interface/api/handler/systemhandler/get_started.go
@@ -0,0 +1,74 @@
+package systemhandler
+
+import (
+	"net/http"
+
+	"github.com/gin-gonic/gin"
+
+	"github.com/hivepaas/hivepaas/hivepaas_app/base"
+	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
+	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
+	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/getstarteduc/getstarteddto"
+)
+
+// RequestDashboardCert Asks for the dashboard's certificate now
+// @Summary Asks for the dashboard's certificate now
+// @Description For the Get started card: asks for a certificate for the dashboard's domain, past
+// @Description the wait a failed attempt leaves. Refused with ERR_CONFLICT while one is being
+// @Description obtained. Answers with where the certificate stands.
+// @Tags    system
+// @Produce json
+// @Id      requestDashboardCert
+// @Success 200 {object} getstarteddto.RequestDashboardCertResp
+// @Failure 400 {object} hperrors.ErrorInfo
+// @Failure 409 {object} hperrors.ErrorInfo
+// @Failure 500 {object} hperrors.ErrorInfo
+// @Router  /system/get-started/dashboard-cert [post]
+func (h *Handler) RequestDashboardCert(ctx *gin.Context) {
+	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
+		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
+		Module:          base.ResourceModuleSystem,
+	})
+	if err != nil {
+		h.RenderError(ctx, err)
+		return
+	}
+
+	resp, err := h.getStartedUC.RequestDashboardCert(h.RequestCtx(ctx), auth,
+		getstarteddto.NewRequestDashboardCertReq())
+	if err != nil {
+		h.RenderError(ctx, err)
+		return
+	}
+
+	ctx.JSON(http.StatusOK, resp)
+}
+
+// DismissGetStarted Closes the Get started card
+// @Summary Closes the Get started card
+// @Description Clears the installation step, which hides the card for every admin.
+// @Tags    system
+// @Produce json
+// @Id      dismissGetStarted
+// @Success 200 {object} getstarteddto.DismissResp
+// @Failure 400 {object} hperrors.ErrorInfo
+// @Failure 500 {object} hperrors.ErrorInfo
+// @Router  /system/get-started/dismiss [post]
+func (h *Handler) DismissGetStarted(ctx *gin.Context) {
+	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
+		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
+		Module:          base.ResourceModuleSystem,
+	})
+	if err != nil {
+		h.RenderError(ctx, err)
+		return
+	}
+
+	resp, err := h.getStartedUC.Dismiss(h.RequestCtx(ctx), auth, getstarteddto.NewDismissReq())
+	if err != nil {
+		h.RenderError(ctx, err)
+		return
+	}
+
+	ctx.JSON(http.StatusOK, resp)
+}
diff --git a/hivepaas_app/interface/api/handler/systemhandler/handler.go b/hivepaas_app/interface/api/handler/systemhandler/handler.go
index 4be81cea..bc0922fc 100644
--- a/hivepaas_app/interface/api/handler/systemhandler/handler.go
+++ b/hivepaas_app/interface/api/handler/systemhandler/handler.go
@@ -5,6 +5,7 @@ import (
 	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/auditloghandler"
 	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
 	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/taskhandler"
+	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/getstarteduc"
 	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/syserroruc"
 	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/sysstatusuc"
 )
@@ -16,6 +17,7 @@ type Handler struct {
 	taskHandler     *taskhandler.Handler
 	sysErrorUC      *syserroruc.UC
 	sysStatusUC     *sysstatusuc.UC
+	getStartedUC    *getstarteduc.UC
 }
 
 func New(
@@ -25,6 +27,7 @@ func New(
 	taskHandler *taskhandler.Handler,
 	sysErrorUC *syserroruc.UC,
 	sysStatusUC *sysstatusuc.UC,
+	getStartedUC *getstarteduc.UC,
 ) *Handler {
 	return &Handler{
 		BaseHandler:     baseHandler,
@@ -33,5 +36,6 @@ func New(
 		taskHandler:     taskHandler,
 		sysErrorUC:      sysErrorUC,
 		sysStatusUC:     sysStatusUC,
+		getStartedUC:    getStartedUC,
 	}
 }
diff --git a/hivepaas_app/interface/api/server/router_system.go b/hivepaas_app/interface/api/server/router_system.go
index 36b34474..6bbcad98 100644
--- a/hivepaas_app/interface/api/server/router_system.go
+++ b/hivepaas_app/interface/api/server/router_system.go
@@ -24,6 +24,12 @@ func (s *HTTPServer) registerSystemRoutes(apiGroup *gin.RouterGroup) {
 		statusGroup.GET("/db", systemHandler.GetDBStats)
 	}
 
+	{ // Get started group: the dashboard's first-login card
+		getStartedGroup := systemGroup.Group("/get-started")
+		getStartedGroup.POST("/dashboard-cert", systemHandler.RequestDashboardCert)
+		getStartedGroup.POST("/dismiss", systemHandler.DismissGetStarted)
+	}
+
 	{ // Error group
 		errorGroup := systemGroup.Group("/errors")
 		errorGroup.GET("", systemHandler.ListSysError)
diff --git a/hivepaas_app/registry/provides.go b/hivepaas_app/registry/provides.go
index 0552ee3a..abb6a221 100644
--- a/hivepaas_app/registry/provides.go
+++ b/hivepaas_app/registry/provides.go
@@ -173,6 +173,7 @@ import (
 	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslprovideruc"
 	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc"
 	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/supportuc"
+	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/getstarteduc"
 	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc"
 	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc"
 	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/syserroruc"
@@ -342,6 +343,7 @@ var Provides = []any{
 	homeuc.New,
 	syserroruc.New,
 	sysstatusuc.New,
+	getstarteduc.New,
 	systembackupuc.New,
 	systemcleanupuc.New,
 	specuc.New,
diff --git a/hivepaas_app/usecase/system/getstarteduc/dashboard_cert_request.go b/hivepaas_app/usecase/system/getstarteduc/dashboard_cert_request.go
new file mode 100644
index 00000000..841db4c2
--- /dev/null
+++ b/hivepaas_app/usecase/system/getstarteduc/dashboard_cert_request.go
@@ -0,0 +1,87 @@
+package getstarteduc
+
+import (
+	"context"
+
+	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
+	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
+	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
+	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
+	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
+	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
+	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/getstarteduc/getstarteddto"
+)
+
+// RequestDashboardCert asks for the dashboard's certificate now, past the wait
+// a failed attempt leaves. One already being obtained is refused rather than
+// asked for twice.
+func (uc *UC) RequestDashboardCert(
+	ctx context.Context,
+	auth *basedto.Auth,
+	_ *getstarteddto.RequestDashboardCertReq,
+) (*getstarteddto.RequestDashboardCertResp, error) {
+	if err := requireAdmin(auth); err != nil {
+		return nil, hperrors.Wrap(err)
+	}
+
+	var tasks []*entity.Task
+	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
+		item, err := uc.getStartedService.DashboardCert(ctx, db)
+		if err != nil {
+			return hperrors.Wrap(err)
+		}
+		if err = refuseWhileObtaining(item); err != nil {
+			return hperrors.Wrap(err)
+		}
+		certRequest, err := uc.getStartedService.RequestDashboardCert(ctx, db, true)
+		if err != nil {
+			return hperrors.Wrap(err)
+		}
+		if err = refuseNotAsked(certRequest); err != nil {
+			return hperrors.Wrap(err)
+		}
+		tasks = certRequest.Tasks
+		return nil
+	})
+	if err != nil {
+		return nil, hperrors.Wrap(err)
+	}
+
+	// A task can be picked up only once its row exists, which is once the
+	// transaction has committed. A failure to schedule it only delays it: the
+	// queue's own scan finds it.
+	_ = uc.taskQueue.ScheduleTask(ctx, tasks...)
+
+	item, err := uc.getStartedService.DashboardCert(ctx, uc.db)
+	if err != nil {
+		return nil, hperrors.Wrap(err)
+	}
+	return &getstarteddto.RequestDashboardCertResp{
+		Meta: &basedto.Meta{},
+		Data: getstarteddto.TransformChecklistItem(item),
+	}, nil
+}
+
+func refuseWhileObtaining(item *getstartedservice.Item) error {
+	if item.Status == getstartedservice.ItemStatusObtaining {
+		return hperrors.NewConflict("The dashboard's certificate").
+			WithExtraDetail("it is being obtained already")
+	}
+	return nil
+}
+
+// refuseNotAsked turns a request that asked for nothing into an answer that
+// says why, rather than a button that seems to do nothing.
+func refuseNotAsked(certRequest *getstartedservice.CertRequest) error {
+	if certRequest.NotAsked != "" {
+		return hperrors.NewConflict("The dashboard's certificate").WithExtraDetail("%s", certRequest.NotAsked)
+	}
+	return nil
+}
+
+func requireAdmin(auth *basedto.Auth) error {
+	if auth == nil || auth.User == nil || !auth.User.IsAdmin() {
+		return hperrors.Wrap(hperrors.ErrForbidden)
+	}
+	return nil
+}
diff --git a/hivepaas_app/usecase/system/getstarteduc/dismiss.go b/hivepaas_app/usecase/system/getstarteduc/dismiss.go
new file mode 100644
index 00000000..11ce65d8
--- /dev/null
+++ b/hivepaas_app/usecase/system/getstarteduc/dismiss.go
@@ -0,0 +1,25 @@
+package getstarteduc
+
+import (
+	"context"
+
+	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
+	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
+	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/getstarteduc/getstarteddto"
+)
+
+// Dismiss closes the Get started card, for every admin: what it lists belongs
+// to the installation.
+func (uc *UC) Dismiss(
+	ctx context.Context,
+	auth *basedto.Auth,
+	_ *getstarteddto.DismissReq,
+) (*getstarteddto.DismissResp, error) {
+	if err := requireAdmin(auth); err != nil {
+		return nil, hperrors.Wrap(err)
+	}
+	if err := uc.getStartedService.Finish(ctx, uc.db); err != nil {
+		return nil, hperrors.Wrap(err)
+	}
+	return &getstarteddto.DismissResp{Meta: &basedto.Meta{}}, nil
+}
diff --git a/hivepaas_app/usecase/system/getstarteduc/getstarteddto/dashboard_cert_request.go b/hivepaas_app/usecase/system/getstarteduc/getstarteddto/dashboard_cert_request.go
new file mode 100644
index 00000000..88c9fc16
--- /dev/null
+++ b/hivepaas_app/usecase/system/getstarteduc/getstarteddto/dashboard_cert_request.go
@@ -0,0 +1,16 @@
+package getstarteddto
+
+import (
+	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
+)
+
+type RequestDashboardCertReq struct{}
+
+func NewRequestDashboardCertReq() *RequestDashboardCertReq {
+	return &RequestDashboardCertReq{}
+}
+
+type RequestDashboardCertResp struct {
+	Meta *basedto.Meta      `json:"meta"`
+	Data *ChecklistItemResp `json:"data"`
+}
diff --git a/hivepaas_app/usecase/system/getstarteduc/getstarteddto/dismiss.go b/hivepaas_app/usecase/system/getstarteduc/getstarteddto/dismiss.go
new file mode 100644
index 00000000..5f5c140f
--- /dev/null
+++ b/hivepaas_app/usecase/system/getstarteduc/getstarteddto/dismiss.go
@@ -0,0 +1,15 @@
+package getstarteddto
+
+import (
+	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
+)
+
+type DismissReq struct{}
+
+func NewDismissReq() *DismissReq {
+	return &DismissReq{}
+}
+
+type DismissResp struct {
+	Meta *basedto.Meta `json:"meta"`
+}
diff --git a/hivepaas_app/usecase/system/getstarteduc/uc.go b/hivepaas_app/usecase/system/getstarteduc/uc.go
new file mode 100644
index 00000000..62cfd3c1
--- /dev/null
+++ b/hivepaas_app/usecase/system/getstarteduc/uc.go
@@ -0,0 +1,27 @@
+// Package getstarteduc serves the dashboard's Get started card: asking again
+// for the dashboard's certificate, and closing the card.
+package getstarteduc
+
+import (
+	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
+	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
+	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
+)
+
+type UC struct {
+	db                *database.DB
+	getStartedService getstartedservice.Service
+	taskQueue         queue.TaskQueue
+}
+
+func New(
+	db *database.DB,
+	getStartedService getstartedservice.Service,
+	taskQueue queue.TaskQueue,
+) *UC {
+	return &UC{
+		db:                db,
+		getStartedService: getStartedService,
+		taskQueue:         taskQueue,
+	}
+}
```

Then regenerate the API document:

```bash
make gen-swag && sed -n 2p docs/openapi/swagger.json
```

Expected: `  "openapi" : "3.0.1",`, and about 110 more lines: the two paths, `getstarteddto.RequestDashboardCertResp` and `getstarteddto.DismissResp`.

Then check the routes against a running backend, without writing anything: an unauthenticated call is refused before any of this code runs.

```bash
go build -o /tmp/hp-gs ./hivepaas_app/cmd/app
(HP_CONFIG_FILE=config/config.local.toml HP_ENV=development HP_HTTP_SERVER_PORT=10099 HP_RUN_MODE=app \
  HP_DEV_MODE_FORCE_AGENT_LOCAL=true /tmp/hp-gs > /tmp/hp-gs.log 2>&1 &)
for p in dashboard-cert dismiss nope; do
  curl -s -o /dev/null -w "$p %{http_code}\n" -X POST localhost:10099/_/system/get-started/$p
done
```

Expected: `dashboard-cert 401`, `dismiss 401`, `nope 404`. Leave the backend running for Task 7.

- [ ] **Step 4: Run the tests to see them pass, then the whole backend**

Run: `go test ./hivepaas_app/usecase/system/getstarteduc/`
Expected: `ok` for `getstarteduc`

```bash
go build ./... && golangci-lint run ./... && go test ./...
```

Expected: the build passes, `0 issues.`, and every package `ok`.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app docs/openapi/swagger.json
git commit -m "feat(get-started): ask again for the dashboard's certificate, and close the card" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 6: The dashboard reads the checklist and calls the endpoints

In `../hivepaas-dashboard`. The dashboard has no unit-test runner; its gate is the type checker and the linter, and Task 7's browser check exercises this code end to end.

**Files:**
- Create: `src/application/modules/home/api/hooks/use-get-started.api.ts`
- Create: `src/application/modules/home/api/services/get-started-services/get-started.api.contracts.ts`
- Create: `src/application/modules/home/api/services/get-started-services/get-started.api.ts`
- Create: `src/application/modules/home/api/services/get-started-services/get-started.api.validator.ts`
- Create: `src/application/modules/home/api/services/get-started-services/index.ts`
- Create: `src/application/modules/home/data/commands/get-started.commands.ts`
- Create: `src/application/modules/home/data/commands/index.ts`
- Create: `src/application/shared/api/services/session-services/setup-checklist.api.schema.ts`
- Create: `src/application/shared/entities/setup-checklist/index.ts`
- Create: `src/application/shared/entities/setup-checklist/setup-checklist.entity.ts`
- Modify: `src/application/modules/home/api/api-context/home.api.context.ts`
- Modify: `src/application/modules/home/api/hooks/index.ts`
- Modify: `src/application/modules/home/api/services/index.ts`
- Modify: `src/application/modules/home/data/index.ts`
- Modify: `src/application/shared/api/services/session-services/index.ts`
- Modify: `src/application/shared/api/services/session-services/session.api.contracts.ts`
- Modify: `src/application/shared/api/services/session-services/session.api.validator.ts`
- Modify: `src/application/shared/entities/index.ts`

**Interfaces:**
- Consumes: the wire from Tasks 4 and 5.
- Produces: `SetupChecklist`, `SetupChecklistItem`, `SetupChecklistItemStatus` (`@application/shared/entities`); `SetupChecklistSchema`, `SetupChecklistItemSchema` (`@application/shared/api/services`); `Session_GetProfile_Res` data gains `setupChecklist: SetupChecklist | null`; `GetStartedApi` (`home.getStarted` in `HomeApiContext`), `useGetStartedApi`; `GetStartedCommands.useRequestDashboardCert()` and `.useDismiss()`, which read the profile again (`QK["session.get-profile"]`) on success.

- [ ] **Step 1: Branch**

```bash
cd ../hivepaas-dashboard && git checkout main && git checkout -b feat/get-started
```

- [ ] **Step 2: Write the code** - save as `/tmp/gs-d1.diff` and `git apply` it:

```diff
diff --git a/src/application/modules/home/api/api-context/home.api.context.ts b/src/application/modules/home/api/api-context/home.api.context.ts
index a158400c..208cd7f6 100644
--- a/src/application/modules/home/api/api-context/home.api.context.ts
+++ b/src/application/modules/home/api/api-context/home.api.context.ts
@@ -1,11 +1,12 @@
 import { createContext } from "react";
 
-import { HomeAttentionApi, HomeAttentionApiValidator } from "../services";
+import { GetStartedApi, GetStartedApiValidator, HomeAttentionApi, HomeAttentionApiValidator } from "../services";
 
 function createApi() {
     return {
         home: {
             attention: new HomeAttentionApi(new HomeAttentionApiValidator()),
+            getStarted: new GetStartedApi(new GetStartedApiValidator()),
         },
     };
 }
diff --git a/src/application/modules/home/api/hooks/index.ts b/src/application/modules/home/api/hooks/index.ts
index 5fbd9383..a149060a 100644
--- a/src/application/modules/home/api/hooks/index.ts
+++ b/src/application/modules/home/api/hooks/index.ts
@@ -1 +1,2 @@
+export * from "./use-get-started.api";
 export * from "./use-home-attention.api";
diff --git a/src/application/modules/home/api/hooks/use-get-started.api.ts b/src/application/modules/home/api/hooks/use-get-started.api.ts
new file mode 100644
index 00000000..b06c7eaf
--- /dev/null
+++ b/src/application/modules/home/api/hooks/use-get-started.api.ts
@@ -0,0 +1,45 @@
+import { use, useMemo } from "react";
+
+import { match } from "oxide.ts";
+import { HomeApiContext } from "~/home/api/api-context";
+
+import { useApiErrorNotifications } from "@infrastructure/api";
+
+function createHook() {
+    return function useGetStartedApi() {
+        const { api } = use(HomeApiContext);
+        const { notifyError } = useApiErrorNotifications();
+
+        const mutations = useMemo(
+            () => ({
+                requestDashboardCert: async () => {
+                    const result = await api.home.getStarted.requestDashboardCert({ data: {} });
+
+                    return match(result, {
+                        Ok: _ => _,
+                        Err: error => {
+                            notifyError({ message: "Failed to ask for the dashboard's certificate", error });
+                            throw error;
+                        },
+                    });
+                },
+                dismiss: async () => {
+                    const result = await api.home.getStarted.dismiss({ data: {} });
+
+                    return match(result, {
+                        Ok: _ => _,
+                        Err: error => {
+                            notifyError({ message: "Failed to close Get started", error });
+                            throw error;
+                        },
+                    });
+                },
+            }),
+            [api, notifyError],
+        );
+
+        return { mutations };
+    };
+}
+
+export const useGetStartedApi = createHook();
diff --git a/src/application/modules/home/api/services/get-started-services/get-started.api.contracts.ts b/src/application/modules/home/api/services/get-started-services/get-started.api.contracts.ts
new file mode 100644
index 00000000..bd11bfb5
--- /dev/null
+++ b/src/application/modules/home/api/services/get-started-services/get-started.api.contracts.ts
@@ -0,0 +1,9 @@
+import type { SetupChecklistItem } from "@application/shared/entities";
+
+import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";
+
+export type GetStarted_RequestDashboardCert_Req = ApiRequestBase<Record<string, never>>;
+export type GetStarted_RequestDashboardCert_Res = ApiResponseBase<SetupChecklistItem>;
+
+export type GetStarted_Dismiss_Req = ApiRequestBase<Record<string, never>>;
+export type GetStarted_Dismiss_Res = ApiResponseBase<{ type: "success" }>;
diff --git a/src/application/modules/home/api/services/get-started-services/get-started.api.ts b/src/application/modules/home/api/services/get-started-services/get-started.api.ts
new file mode 100644
index 00000000..2db3ebae
--- /dev/null
+++ b/src/application/modules/home/api/services/get-started-services/get-started.api.ts
@@ -0,0 +1,44 @@
+import { Err, Ok, type Result } from "oxide.ts";
+import { catchError, from, lastValueFrom, map, of } from "rxjs";
+
+import { BaseApi, parseApiError } from "@infrastructure/api";
+
+import type {
+    GetStarted_Dismiss_Req,
+    GetStarted_Dismiss_Res,
+    GetStarted_RequestDashboardCert_Req,
+    GetStarted_RequestDashboardCert_Res,
+} from "./get-started.api.contracts";
+import type { GetStartedApiValidator } from "./get-started.api.validator";
+
+export class GetStartedApi extends BaseApi {
+    public constructor(private readonly validator: GetStartedApiValidator) {
+        super();
+    }
+
+    async requestDashboardCert(
+        _request: GetStarted_RequestDashboardCert_Req,
+        signal?: AbortSignal,
+    ): Promise<Result<GetStarted_RequestDashboardCert_Res, Error>> {
+        return lastValueFrom(
+            from(this.client.v1.post("/system/get-started/dashboard-cert", {}, { signal })).pipe(
+                map(this.validator.requestDashboardCert),
+                map(res => Ok(res)),
+                catchError(error => of(Err(parseApiError(error)))),
+            ),
+        );
+    }
+
+    async dismiss(
+        _request: GetStarted_Dismiss_Req,
+        signal?: AbortSignal,
+    ): Promise<Result<GetStarted_Dismiss_Res, Error>> {
+        return lastValueFrom(
+            from(this.client.v1.post("/system/get-started/dismiss", {}, { signal })).pipe(
+                map(this.validator.dismiss),
+                map(res => Ok(res)),
+                catchError(error => of(Err(parseApiError(error)))),
+            ),
+        );
+    }
+}
diff --git a/src/application/modules/home/api/services/get-started-services/get-started.api.validator.ts b/src/application/modules/home/api/services/get-started-services/get-started.api.validator.ts
new file mode 100644
index 00000000..71ad3050
--- /dev/null
+++ b/src/application/modules/home/api/services/get-started-services/get-started.api.validator.ts
@@ -0,0 +1,24 @@
+import { type AxiosResponse } from "axios";
+import { z } from "zod";
+
+import { SetupChecklistItemSchema } from "@application/shared/api/services";
+
+import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";
+
+import type { GetStarted_Dismiss_Res, GetStarted_RequestDashboardCert_Res } from "./get-started.api.contracts";
+
+const RequestDashboardCertSchema = z.object({
+    data: SetupChecklistItemSchema,
+    meta: BaseMetaApiSchema.nullish(),
+});
+
+export class GetStartedApiValidator {
+    requestDashboardCert = (response: AxiosResponse): GetStarted_RequestDashboardCert_Res => {
+        const { data, meta } = parseApiResponse({ response, schema: RequestDashboardCertSchema });
+        return { data, meta };
+    };
+
+    dismiss = (_: AxiosResponse): GetStarted_Dismiss_Res => {
+        return { data: { type: "success" } };
+    };
+}
diff --git a/src/application/modules/home/api/services/get-started-services/index.ts b/src/application/modules/home/api/services/get-started-services/index.ts
new file mode 100644
index 00000000..122a0dff
--- /dev/null
+++ b/src/application/modules/home/api/services/get-started-services/index.ts
@@ -0,0 +1,3 @@
+export * from "./get-started.api.contracts";
+export * from "./get-started.api.validator";
+export * from "./get-started.api";
diff --git a/src/application/modules/home/api/services/index.ts b/src/application/modules/home/api/services/index.ts
index 59f2d052..ca913b59 100644
--- a/src/application/modules/home/api/services/index.ts
+++ b/src/application/modules/home/api/services/index.ts
@@ -1 +1,2 @@
+export * from "./get-started-services";
 export * from "./home-attention-services";
diff --git a/src/application/modules/home/data/commands/get-started.commands.ts b/src/application/modules/home/data/commands/get-started.commands.ts
new file mode 100644
index 00000000..4f163788
--- /dev/null
+++ b/src/application/modules/home/data/commands/get-started.commands.ts
@@ -0,0 +1,51 @@
+import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
+import { useGetStartedApi } from "~/home/api";
+import type { GetStarted_Dismiss_Res, GetStarted_RequestDashboardCert_Res } from "~/home/api/services";
+
+import { QK } from "@application/shared/data/constants";
+
+/**
+ * The checklist comes with the profile, so after either call the profile is
+ * read again: it has the certificate being obtained, or no checklist at all.
+ */
+function useRefreshProfile() {
+    const queryClient = useQueryClient();
+
+    return () => queryClient.invalidateQueries({ queryKey: [QK["session.get-profile"]] });
+}
+
+function useRequestDashboardCert({
+    onSuccess,
+    ...options
+}: Omit<UseMutationOptions<GetStarted_RequestDashboardCert_Res>, "mutationFn"> = {}) {
+    const { mutations } = useGetStartedApi();
+    const refreshProfile = useRefreshProfile();
+
+    return useMutation({
+        mutationFn: mutations.requestDashboardCert,
+        onSuccess: (response, ...rest) => {
+            void refreshProfile();
+            onSuccess?.(response, ...rest);
+        },
+        ...options,
+    });
+}
+
+function useDismiss({ onSuccess, ...options }: Omit<UseMutationOptions<GetStarted_Dismiss_Res>, "mutationFn"> = {}) {
+    const { mutations } = useGetStartedApi();
+    const refreshProfile = useRefreshProfile();
+
+    return useMutation({
+        mutationFn: mutations.dismiss,
+        onSuccess: (response, ...rest) => {
+            void refreshProfile();
+            onSuccess?.(response, ...rest);
+        },
+        ...options,
+    });
+}
+
+export const GetStartedCommands = Object.freeze({
+    useRequestDashboardCert,
+    useDismiss,
+});
diff --git a/src/application/modules/home/data/commands/index.ts b/src/application/modules/home/data/commands/index.ts
new file mode 100644
index 00000000..095eaf49
--- /dev/null
+++ b/src/application/modules/home/data/commands/index.ts
@@ -0,0 +1 @@
+export * from "./get-started.commands";
diff --git a/src/application/modules/home/data/index.ts b/src/application/modules/home/data/index.ts
index 9a67dd7f..3fd7eb0e 100644
--- a/src/application/modules/home/data/index.ts
+++ b/src/application/modules/home/data/index.ts
@@ -1,2 +1,3 @@
+export * from "./commands";
 export * from "./constants";
 export * from "./queries";
diff --git a/src/application/shared/api/services/session-services/index.ts b/src/application/shared/api/services/session-services/index.ts
index 2b69ac9e..af9f4f66 100644
--- a/src/application/shared/api/services/session-services/index.ts
+++ b/src/application/shared/api/services/session-services/index.ts
@@ -1,3 +1,4 @@
 export * from "./session.api.contracts";
 export * from "./session.api";
 export * from "./session.api.validator";
+export * from "./setup-checklist.api.schema";
diff --git a/src/application/shared/api/services/session-services/session.api.contracts.ts b/src/application/shared/api/services/session-services/session.api.contracts.ts
index 3e4fab54..9ad139be 100644
--- a/src/application/shared/api/services/session-services/session.api.contracts.ts
+++ b/src/application/shared/api/services/session-services/session.api.contracts.ts
@@ -1,11 +1,14 @@
-import { type Profile } from "@application/shared/entities";
+import { type Profile, type SetupChecklist } from "@application/shared/entities";
 
 import { type ApiResponseBase } from "@infrastructure/api";
 
 /**
- * Get profile
+ * Get profile. `setupChecklist` is what the installation still has to do, given
+ * to an admin while `nextStep` is `hivepaas/get-started`.
  */
-export type Session_GetProfile_Res = ApiResponseBase<Profile & { nextStep?: string }>;
+export type Session_GetProfile_Res = ApiResponseBase<
+    Profile & { nextStep?: string; setupChecklist: SetupChecklist | null }
+>;
 
 /**
  * Logout
diff --git a/src/application/shared/api/services/session-services/session.api.validator.ts b/src/application/shared/api/services/session-services/session.api.validator.ts
index 7d666050..d4fa21d6 100644
--- a/src/application/shared/api/services/session-services/session.api.validator.ts
+++ b/src/application/shared/api/services/session-services/session.api.validator.ts
@@ -8,6 +8,8 @@ import type { ModulePermission, ProjectPermission } from "@application/shared/pe
 
 import { parseApiResponse } from "@infrastructure/api";
 
+import { SetupChecklistSchema } from "./setup-checklist.api.schema";
+
 /**
  * Get account API response schema
  */
@@ -119,6 +121,7 @@ function mapProjectAccessesToProjectPermissions(
 const GetProfileSchema = z.object({
     data: z.object({
         nextStep: z.string().optional(),
+        setupChecklist: SetupChecklistSchema.nullish(),
         user: z.object({
             id: z.string(),
             username: z.string(),
@@ -148,7 +151,7 @@ export class SessionApiValidator {
      */
     getProfile = (response: AxiosResponse): Session_GetProfile_Res => {
         const {
-            data: { user, nextStep },
+            data: { user, nextStep, setupChecklist },
         } = parseApiResponse({
             response,
             schema: GetProfileSchema,
@@ -172,6 +175,7 @@ export class SessionApiValidator {
                 createdAt: user.createdAt,
                 lastAccess: user.lastAccess ?? null,
                 nextStep,
+                setupChecklist: setupChecklist ?? null,
                 position: user.position ?? "",
                 status: user.status,
                 projectAccesses,
diff --git a/src/application/shared/api/services/session-services/setup-checklist.api.schema.ts b/src/application/shared/api/services/session-services/setup-checklist.api.schema.ts
new file mode 100644
index 00000000..ea29f93c
--- /dev/null
+++ b/src/application/shared/api/services/session-services/setup-checklist.api.schema.ts
@@ -0,0 +1,25 @@
+import { z } from "zod";
+
+import type { SetupChecklist, SetupChecklistItem } from "@application/shared/entities";
+
+export const SetupChecklistItemSchema = z
+    .object({
+        status: z.enum(["todo", "obtaining", "failed", "done"]).catch("todo"),
+        domain: z.string().nullish(),
+        error: z.string().nullish(),
+    })
+    .transform(
+        (item): SetupChecklistItem => ({
+            status: item.status,
+            domain: item.domain ?? "",
+            error: item.error ?? "",
+        }),
+    );
+
+export const SetupChecklistSchema = z
+    .object({
+        dashboardCert: SetupChecklistItemSchema,
+        twoFactor: SetupChecklistItemSchema,
+        githubApp: SetupChecklistItemSchema,
+    })
+    .transform((checklist): SetupChecklist => checklist);
diff --git a/src/application/shared/entities/index.ts b/src/application/shared/entities/index.ts
index 208c5a74..b84e8432 100644
--- a/src/application/shared/entities/index.ts
+++ b/src/application/shared/entities/index.ts
@@ -1,2 +1,3 @@
 export * from "./profile";
 export * from "./public";
+export * from "./setup-checklist";
diff --git a/src/application/shared/entities/setup-checklist/index.ts b/src/application/shared/entities/setup-checklist/index.ts
new file mode 100644
index 00000000..5224b05a
--- /dev/null
+++ b/src/application/shared/entities/setup-checklist/index.ts
@@ -0,0 +1 @@
+export * from "./setup-checklist.entity";
diff --git a/src/application/shared/entities/setup-checklist/setup-checklist.entity.ts b/src/application/shared/entities/setup-checklist/setup-checklist.entity.ts
new file mode 100644
index 00000000..438b7ece
--- /dev/null
+++ b/src/application/shared/entities/setup-checklist/setup-checklist.entity.ts
@@ -0,0 +1,23 @@
+/**
+ * Where one thing a new installation still has to do stands, worked out by the
+ * server from what exists.
+ */
+export type SetupChecklistItemStatus = "todo" | "obtaining" | "failed" | "done";
+
+export interface SetupChecklistItem {
+    status: SetupChecklistItemStatus;
+    /** The dashboard's domain, for the certificate item. */
+    domain: string;
+    /** Why the last attempt failed, for the certificate item. */
+    error: string;
+}
+
+/**
+ * What the Get started card lists, for an admin while the installation step is
+ * `hivepaas/get-started`.
+ */
+export interface SetupChecklist {
+    dashboardCert: SetupChecklistItem;
+    twoFactor: SetupChecklistItem;
+    githubApp: SetupChecklistItem;
+}
```

- [ ] **Step 3: Check it**

```bash
npx tsc --noEmit && npm run lint && npx prettier --check src
```

Expected: no output from `tsc`, no problems from `eslint`, `All matched files use Prettier code style!`

- [ ] **Step 4: Commit**

```bash
git add src
git commit -m "feat(home): read the Get started checklist, and ask for the dashboard's certificate or close it" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 7: The Get started card

**Files:**
- Create: `src/application/modules/home/routes/home/building-blocks/get-started.card.com.tsx`
- Modify: `src/application/modules/home/routes/home/building-blocks/index.ts`
- Modify: `src/application/modules/home/routes/home/route/home.route.com.tsx`

**Interfaces:**
- Consumes: `SessionQueries.useGetProfile` (its `session.get-profile` query is the one the 2FA dialog reads again on success, so the card follows it); `GetStartedCommands` (Task 6); `useF2aSetupDialog` (shared, already mounted by the shared dialogs container); `useProvisionGithubAppDialog` and `ProvisionGithubAppDialog` (settings module - the home page has no dialogs container, so the card renders the dialog itself).
- Produces: `GetStartedCard`, first on the home page, for an admin whose profile has `setupChecklist`.

- [ ] **Step 1: Write the code** - save as `/tmp/gs-d2.diff` and `git apply` it:

```diff
diff --git a/src/application/modules/home/routes/home/building-blocks/get-started.card.com.tsx b/src/application/modules/home/routes/home/building-blocks/get-started.card.com.tsx
new file mode 100644
index 00000000..4c782924
--- /dev/null
+++ b/src/application/modules/home/routes/home/building-blocks/get-started.card.com.tsx
@@ -0,0 +1,263 @@
+import { useEffect, useRef } from "react";
+
+import { CircleCheck, CircleX, KeyRound, Loader2, ShieldCheck, X } from "lucide-react";
+import { toast } from "sonner";
+import { GetStartedCommands } from "~/home/data";
+import { ProvisionGithubAppDialog, useProvisionGithubAppDialog } from "~/settings/dialogs/provision-github-app";
+
+import { useProfileContext } from "@application/shared/context";
+import { SessionQueries } from "@application/shared/data/queries";
+import { useF2aSetupDialog } from "@application/shared/dialogs";
+import type { SetupChecklist, SetupChecklistItem } from "@application/shared/entities";
+import { EUserRole } from "@application/shared/enums";
+
+import { Badge } from "@/components/ui/badge";
+import { Button } from "@/components/ui/button";
+import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
+
+/** How often the profile is read again while the certificate is being obtained. */
+const OBTAINING_REFRESH_MS = 5_000;
+
+const REOPEN_BROWSER =
+    "Certificate installed. Quit and reopen your browser to see this site as secure: " +
+    "a browser keeps the connection it opened with the old certificate, even in a new tab.";
+
+function StatusIcon({ item }: { item: SetupChecklistItem }) {
+    switch (item.status) {
+        case "done":
+            return <CircleCheck className="mt-0.5 size-4 shrink-0 text-green-600 dark:text-green-500" />;
+        case "failed":
+            return <CircleX className="mt-0.5 size-4 shrink-0 text-destructive" />;
+        case "obtaining":
+            return <Loader2 className="mt-0.5 size-4 shrink-0 animate-spin text-muted-foreground" />;
+        default:
+            return <span className="mt-1 size-3.5 shrink-0 rounded-full border-2 border-muted-foreground/50" />;
+    }
+}
+
+interface RowProps {
+    item: SetupChecklistItem;
+    title: string;
+    tag?: string;
+    children: React.ReactNode;
+    action?: React.ReactNode;
+}
+
+function Row({ item, title, tag, children, action }: RowProps) {
+    return (
+        <li className="flex items-start gap-3 px-5 py-3.5 border-b border-border/60 last:border-b-0">
+            <StatusIcon item={item} />
+            {/* The button goes under the text on a phone, beside it on anything wider. */}
+            <div className="flex min-w-0 grow flex-col gap-2 sm:flex-row sm:items-start sm:gap-3">
+                <div className="flex min-w-0 grow flex-col gap-1">
+                    <span className="text-sm font-semibold">
+                        {title}
+                        {tag && <span className="font-normal text-muted-foreground"> · {tag}</span>}
+                    </span>
+                    <div className="text-[13px] text-muted-foreground break-words">{children}</div>
+                </div>
+                {action && item.status !== "done" && <div className="shrink-0">{action}</div>}
+            </div>
+        </li>
+    );
+}
+
+function DashboardCertRow({ item }: { item: SetupChecklistItem }) {
+    const { mutate: requestCert, isPending } = GetStartedCommands.useRequestDashboardCert();
+    const domain = item.domain || "the dashboard's domain";
+
+    let detail: React.ReactNode;
+    switch (item.status) {
+        case "done":
+            detail = REOPEN_BROWSER;
+            break;
+        case "obtaining":
+            detail = `Getting a certificate for ${domain} from Let's Encrypt. This takes a minute or two.`;
+            break;
+        case "failed":
+            detail = (
+                <>
+                    <span className="text-destructive">The last attempt failed: {item.error}</span>
+                    <br />
+                    {`Check that ${domain} points at this server's IP and that port 80 is open, then try again.`}
+                </>
+            );
+            break;
+        default:
+            detail = `The dashboard still uses a self-signed certificate, which browsers warn about. For one from Let's Encrypt, ${domain} has to point at this server's IP and port 80 has to be open.`;
+    }
+
+    return (
+        <Row
+            item={item}
+            title="Secure the dashboard"
+            action={
+                <Button
+                    size="sm"
+                    variant="outline"
+                    isLoading={isPending}
+                    disabled={isPending || item.status === "obtaining"}
+                    onClick={() => {
+                        requestCert();
+                    }}
+                >
+                    <ShieldCheck className="size-4" />
+                    {item.status === "failed" ? "Try again" : "Get the certificate"}
+                </Button>
+            }
+        >
+            {detail}
+        </Row>
+    );
+}
+
+function TwoFactorRow({ item }: { item: SetupChecklistItem }) {
+    const dialog = useF2aSetupDialog({
+        onClose: () => {
+            dialog.actions.close();
+        },
+    });
+
+    return (
+        <Row
+            item={item}
+            title="Turn on two-factor authentication"
+            tag="Recommended"
+            action={
+                <Button
+                    size="sm"
+                    variant="outline"
+                    onClick={dialog.actions.open}
+                >
+                    <KeyRound className="size-4" />
+                    Set up
+                </Button>
+            }
+        >
+            {item.status === "done"
+                ? "Your account asks for a code from your authenticator app at sign-in."
+                : "Ask for a code from an authenticator app at sign-in, on top of your password."}
+        </Row>
+    );
+}
+
+function GithubAppRow({ item, certDone }: { item: SetupChecklistItem; certDone: boolean }) {
+    const dialog = useProvisionGithubAppDialog();
+
+    return (
+        <Row
+            item={item}
+            title="Connect a GitHub App"
+            tag="Optional"
+            action={
+                <Button
+                    size="sm"
+                    variant="outline"
+                    onClick={() => {
+                        dialog.actions.open({ type: "settings" });
+                    }}
+                >
+                    Connect
+                </Button>
+            }
+        >
+            {item.status === "done" ? (
+                "A GitHub App is connected."
+            ) : (
+                <>
+                    Sign in with GitHub, and create apps from your repositories.
+                    {!certDone && (
+                        <>
+                            {" "}
+                            <span className="text-amber-600 dark:text-amber-400">
+                                GitHub sends its webhooks only to a site with a valid certificate: secure the dashboard
+                                first.
+                            </span>
+                        </>
+                    )}
+                </>
+            )}
+        </Row>
+    );
+}
+
+/**
+ * Tells the admin when the certificate arrives while they watch, since the row
+ * saying so goes with the card once everything is done.
+ */
+function useCertArrivedToast(checklist: SetupChecklist | null) {
+    const previous = useRef(checklist?.dashboardCert.status);
+    const status = checklist?.dashboardCert.status;
+
+    useEffect(() => {
+        if (previous.current === "obtaining" && status === "done") {
+            toast.success(REOPEN_BROWSER, { duration: 15_000 });
+        }
+        previous.current = status;
+    }, [status]);
+}
+
+/**
+ * What a new installation still has to do, for an admin, until it is done or
+ * closed. The server works out each item from what exists, and clears the step
+ * for every admin when all three are done or one of them closes the card.
+ */
+export function GetStartedCard() {
+    const isAdmin = useProfileContext(state => state.profile?.role === EUserRole.Admin);
+
+    const { data } = SessionQueries.useGetProfile({
+        enabled: isAdmin,
+        refetchInterval: query =>
+            query.state.data?.data.setupChecklist?.dashboardCert.status === "obtaining" ? OBTAINING_REFRESH_MS : false,
+    });
+    const { mutate: dismiss, isPending: isDismissing } = GetStartedCommands.useDismiss();
+
+    const checklist = isAdmin ? (data?.data.setupChecklist ?? null) : null;
+    useCertArrivedToast(checklist);
+
+    if (!checklist) {
+        return null;
+    }
+
+    const left = [checklist.dashboardCert, checklist.twoFactor, checklist.githubApp].filter(
+        item => item.status !== "done",
+    ).length;
+
+    return (
+        <Card className="gap-0 py-0">
+            <CardHeader className="flex flex-row items-center gap-2 border-b px-5 py-4 [.border-b]:pb-4">
+                <CardTitle className="text-[15px]">Get started</CardTitle>
+                <Badge
+                    variant="secondary"
+                    className="rounded-full px-2"
+                >
+                    {left} left
+                </Badge>
+                <Button
+                    size="icon-sm"
+                    variant="ghost"
+                    className="ml-auto"
+                    aria-label="Close Get started"
+                    title="Close for every admin"
+                    disabled={isDismissing}
+                    onClick={() => {
+                        dismiss();
+                    }}
+                >
+                    <X className="size-4" />
+                </Button>
+            </CardHeader>
+            <CardContent className="px-0">
+                <ul>
+                    <DashboardCertRow item={checklist.dashboardCert} />
+                    <TwoFactorRow item={checklist.twoFactor} />
+                    <GithubAppRow
+                        item={checklist.githubApp}
+                        certDone={checklist.dashboardCert.status === "done"}
+                    />
+                </ul>
+            </CardContent>
+            <ProvisionGithubAppDialog />
+        </Card>
+    );
+}
diff --git a/src/application/modules/home/routes/home/building-blocks/index.ts b/src/application/modules/home/routes/home/building-blocks/index.ts
index 552de8e2..9b14b4a3 100644
--- a/src/application/modules/home/routes/home/building-blocks/index.ts
+++ b/src/application/modules/home/routes/home/building-blocks/index.ts
@@ -1,3 +1,4 @@
+export * from "./get-started.card.com";
 export * from "./needs-attention.card.com";
 export * from "./nodes.card.com";
 export * from "./recent-tasks.card.com";
diff --git a/src/application/modules/home/routes/home/route/home.route.com.tsx b/src/application/modules/home/routes/home/route/home.route.com.tsx
index ac1f8f3a..2b2d61fe 100644
--- a/src/application/modules/home/routes/home/route/home.route.com.tsx
+++ b/src/application/modules/home/routes/home/route/home.route.com.tsx
@@ -8,7 +8,14 @@ import { DEFAULT_PAGINATED_DATA, MODULE_IDS, ROUTE } from "@application/shared/c
 import { EUserRole } from "@application/shared/enums";
 import { useConditionalModule, useProjectPermissionsStore } from "@application/shared/permissions";
 
-import { NeedsAttentionCard, NodesCard, RecentTasksCard, SummaryTiles, UpdateAvailableBadge } from "../building-blocks";
+import {
+    GetStartedCard,
+    NeedsAttentionCard,
+    NodesCard,
+    RecentTasksCard,
+    SummaryTiles,
+    UpdateAvailableBadge,
+} from "../building-blocks";
 
 function plural(count: number, noun: string) {
     return `${count} ${noun}${count === 1 ? "" : "s"}`;
@@ -104,6 +111,8 @@ export function HomeRoute() {
                 <UpdateAvailableBadge />
             </header>
 
+            <GetStartedCard />
+
             <SummaryTiles tiles={tiles} />
 
             <div
```

- [ ] **Step 2: Check it**

```bash
npx tsc --noEmit && npm run lint && npx prettier --check src
```

Expected: as in Task 6.

- [ ] **Step 3: See it work in a browser**

The backend from Task 5 is on 10099. The check answers `GET /sessions/me` with the real profile plus a checklist it changes as it goes, and answers the two POSTs and the 2FA setup call itself, so nothing is written to the database. It needs a token for an admin (the dev helper), the dashboard on 4322 talking to its own origin, and a headless Chrome.

```bash
SCR=$(mktemp -d)
ADMIN=$(PGPASSWORD=abc123 psql -h localhost -p 35432 -U hivepaas -d hivepaas -tAc \
  "SELECT id FROM users WHERE deleted_at IS NULL AND role='admin' ORDER BY created_at LIMIT 1")
curl -sS -u hivepaas:abc123 -X POST "http://localhost:10099/_/internal/dev-helper/dev-mode-login?userId=$ADMIN" \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["accessToken"])' > $SCR/token.txt
(PORT=4322 VITE_HP_DASHBOARD_BASE_URL= VITE_HP_API_PROXY_TARGET=http://localhost:10099 \
  npx vite --port 4322 --strictPort > $SCR/vite.log 2>&1 &)
("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --headless=new --remote-debugging-port=9333 \
  --user-data-dir=$SCR/chrome --no-first-run --window-size=1400,1200 about:blank > $SCR/chrome.log 2>&1 &)
```

Save these two files in `$SCR` (they are not committed):

`$SCR/cdp-lib.mjs`:

```js
export async function connect() {
    const targets = await (await fetch("http://localhost:9333/json")).json();
    const page = targets.find(t => t.type === "page");
    const ws = new WebSocket(page.webSocketDebuggerUrl);
    await new Promise(r => ws.addEventListener("open", r, { once: true }));
    let id = 0;
    const pending = new Map();
    const handlers = [];
    const logs = [];
    ws.addEventListener("message", e => {
        const m = JSON.parse(e.data);
        if (m.id && pending.has(m.id)) { pending.get(m.id)(m); pending.delete(m.id); return; }
        if (m.method === "Runtime.exceptionThrown") logs.push(m.params.exceptionDetails.text + " " + (m.params.exceptionDetails.exception?.description ?? ""));
        if (m.method === "Runtime.consoleAPICalled" && m.params.type === "error") logs.push(m.params.args.map(a => a.value ?? a.description).join(" "));
        for (const h of handlers) h(m);
    });
    const send = (method, params = {}) => new Promise(r => { const i = ++id; pending.set(i, r); ws.send(JSON.stringify({ id: i, method, params })); });
    const ev = async expr => (await send("Runtime.evaluate", { expression: expr, returnByValue: true, awaitPromise: true })).result?.result?.value;
    const sleep = ms => new Promise(r => setTimeout(r, ms));
    await send("Runtime.enable");
    return { ws, send, ev, sleep, logs, on: h => handlers.push(h) };
}
```

`$SCR/ui-get-started.mjs`:

```js
import { connect } from "./cdp-lib.mjs";
import { readFileSync, writeFileSync } from "node:fs";
const SP = process.env.SP;
const token = readFileSync(SP + "/token.txt", "utf8").trim();
const { send, ev, sleep, logs, on, ws } = await connect();
await send("Emulation.setDeviceMetricsOverride", { width: 1400, height: 1200, deviceScaleFactor: 1, mobile: false });

const item = (status, extra = {}) => ({ status, ...extra });
let checklist = {
    dashboardCert: item("todo", { domain: "dash.example.com" }),
    twoFactor: item("todo"),
    githubApp: item("todo"),
};
let meCalls = 0;
const posts = [];
const b64 = s => Buffer.from(s).toString("base64");

on(async m => {
    if (m.method !== "Fetch.requestPaused") return;
    const { requestId, request, responseStatusCode } = m.params;
    const url = request.url;
    if (url.includes("/sessions/me") && responseStatusCode !== undefined) {
        meCalls++;
        const body = await send("Fetch.getResponseBody", { requestId });
        const raw = body.result.base64Encoded ? Buffer.from(body.result.body, "base64").toString() : body.result.body;
        const json = JSON.parse(raw);
        if (checklist) { json.data.nextStep = "hivepaas/get-started"; json.data.setupChecklist = checklist; }
        else { delete json.data.nextStep; }
        await send("Fetch.fulfillRequest", { requestId, responseCode: 200,
            responseHeaders: [{ name: "Content-Type", value: "application/json" }], body: b64(JSON.stringify(json)) });
        return;
    }
    if (url.includes("/system/get-started/")) {
        posts.push(request.method + " " + url.replace(/.*\/_/, ""));
        let resp = { meta: {} };
        if (url.endsWith("/dashboard-cert")) {
            checklist.dashboardCert = item("obtaining", { domain: "dash.example.com" });
            resp = { meta: {}, data: checklist.dashboardCert };
        } else {
            checklist = null;
        }
        await send("Fetch.fulfillRequest", { requestId, responseCode: 200,
            responseHeaders: [{ name: "Content-Type", value: "application/json" }], body: b64(JSON.stringify(resp)) });
        return;
    }
    if (url.includes("/mfa/totp-begin-setup")) {
        posts.push("mfa begin");
        await send("Fetch.fulfillRequest", { requestId, responseCode: 200,
            responseHeaders: [{ name: "Content-Type", value: "application/json" }],
            body: b64(JSON.stringify({ meta: {}, data: { totpToken: "t", secret: "ABCDEF", qrCode: { dataBase64: "" } } })) });
        return;
    }
    await send("Fetch.continueRequest", { requestId });
});
await send("Fetch.enable", { patterns: [
    { urlPattern: "*/_/sessions/me*", requestStage: "Response" },
    { urlPattern: "*/_/system/get-started/*", requestStage: "Request" },
    { urlPattern: "*/_/users/current/mfa/*", requestStage: "Request" },
] });

const shot = name => send("Page.captureScreenshot", { format: "png" }).then(s => writeFileSync(`${SP}/${name}.png`, Buffer.from(s.result.data, "base64")));
const card = () => ev(`(() => { const t = [...document.querySelectorAll('[data-slot=card-title], div')].find(e => e.childElementCount === 0 && e.textContent.trim() === 'Get started'); const c = t?.closest('[data-slot=card]') ?? t?.parentElement?.parentElement; return c ? c.innerText : null; })()`);
const clickButton = label => ev(`(() => { const b = [...document.querySelectorAll('button')].find(b => b.innerText.trim() === ${JSON.stringify(label)} || b.getAttribute('aria-label') === ${JSON.stringify(label)}); if (!b) return 'none'; if (b.disabled) return 'disabled'; b.click(); return 'ok'; })()`);

await send("Page.navigate", { url: "http://localhost:4322/home/" }); await sleep(1500);
await ev(`localStorage.setItem("token", ${JSON.stringify(token)})`);
await send("Page.navigate", { url: "http://localhost:4322/home/" }); await sleep(5000);
logs.length = 0; // what signing in logged is not the card's
console.log("1 url:", await ev("location.pathname"));
console.log("1 card:", JSON.stringify(await card()));
await shot("gs-1-todo");

console.log("2 click cert:", await clickButton("Get the certificate"));
await sleep(1500);
console.log("2 card:", JSON.stringify(await card()));
console.log("2 button:", await clickButton("Get the certificate"));
const before = meCalls; await sleep(11000);
console.log("2 polls in 11s:", meCalls - before);
await shot("gs-2-obtaining");

checklist.dashboardCert = item("done", { domain: "dash.example.com" });
await sleep(6000);
console.log("3 card:", JSON.stringify(await card()));
console.log("3 toast:", JSON.stringify(await ev(`[...document.querySelectorAll('[data-sonner-toast]')].map(t => t.innerText).join(' | ')`)));
const afterDone = meCalls; await sleep(11000);
console.log("3 polls after done in 11s:", meCalls - afterDone);
await shot("gs-3-done");

console.log("4 2fa:", await clickButton("Set up")); await sleep(1500);
console.log("4 dialog:", JSON.stringify(await ev(`[...document.querySelectorAll('[role=dialog]')].map(d => d.innerText.slice(0, 80)).join(' | ')`)));
await send("Input.dispatchKeyEvent", { type: "keyDown", key: "Escape", code: "Escape", windowsVirtualKeyCode: 27 });
await send("Input.dispatchKeyEvent", { type: "keyUp", key: "Escape", code: "Escape", windowsVirtualKeyCode: 27 });
await sleep(800);
console.log("4 github:", await clickButton("Connect")); await sleep(1500);
console.log("4 dialog:", JSON.stringify(await ev(`[...document.querySelectorAll('[role=dialog]')].map(d => d.innerText.slice(0, 80)).join(' | ')`)));
await shot("gs-4-github");
await send("Input.dispatchKeyEvent", { type: "keyDown", key: "Escape", code: "Escape", windowsVirtualKeyCode: 27 });
await send("Input.dispatchKeyEvent", { type: "keyUp", key: "Escape", code: "Escape", windowsVirtualKeyCode: 27 });
await sleep(800);

checklist.dashboardCert = item("failed", { domain: "dash.example.com", error: "acme: error: 400 :: urn:ietf:params:acme:error:dns :: NXDOMAIN looking up A for dash.example.com" });
await send("Page.reload"); await sleep(5000);
console.log("5 card:", JSON.stringify(await card()));
await shot("gs-5-failed");

console.log("6 close:", await clickButton("Close Get started")); await sleep(2500);
console.log("6 card:", JSON.stringify(await card()));
console.log("posts:", posts.join(", "));
console.log("errors:", logs.join(" / ") || "none");
ws.close();
```

Run: `cd $SCR && sleep 5 && SP=$SCR node ui-get-started.mjs`

Expected, in order (long lines cut here):

```
1 url: /home/
1 card: "Get started\n3 left\nSecure the dashboard\nThe dashboard still uses a self-signed certificate, ...
2 click cert: ok
2 card: "Get started\n3 left\nSecure the dashboard\nGetting a certificate for dash.example.com from Let's Encrypt. ...
2 button: disabled
2 polls in 11s: 2
3 card: "Get started\n2 left\nSecure the dashboard\nCertificate installed. Quit and reopen your browser ...
3 toast: "Certificate installed. Quit and reopen your browser to see this site as secure: ...
3 polls after done in 11s: 0
4 2fa: ok
4 dialog: "Activate 2FA\n\nSet up, change, or deactivate two-factor authentication.\n\nClose"
4 github: ok
4 dialog: "Provision Github app\nImportant: When you click begin, you will be redirected to "
5 card: "Get started\n3 left\nSecure the dashboard\nThe last attempt failed: acme: error: 400 :: ...
6 close: ok
6 card: null
posts: POST /system/get-started/dashboard-cert, mfa begin, POST /system/get-started/dismiss
errors: none
```

Look at `$SCR/gs-1-todo.png` and `$SCR/gs-5-failed.png`. At phone width (set `width: 390, ... mobile: true` in the script's `setDeviceMetricsOverride`) each row's button sits under its text and the page does not scroll sideways.

Then stop what the check started: `pkill -f "remote-debugging-port=9333"; pkill -f "vite --port 4322"; pkill -f /tmp/hp-gs`.

- [ ] **Step 4: Commit**

```bash
git add src
git commit -m "feat(home): the Get started card - secure the dashboard, two-factor authentication, a GitHub App" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### After the last task: review and merge

- [ ] Review the whole of both branches (`git diff main...feat/get-started` in each repo) against the spec, the Decisions beyond it, and the Review Focus.
- [ ] Merge each repo locally, never push:

```bash
git checkout main && git merge --no-ff feat/get-started -m "Merge branch 'feat/get-started'" && git branch -d feat/get-started
```

In `hivepaas`, run `go build ./... && go test ./...` on the merged `main`; in `hivepaas-dashboard`, `npx tsc --noEmit && npm run lint`.

- [ ] Tell the user the e2e of the installer (`deployment/release/install_e2e.sh`) is theirs to run on Linux: its domain resolves nowhere, so the dashboard comes up and `GetMe` says `dashboardCert: failed` once the task has given up (about 4 minutes).
