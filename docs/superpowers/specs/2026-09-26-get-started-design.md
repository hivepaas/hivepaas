# Get Started: the First Boot and the First Login

A fresh install leaves three things an admin should do and nothing that says so:
give the dashboard a real certificate, turn on two-factor authentication, and
connect a GitHub App. This design has HivePaaS try the first on its own at the
first boot, and show the three as a **Get started** card on the dashboard's home
page until they are done or the card is closed.

---

## Decisions

1. **The first boot asks for the dashboard's certificate itself.** The machinery
   exists - `EnsureCertsForDomains` and the `tasksslobtain` task, with the
   installation's default domain settings (`AutoObtain` on, Let's Encrypt, the
   admin's email) - but the dashboard's domain never goes through it: the
   HivePaaS app's routing is written straight to the database at the first boot.
   Where the domain already points at the server and port 80 is open, the
   certificate is there before anyone opens the dashboard.
2. **Nothing is forced.** Two-factor authentication stays a recommendation, and
   the first admin keeps `password-only`. The card never blocks the dashboard.
3. **One card for every admin.** The card is shown while the installation step is
   `hivepaas/get-started`, and closing it - or finishing all three - clears the
   step for everyone: certificates and GitHub Apps belong to the installation,
   not to one admin.
4. **Each item's state is worked out, not ticked.** The backend reads it from
   what exists, so an item done anywhere - the card, the settings pages - shows
   as done.
5. **A certificate obtained while the dashboard is open is not forced on the
   browser.** A browser keeps the connection it opened with the self-signed
   certificate, and a new tab reuses it; the card tells the admin to quit and
   reopen the browser. Restarting Traefik to close every connection is left out.

## 1. The installation step

`base.InstallationStepObtainAppSSL` (`hivepaas/obtain-ssl`) is replaced by
`base.InstallationStepGetStarted` (`hivepaas/get-started`): the step
`sysInstallationInitData` sets once the data is created. No backward
compatibility: an installation left at `hivepaas/obtain-ssl` shows no card.

## 2. The first boot

At the end of the first boot, after the transaction that creates the data has
committed, the app asks for the dashboard's certificate:

- the HivePaaS app is `hpAppService.LoadAppByKey(base.HivepaasAppKey)`; its
  routing settings' enabled domains without a certificate are the ones asked for;
- `domainService.EnsureCertsForDomains` with the app's scope, project and ID, and
  the tasks it returns scheduled on the task queue, as
  `routing_settings_update.go` does;
- a failure here is logged and does not stop the boot: the card offers the
  attempt again.

The task, when it succeeds, attaches the certificate to the domain and applies
the routing again, as it does for any app.

## 3. The API

**`GetMe`** gains `setupChecklist`, for an admin while the step is
`hivepaas/get-started`:

```json
"setupChecklist": {
  "dashboardCert": {"status": "failed", "domain": "hivepaas.example.com",
                    "error": "acme: ... NXDOMAIN ..."},
  "twoFactor":     {"status": "todo"},
  "githubApp":     {"status": "done"}
}
```

- **`dashboardCert`**, for the HivePaaS app's first enabled HTTP domain:
  - `done` - the domain names a certificate that has content, has not expired,
    and is not self-signed;
  - `obtaining` - a certificate setting for the domain exists with no content
    and no `lastError`: a task is on it;
  - `failed` - such a setting has a `lastError`, which is returned;
  - `todo` - none of these.
- **`twoFactor`**: `done` when the admin asking has a TOTP secret.
- **`githubApp`**: `done` when an active GitHub App setting exists.
- Absent for a non-admin, and once the step is cleared.

**`POST /system/get-started/dashboard-cert`** (admin): asks for the dashboard's
certificate now, as the first boot does, but past the wait a failure leaves
(`RetryAfter`, six hours): the admin asking is the reason to try. Refused with
`ERR_CONFLICT` while one is being obtained. Answers with the new `dashboardCert`.
`EnsureCertsReq` gains `IgnoreRetryAfter`.

**`POST /system/get-started/dismiss`** (admin): clears the installation step
(`next_step = ''`), for every admin. `GetMe` clears it too, when it finds all
three items `done`.

## 4. The card

On the home page, for an admin whose `GetMe` has `setupChecklist`:

- **Secure the dashboard** - its status; the conditions (the domain points at
  this server's IP, port 80 is open); **Get the certificate**, which calls the
  endpoint and, while `obtaining`, reloads `GetMe` every few seconds. `failed`
  shows the error. `done` shows: "Certificate installed. Quit and reopen your
  browser to see this site as secure."
- **Turn on two-factor authentication** (recommended) - opens the existing
  two-factor setup dialog.
- **Connect a GitHub App** (optional) - what it gives: signing in with GitHub,
  and apps created from repositories; opens the existing `provision-github-app`
  dialog. While the dashboard's certificate is not `done`: GitHub's webhooks
  need a valid certificate, so do that first.
- A close button, which calls `dismiss`.

## 5. Testing

- **Go:** the checklist's state for each item and status; `dismiss`, and the
  clearing when all three are done; the retry past `RetryAfter`, and the
  conflict while obtaining; the first boot asking for the certificate, and
  going on when that fails.
- **Dashboard:** typecheck and lint; a run against the backend: the card, each
  button, closing it.
- **The installer's e2e**, whose domain resolves nowhere: the dashboard comes
  up, and `GetMe` says `dashboardCert: failed`.

## Not in this design

- **Requiring two-factor authentication** of the first admin.
- **Restarting Traefik** when the dashboard's certificate arrives.
- **A wizard** or a page of its own.
- **Checking DNS** before asking for the certificate: the authority's answer
  says what is wrong.
