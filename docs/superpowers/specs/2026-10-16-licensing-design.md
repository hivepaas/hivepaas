# Plans and licensing

Status: settled 2026-10-16. Phases 1 (HivePaaS core) and 2 (portal) built;
phase 3 (dashboard) being built. The limits are enforced from the first release that carries them,
so that release also carries the production license public keys.

HivePaaS stays Apache 2.0 and self-hosted. What is sold is a license key that
lifts the limits of the free plan. Paddle takes the payments, as merchant of
record. The portal (`hivepaas-portal`, `hivepaas-portal-web`) turns
subscriptions into keys. An installation checks its key offline.

## What this is, and is not

- **It is a fair-use gate, not DRM.** The limits live in Apache 2.0 code, so a
  fork can remove them, and that is allowed. The design keeps an honest
  customer's path short and leaves the rest alone. It does not obfuscate,
  phone home to permit, or check integrity at run time.
- **A limit only ever stops creating.** Nothing that runs is stopped, scaled
  down or deleted: not over a limit, not when a license expires, not on a
  downgrade. An installation over its limits, such as one of today's beta
  installs on the free plan, keeps everything and can add nothing more of that
  kind until it is under the limit or licensed.
- **Offline on every plan.** A key is verified against public keys compiled
  into the binary. Fetching renewals is a convenience, never a condition.

## Plans

|                        | **Bee**                          | **Comb**  | **Hive**   | **Colony** | **Apiary**  |
|------------------------|----------------------------------|-----------|------------|------------|-------------|
| Monthly                | free                             | $4.99     | $12.99     | $24.99     | quoted      |
| Yearly (20% off)       | -                                | $47.99    | $124.99    | $239.99    | quoted      |
| Nodes                  | 2                                | 5         | 10         | 25         | set per key |
| Users                  | 2                                | 5         | 10         | 30         | set per key |
| Apps                   | 20                               | -         | -          | -          | -           |
| Functions              | 20                               | -         | -          | -          | -           |
| Previews at once       | 2                                | -         | -          | -          | -           |
| Sign-in providers      | GitHub, GitHub App, GitLab, Gitea | all      | all        | all        | all         |
| Installations per key  | 1                                | 1         | 1          | 1          | set per key |
| Trial                  | -                                | -         | 14 days, no card | -    | -           |

A dash means no limit. Bee is what an installation runs without a valid key.
Apiary is sold by contact. It is for many nodes or users, several
installations, invoices instead of cards, and support terms. The portal issues
its keys by hand, outside Paddle's checkout.

## What counts

Each count is installation-wide and read when something is created.

- **Apps**: apps not deleted, excluding:
  - a preview (`ParentID` set);
  - an app that belongs to another (`LogicalParentID` set): a database cloned
    for a preview, or one a template makes beside its app. App lists leave
    them out too;
  - an app of the hidden `hivepaas` project (system apps);
  - a function.
- **Functions**: apps whose app-kind setting has `category = function`, with
  the same exclusions. An app cannot change to or from a function
  (`ERR_APP_KIND_FUNCTION_UNCHANGEABLE`), and templates cannot make one, so
  creating, importing and cloning a function are what is checked.
- **Users**: users `active` or `pending`. An invitation holds a seat, or a
  limit of 2 could be invited past. `disabled` users do not count.
- **Nodes**: every node of the swarm, whatever its state. Removing a node from
  the swarm frees its place; draining it does not.
- **Previews**: previews that exist: apps with `ParentID` set and no
  `LogicalParentID`. The databases a preview clones have `ParentID` too, but
  `LogicalParentID` names the preview: they belong to it and are not counted.
- **Sign-in providers**: on Bee, an OAuth setting of kind `google`,
  `microsoft-online` or `openid-connect` cannot be created, enabled, or made
  of such a kind. Ones
  that exist keep signing people in. Turning them off would lock out users who
  have no password, and the rule is never to break what runs. The dashboard
  marks them.

## What a limit does

