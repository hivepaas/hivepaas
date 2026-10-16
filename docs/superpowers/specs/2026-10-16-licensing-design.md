# Plans and licensing

Status: draft for review, 2026-10-16. Nothing built.

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

A dash means no limit. Bee is what an installation runs without a valid key.
Apiary is sold by contact. It is for many nodes or users, several
installations, invoices instead of cards, and support terms. The portal issues
its keys by hand, outside Paddle's checkout.

## What counts

Each count is installation-wide and read when something is created.

- **Apps**: apps not deleted, excluding:
  - a preview (`ParentID` set);
  - a database cloned for a preview (`LogicalParentID` set);
  - an app of the hidden `hivepaas` project (system apps);
  - a function.
- **Functions**: apps whose app-kind setting has `category = function`, with
  the same exclusions. An app can change category (`appsettingsuc/function_rules.go`),
  so that change is checked too.
- **Users**: users `active` or `pending`. An invitation holds a seat, or a
  limit of 2 could be invited past. `disabled` users do not count.
- **Nodes**: every node of the swarm, whatever its state. Removing a node from
  the swarm frees its place; draining it does not.
- **Previews**: previews that exist, which are apps with `ParentID` set. The
  databases a preview clones belong to it and are not counted again.
- **Sign-in providers**: on Bee, an OAuth setting of kind `google`,
  `microsoft-online` or `openid-connect` cannot be created or enabled. Ones
  that exist keep signing people in. Turning them off would lock out users who
  have no password, and the rule is never to break what runs. The dashboard
  marks them.

## What a limit does

| Limit      | Checked at                                                                                          |
|------------|-----------------------------------------------------------------------------------------------------|
| Apps       | create app, from template, clone, spec import, project from compose, category change from function |
| Functions  | create function, clone of a function, category change to function                                   |
| Users      | invite, create, re-enable a disabled user                                                           |
| Nodes      | the node join command (`nodeuc/node_join*.go`)                                                      |
| Previews   | creating a preview, by webhook or by hand                                                           |
| Sign-in    | creating or enabling an OAuth setting of a paid kind                                                |

- **The error.** A refusal is `ERR_PLAN_LIMIT_REACHED`, with params `Limit`,
  `Max` and `Plan`, and its translation. The dashboard turns it into an
  upgrade prompt. The CLI and MCP pass it on as they pass any error.
- **No race past a limit.** Two creations at once must not both see 19 and
  make 21. The creating transaction takes
  `pg_advisory_xact_lock(hashtext('plan-limit:<kind>'))` before it counts. One
  lock per kind is enough, because counts are installation-wide and creations
  are rare.
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
signature and an ML-DSA-65 signature, both required (`pkg/releasesig.Open`).
It uses its own key pair, kept apart from the release keys:
- releases are signed offline, by hand (`tools/releasesign`);
- licenses are signed by the portal, unattended;
- a leaked license key must never sign a release.

```json
{
  "version": 1,
  "keyId": "lic-2026-1",
  "licenseId": "lic_01K…",
  "subscriptionId": "sub_01h…",
  "customer": { "id": "acc_…", "name": "Acme", "email": "ops@acme.example" },
  "plan": "hive",
  "limits": { "nodes": 10, "users": 10 },
  "installations": ["<swarm cluster ID>"],
  "issuedAt": "2026-11-01T00:00:00Z",
  "expiresAt": "2026-12-01T00:00:00Z",
  "graceDays": 14
}
```

- **Limits.** A limit left out of `limits` is no limit. Bee's limits are not in
  any key: they are compiled in, as the default.
- **Installation binding.** The key names the swarm cluster ID (`docker info`
  → `Swarm.Cluster.ID`) of the installation it is for. System > License shows
  that ID, and checkout and the portal ask for it. A key for another
  installation is reported as such and not applied. Re-initialising a swarm,
  for example in disaster recovery, makes a new ID. The portal lets the owner
  move their key to the new one, a few times a year by themselves and more
  through support. Apiary keys may name several installations.
- **Validity.** A key holds from `issuedAt` to `expiresAt`, the end of the
  period paid for. It then works for `graceDays` more, with a banner. After
  that the installation is on Bee, which only stops creating.
- **Offline renewal.** At every renewal the portal emails the new key and
  keeps it on the account. A yearly plan needs this once a year, a monthly plan
  once a month.
