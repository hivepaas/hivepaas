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
3. **One card for every admin, and only the certificate is required.** The card
   is shown while the installation step is `hivepaas/get-started`; closing it, or
   the dashboard's certificate being done, clears the step for everyone. Two-factor
   authentication belongs to each admin and has its own warning, so the card does
   not list it; a GitHub App is a suggestion, shown with no state of its own.
4. **The certificate's state is worked out, not ticked,** from what exists, and
   only when the card asks: `GetMe` does no more work than before.
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
  the tasks it returns scheduled on the task queue once the queue has started
  (`DashboardCertOnFirstBoot`, after `InitTaskQueue`);
- the installer starts the app only after every service is tuned and the
  database migrated, so nothing restarts it while the certificate is obtained,
  and waits a little for the certificate before saying done (the installer's
  spec, §2);
- a failure here is logged and does not stop the boot: the card offers the
  attempt again.

The task, when it succeeds, attaches the certificate to the domain and applies
the routing again, as it does for any app.

## 3. The API

**`GetMe`** is unchanged: it gives an admin `nextStep`, which says whether the card
is shown.

**`GET /system/get-started/dashboard-cert`** (admin), asked by the card on the home
page only, and every few seconds only while `obtaining`:

```json
{"status": "failed", "domain": "hivepaas.example.com", "error": "acme: ... NXDOMAIN ..."}
```

For the HivePaaS app's first enabled HTTP domain:

- `done` - the domain names a certificate that has content, has not expired, and
  is not self-signed (nothing else is looked up then);
- `obtaining` - an obtain task on the certificate setting named for the domain
  is waiting, running, or failed with a retry to come;
- `failed` - that setting has a `lastError`, which is returned;
- `todo` - none of these.

The first answer that finds it `done` also clears the step, and still says
`done`, so the card can say so until the page is loaded again.

**`POST /system/get-started/dashboard-cert`** (admin): asks for the dashboard's
certificate now, as the first boot does, but past the wait a failure leaves
(`RetryAfter`, six hours): the admin asking is the reason to try. Refused with
`ERR_CONFLICT` while one is being obtained. Answers with where the certificate stands, as the `GET` does.
`EnsureCertsReq` gains `IgnoreRetryAfter`.

**`POST /system/get-started/dismiss`** (admin): clears the installation step
(`next_step = ''`), for every admin. Clearing touches only `hivepaas/get-started`.

## 4. The card

On the home page, for an admin whose `GetMe` says `nextStep: hivepaas/get-started`:

- **Secure the dashboard** - its status; the conditions (the domain points at
  this server's IP, port 80 is open); **Get the certificate**, which calls the
  endpoint and, while `obtaining`, reloads `GetMe` every few seconds. `failed`
  shows the error. `done` shows: "Certificate installed. Quit and reopen your
  browser to see this site as secure."
- **Connect a GitHub App** (optional, a suggestion with no state) - what it
  gives: signing in with GitHub, and apps created from repositories; opens the
  existing `provision-github-app` dialog. While the dashboard's certificate is not
  `done`: GitHub's webhooks need a valid certificate, so do that first.
- A close button, which calls `dismiss`.

## 5. Testing

- **Go:** the certificate's state for each status; the step cleared by the
  answer that finds it done, and only that step; the retry past `RetryAfter`, and the
  conflict while obtaining; the first boot asking for the certificate, and
  going on when that fails.
- **Dashboard:** typecheck and lint; a run against the backend: the card, each
  button, closing it.
- **The installer's e2e**, whose domain resolves nowhere: the dashboard comes
  up, and `GET /system/get-started/dashboard-cert` says `failed`.

## Not in this design

- **Requiring two-factor authentication** of the first admin, or listing it on
  the card: each admin's own warning does that.
- **Tracking a GitHub App** on the card.
- **Restarting Traefik** when the dashboard's certificate arrives.
- **A wizard** or a page of its own.
- **Checking DNS** before asking for the certificate: the authority's answer
  says what is wrong.