| Limit      | Checked at                                                                                          |
|------------|-----------------------------------------------------------------------------------------------------|
| Limit      | Checked at                                                                     |
|------------|--------------------------------------------------------------------------------|
| Apps       | create app, from template, clone, spec import, project from compose            |
| Functions  | create function, spec import, clone of a function that keeps its deployment settings |
| Users      | invite (a new person), re-enable a disabled user                               |
| Nodes      | the node join command (`nodeuc/node_join*.go`)                                 |
| Previews   | creating a preview, by webhook or by hand                                      |
| Sign-in    | creating, enabling or changing the kind of an OAuth setting, spec import too   |

- **The error.** A refusal is `ERR_PLAN_LIMIT_REACHED` (412), and its
  translation. Its params `Limit`, `Max` and `Plan` are on the wire, in the
  error's `params`, the one error that exposes params. The dashboard turns it
  into an upgrade prompt. The CLI and MCP pass it on as they pass any error.
- **Clones and previews are tasks.** The request is checked, for an answer on
  screen, and the task that makes the app checks again under the lock: queued
  tasks are not counted.
- **No race past a limit.** Two creations at once must not both see 19 and
  make 21. The creating transaction takes
  `pg_advisory_xact_lock(hashtext('plan-limit:<kind>'))` before it counts. One
  lock per kind is enough, because counts are installation-wide and creations
  are rare. A kind with no limit takes no lock, so paid plans never wait on
  it. Nodes and sign-in providers are not counted in the database and take
  none.
- **Previews are made by webhooks.** Nobody is looking at a screen when a pull
  request opens a preview. A refused preview comments on the pull request, as
  a preview's deployment does, and tells the app's notification target. For
  example: "Bee runs 2 previews at a time: close one, or upgrade."
- **Nodes can join without HivePaaS.** Anyone with the swarm's join token can
  run `docker swarm join`, and HivePaaS cannot prevent it. HivePaaS refuses
  its own join command over the limit. It shows a node count over the limit
  as a banner and does nothing else.

## The license key

A key is the `releasesig` envelope: the payload as base64, with an ed25519
signature and an ML-DSA-65 signature, both required (`pkg/releasesig.Open`),
signed under the context `hivepaas-license-v1`. As text, a key is that
envelope's JSON as base64 on one line, which survives email and copying;
whitespace in a pasted key is ignored, and the envelope's JSON is accepted as
it is. It uses its own key pair, kept apart from the release keys:
- releases are signed offline, by hand (`tools/releasesign`);
- licenses are signed by the portal, unattended;
- a leaked license key must never sign a release.

```json
{
  "version": 1,
  "licenseId": "lic_01K…",
  "subscriptionId": "sub_01h…",
  "customer": { "id": "acc_…", "name": "Acme", "email": "ops@acme.example" },
  "plan": "hive",
  "limits": { "nodes": 10, "users": 10 },
  "installations": ["<swarm cluster ID>"],
  "issuedAt": "2026-11-01T00:00:00Z",
  "expiresAt": "2026-12-01T00:00:00Z",
  "graceDays": 14,
  "trial": false
}
```

- **Limits.** A limit left out of `limits` is no limit. Bee's limits are not in
  any key: they are compiled in, as the default.
- **Installation binding.** The key names the swarm cluster ID (`docker info`
  → `Swarm.Cluster.ID`) of the installation it is for. System > License shows
  that ID, and checkout and the portal ask for it. A key for another
  installation is reported as such and not applied. Apiary keys may name
  several installations.
- **Moving a key.** Re-initialising a swarm, for example in disaster recovery,
  makes a new ID. The owner moves their license to the new ID in the portal,
  up to **10 times in any 365 days**, and beyond that through support.
  - A move issues a key for the new ID and marks the old one superseded.
  - An online installation still on the old ID learns of the move at its next
    refresh, and drops to Bee once the old key's grace runs out.
  - An offline one keeps the old key until it expires. That is accepted: the
    limit on moves bounds it, and the gate is fair-use, not DRM.
- **Validity.** A key holds from `issuedAt` to `expiresAt`, the end of the
  period paid for. It then works for `graceDays` more, with a banner. After
  that the installation is on Bee, which only stops creating.
- **Offline renewal.** At every renewal the portal emails the new key and
  keeps it on the account. A yearly plan needs this once a year, a monthly plan
  once a month.