- **Online renewal.** A daily task posts the current key and the cluster ID to
  the portal (`POST /api/licenses/refresh`) and stores any newer key it gets
  back. It sends nothing else. It is on by default and can be switched off.
  When it fails, the next day tries again.
- **Clock.** An offline installation can set its clock back, and nothing can
  stop that. As a cheap check, the service keeps the latest time it has seen
  and reads a clock that went back more than a day as that time.
- **Key rotation.** The binary carries the license public keys by `keyId`, so
  a new key pair is added in one release and used by the portal in a later
  one.

## Paddle and the portal

- **Catalog.** Three products (Comb, Hive, Colony), each with a monthly and a
  yearly price, in sandbox and live. The checkout passes `custom_data`:
  `{ accountId, installationId }`.
- **Webhooks.** The portal verifies each one's signature, applies it
  idempotently by event ID, and mirrors the subscription in its database:

  | Event                    | Portal                                                                   |
  |--------------------------|--------------------------------------------------------------------------|
  | `subscription.created`   | the subscription and its first key                                       |
  | `transaction.completed`  | a renewal: the key for the new period, emailed                           |
  | `subscription.updated`   | a plan change: a key with the new limits at once; a scheduled change at its date |
  | `subscription.past_due`  | nothing more: the key's grace covers Paddle's retries                    |
  | `subscription.canceled`  | no further keys; the last one runs out with its grace                    |

- **Account.** The portal's account screens list the subscriptions, the keys
  to copy, and the installation each key is for. They open Paddle's customer
  portal for card, invoices and cancellation.
- **Signing.** The license private keys live only in the portal's secret
  store.

## In HivePaaS

- **Storage.** A global setting of a new type `license` holds the key as
  pasted. It is not a secret, but it carries the customer's email, so it is
  never logged.
- **`licenseservice`.** `Current()` returns the plan, the effective limits,
  the status (`none`, `valid`, `grace`, `expired`, `invalid`,
  `other-installation`), the dates and the installation ID. It is cached and
  read again when the setting changes. `Check(ctx, db, kind)` takes the lock,
  counts, and returns `ERR_PLAN_LIMIT_REACHED`. The usecases listed under
  "What a limit does" call it.
- **API.** All of these need the System module; the first only needs read.
  - `GET /system/license` returns the status above, with usage per limit.
  - `PUT /system/license` takes a key.
  - `POST /system/license/refresh` asks the portal now.
- **Development and e2e.** The e2e environment makes far more than 20 apps.
  In development mode the binary also accepts a development public key.
  `env/up.sh` installs a Colony key signed by it. A `tools/licensesign` signs
  such keys, and keys for support, the way `tools/releasesign` signs releases.

## Dashboard

- **System > License**:
  - the plan and the status with its dates;
  - usage against each limit;
  - the installation ID, with a copy button;
  - a box to paste a key, and "Check for a newer key".
  - an Upgrade link to the portal's pricing, carrying the installation ID.
- **Banners**, shown to admins: grace (n days left), expired (now on Bee), and
  over a limit ("7 of 5 nodes: no node can be added").
- **Limit errors.** A refused creation opens a dialog that names the limit and
  links to Upgrade.
- **Sign-in providers.** On Bee, a paid sign-in kind shows as unavailable in
  Settings > OAuth. One that already exists is marked as working but not
  editable.

## Phases

1. **HivePaaS core:**
   - the key format and its verification, and the `license` setting;
   - `licenseservice` and Bee's defaults;
   - each check and its error, and the API;
   - `tools/licensesign`, and a development key for e2e.
2. **Portal:**
   - accounts and the Paddle sandbox catalog;
   - webhooks, key issuance, renewal emails and the refresh endpoint;
   - portal-web's pricing, checkout and account pages.
3. **Dashboard:** System > License, the banners, the limit dialog, and the
   sign-in providers on Bee.
4. **Live:**
   - Paddle live and the production license keys;
   - the website's pricing page and the docs, including which data online
     renewal sends.

## Open questions

- **Trial.** Should a paid plan come with a trial, for example 14 days of
  Hive through Paddle's trial periods?
- **Refunds.** What is the policy? Paddle, as merchant of record, handles tax.
- **portal-web's stack.** The Paddle tooling at hand assumes Next.js.
- **Swarm re-initialisation.** How many self-service moves of a key a year?