- **Online renewal.** A daily task posts the current key and the cluster ID to
  the portal (`POST /api/licenses/refresh`, at `https://portal.hivepaas.com` by
  default). It stores any newer key it gets back, or the news that the key was
  moved to another installation. It sends nothing else. It is on by default
  and can be switched off. When it fails, the next day tries again.
  - The key is what authenticates the request: the portal answers only for a
    key it issued, and only for the installation that key names. So an
    installation with no key cannot refresh, and its first key, trial or paid,
    is pasted once; renewals then arrive by themselves. Answering by
    installation ID alone would hand the customer's email to anyone who knows
    the ID.
  - The answer is `newer` (with the key to install: a renewal, a plan change,
    or a paid key replacing a trial), `moved` (with when), or `current`.
- **Clock.** An offline installation can set its clock back, and nothing can
  stop that. As a cheap check, the service keeps the latest time it has seen
  and reads a clock that went back more than a day as that time.
- **Key rotation.** The binary carries the license public keys
  (`pkg/license/licensekeys`), each named by its key id, the id the envelope's
  signatures give. A new key pair is added in one release and used by the
  portal in a later one.

## Trial

A 14-day trial of Hive, with no card. The portal issues it, not Paddle: it is a
key like any other, so it works offline and needs no new path in HivePaaS.

- **Starting a trial.** System > License has "Start a 14-day trial", which
  opens the portal with the installation ID. The person signs in or signs up
  (with an email address, verified; GitHub or Google sign-up may come later)
  and starts the trial. The key is shown to copy and is emailed.
- **What the key holds.** Hive's limits, `"trial": true`, `expiresAt` 14 days
  out, and `graceDays: 0`. When it runs out, the installation is on Bee, which
  only stops creating.
- **Once each.** One trial per installation ID and one per account. The portal
  keeps both. A trial is not moved to another installation.
- **Buying.** A trial can be turned into a paid plan at any time through the
  normal checkout. The paid key replaces it.
- **Paddle's own trials** need a card and convert to paid on their own. They
  may come later, as a price with a trial period, if conversion calls for
  them.

## Paddle and the portal

- **Catalog.** Three products (Comb, Hive, Colony), each with a monthly and a
  yearly price, in sandbox and live. Each price carries `custom_data`
  `{ plan, period }`, which is how the portal knows what it sells: no price ID
  is written into code.
- **Checkout.** The portal creates the checkout's transaction itself, with
  `custom_data` `{ accountId, installationId, checkoutId }`, and the web opens
  Paddle's checkout on that transaction. What the browser sends is never
  trusted for whom a key is for. The portal refuses a purchase for an
  installation that already has a license in force: its plan is changed
  instead.
- **Webhooks.** The portal verifies each one's signature, applies it
  idempotently by event ID, and mirrors the subscription in its database:

  | Event                    | Portal                                                                   |
  |--------------------------|--------------------------------------------------------------------------|
  | `subscription.created`   | the subscription and its first key                                       |
  | `transaction.completed`  | a renewal: the key for the new period, emailed                           |
  | `subscription.updated`   | a plan change: a key with the new limits at once. A scheduled cancel or pause changes nothing until it takes effect |
  | `subscription.past_due`  | nothing more: the key's grace covers Paddle's retries                    |
  | `subscription.paused`    | no further keys until it resumes                                         |
  | `subscription.canceled`  | no further keys; the last one runs out with its grace                    |

  A key runs to the end of the latest period paid for, not of the period
  Paddle is in: that one moves on even while a renewal is failing. Paddle has
  no scheduled plan change: a plan change takes effect at once, and only its
  billing can wait.

- **Account.** One account per person for now; the data allows team members
  later. The portal's account screens list the subscriptions, the keys to
  copy, and the installation each key is for. They open Paddle's customer
  portal for card, invoices and cancellation.
- **Plan changes** are made in the portal, as Paddle's customer portal cannot
  make them: an upgrade (a higher plan, or monthly to yearly) is charged now,
  prorated; a downgrade from the next period. The key with the new limits is
  issued at once.
- **Moves** are made in the portal, for a subscription's license; a manual
  (Apiary) license is moved by staff.
- **Staff.** Apiary and other manual keys, support moves, revoking a license
  and events Paddle sent that could not be applied go through a staff API,
  documented in the portal's `docs/licensing.md`.
- **Signing.** The license private keys live only in the portal's secret
  store. The portal signs with its own small signer, held to `tools/licensesign`
  byte for byte by a test. A sandbox portal signs with the development key, so
  only a HivePaaS in development trusts what it issues; a live one with the
  production keys, and refuses to start with the development key.
- **Grace and Paddle's retries.** The 14 days of grace cover Paddle's retries
  of a failed renewal only while its payment recovery runs 14 days or less:
  set it so in live, or raise the grace.

## In HivePaaS

- **Storage.** A global setting of a new type `license` holds the key as
  pasted. It is not a secret, but it carries the customer's email, so it is
  never logged.
- **`licenseservice`.** `Current()` returns the plan, the effective limits,
  the status (`none`, `valid`, `trial`, `grace`, `expired`, `invalid`,
  `other-installation`, `moved`), the dates and the installation ID. It is cached and
  read again when the setting changes. `Check(ctx, db, kind)` takes the lock,
  counts, and returns `ERR_PLAN_LIMIT_REACHED`. The usecases listed under
  "What a limit does" call it.
- **API.** All of these need the System module; the first only needs read.
  - `GET /system/license` returns the status above, with usage per limit, and
    the customer's name, never the email.
  - `PUT /system/license` takes a key. It refuses one that is invalid, for
    another installation, or past its grace, and keeps the key in force. Any
    other key replaces the one stored, a lower plan included. There is no
    DELETE.
  - `POST /system/license/refresh` asks the portal now.
- **Development and e2e.** The e2e environment makes far more than 20 apps.
  In development (`HP_ENV=development`, which also opens the dev-helper
  endpoints, so no production runs it) the binary also accepts a development
  public key. Its private half is committed with `tools/licensesign`, which
  signs such keys, and keys for support, the way `tools/releasesign` signs
  releases. `env/up.sh` installs a Colony key signed by it.

## Dashboard

- **System > License** (`/system/license/overview/`):
  - the plan and the status with its dates, and whether the key is a
    development one;
  - usage against each limit;
  - the installation ID, with a copy button;
  - a box to paste a key, which asks before replacing a key in force, and
    "Check for a newer key";
  - an Upgrade link to the portal's pricing, carrying the installation ID, and
    "Start a 14-day trial" while there is no key.
- **Banners**, shown to whoever may read the System module, as the license
  itself is; each can be put off for a day, or a week once on Bee:
  - trial (n days left);
  - grace (n days left);
  - expired (now on Bee);
  - moved (the key now belongs to another installation);
  - a key for another installation, or one that cannot be verified (on Bee);
  - over a limit ("7 of 5 nodes: no node can be added"). Sign-in providers
    are left out: Settings > OAuth marks them.
- **Limit errors.** A refused creation opens a dialog that names the limit and
  links to Upgrade. A clone or a preview refused in its task is told where
  that task's failure is: the task, and for a preview its pull request.
- **Sign-in providers.** On Bee, a paid sign-in kind shows as unavailable in
  Settings > OAuth. One that already exists is marked as kept: it can still
  be edited, so an expiring client secret can be replaced, but once disabled
  it cannot be enabled again on Bee.

## Phases

1. **HivePaaS core:**
   - the key format and its verification, and the `license` setting;
   - `licenseservice` and Bee's defaults;
   - each check and its error, and the API;
   - `tools/licensesign`, and a development key for e2e;
   - the production license public keys, since the limits apply from this
     release.
2. **Portal:**
   - accounts and the Paddle sandbox catalog;
   - webhooks, key issuance, renewal emails and the refresh endpoint;
   - trials, and moving a key to a new swarm (10 a year);
   - portal-web's pricing, checkout and account pages.
3. **Dashboard:** System > License, the banners, the limit dialog, and the
   sign-in providers on Bee.
4. **Live:**
   - Paddle live, and the production license private keys in the portal;
   - the website's pricing page and the docs, including which data online
     renewal sends.

## Open questions

- **Refunds.** What is the policy? Paddle, as merchant of record, handles tax.
  For now nothing is automatic: staff revoke a license through the staff API,
  and a key already issued runs out on its own.
