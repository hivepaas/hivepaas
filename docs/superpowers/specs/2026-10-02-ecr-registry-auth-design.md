# Amazon ECR registry credentials that renew themselves

## Why

A registry credential is a username and a password, kept encrypted and handed to
Docker as they are. Amazon ECR has no such password: `docker login` to ECR takes
an authorization token (`GetAuthorizationToken`, `aws ecr get-login-password`),
user `AWS`, that **expires after 12 hours**, and only AWS credentials can get a
new one. A token pasted into HivePaaS works for 12 hours, then every pull and
push fails - including those nobody started: Swarm pulling the image again to
start a task on another node, or after one died.

Google Artifact Registry (an access token, 1 hour) and Azure ACR (an AAD token,
3 hours) have the same shape when used with short tokens, but both also take a
credential that does not expire: a service account's JSON key, user
`_json_key_base64`; an admin user or a service principal. ECR does not. This
spec does ECR, on a shape the other two can take later - and lets Google's JSON
key in, which today's 100-character password limit keeps out.

**In scope:** a registry credential of kind Amazon ECR - its AWS credentials,
the token got from them when needed, the token renewed where Swarm keeps it; the
API, the dashboard, the spec export, MCP, the docs. **Not in scope:** Google and
Azure; ECR Public; pull-through cache rules; creating repositories.

## Where a credential is used today

Every use reads `RegistryAuth.Username` and `Password` as stored:

| Use | Where |
|---|---|
| Pulling an app's image at deploy, and resolving its digest | `appdeploymentserviceimpl/image_deploy_pull_img.go` (`GenerateAuthHeader`) |
| Creating or updating the app's Swarm service, which **stores** the credential in the service's spec (`EncodedRegistryAuth` → `ContainerSpec.PullOptions`) for every later pull | `image_deploy_apply_svc.go`, `repo_deploy_apply_svc.go`, `function_deploy_apply_svc.go` |
| A build's registries, for its `FROM` images: every active credential of the project | `imagebuildserviceimpl/build_registry.go` |
| Pushing a built image (`pushToRegistry`) | `imagebuildserviceimpl/inputs.go` |
| Testing a credential | `registryauthuc/test_conn.go` |
| The system registry's own credential | `registryserviceimpl` - never ECR, unchanged |

The second row is the one a fixed token cannot survive: what Swarm stored at
deploy is what it uses at 3 a.m. the next day.

## Design

### The credential

`entity.RegistryAuth` gains a kind and, for ECR, its AWS side:

```go
type RegistryAuth struct {
	// Kind is how the credential signs in: "" (a username and a password, as
	// today) or "aws-ecr".
	Kind     base.RegistryAuthKind `json:"kind,omitempty"`
	Username string                `json:"username"`
	Password EncryptedField        `json:"password"`
	Address  string                `json:"address"`
	// ...
	ECR *RegistryAuthECR `json:"ecr,omitempty"`
	// Token is the last token got for the credential, and when it expires:
	// derived from the keys, never answered by the API or exported.
	Token          EncryptedField `json:"token,omitzero"`
	TokenExpiresAt time.Time      `json:"tokenExpiresAt,omitzero"`
}

type RegistryAuthECR struct {
	Region          string         `json:"region"`
	AccessKeyID     string         `json:"accessKeyId"`
	SecretAccessKey EncryptedField `json:"secretAccessKey"`
	// RoleARN, when set, is assumed with the keys above first.
	RoleARN string `json:"roleArn,omitempty"`
	// RegistryID is the account whose registry it signs in to; the keys'
	// own when empty.
	RegistryID string `json:"registryId,omitempty"`
}
```

- `Address` is the registry's, `<account>.dkr.ecr.<region>.amazonaws.com`,
  filled in from the token's `proxyEndpoint` when the credential is tested or
  saved; `Username` is `AWS`; `Password` stays empty.
- The secret access key is an `EncryptedField`, masked as a password is.
- Keys on the host (the EC2 instance's role, through the SDK's default chain)
  are a later option: they make the credential depend on which node asks, and
  the app and worker are not always on EC2.
- A setting migration is not needed: `kind` empty is today's behaviour.

### Getting a token: when it is needed, kept in the database

`registryauthservice` (new, or in `registryservice`) answers what Docker takes:

```go
// AuthConfig is the credential as Docker takes it: a username and password
// credential as stored; an ECR one with a token that lives at least as long
// as the renewal's interval and an hour more.
AuthConfig(ctx context.Context, db database.IDB, setting *entity.Setting) (*registry.AuthConfig, error)
// AuthHeader is AuthConfig, encoded as Docker's X-Registry-Auth.
AuthHeader(ctx context.Context, db database.IDB, setting *entity.Setting) (string, error)
```

- **The token is got when it is needed**, not on a timer: a deploy, a build, a
  push, a test of the credential, a renewal run. The stored token is used while
  it has **at least the renewal's interval plus an hour** left - 7 hours with
  the default 6-hour interval; otherwise a new one is got with
  `ecr.GetAuthorizationToken` (aws-sdk-go-v2's `service/ecr`; the SDK is in the
  module already, for S3), its base64 `AWS:<password>` decoded.
- The rule is the renewal's interval plus an hour because **every token handed
  to Swarm must live until the next renewal**: a deploy hands one over as the
  renewal does. With a 12-hour token and a 6-hour interval, the stored one is
  under 7 hours at every run, so each run gets a new one: about two AWS calls
  per credential every 12 hours.
- **Kept in the setting**, `token` encrypted as a password is, with
  `tokenExpiresAt`: an app, a worker and an updater share it, and it survives a
  restart. It is written by its own `UPDATE` of those two fields, **without
  bumping `updateVer`**: a renewal must not make the person editing the
  credential lose their edit to a version conflict. The row is locked
  (`SELECT ... FOR UPDATE`) while a token is got, so two deploys at once ask AWS
  once.
- **Saving new keys clears the token.** It is never in an API answer, the
  audit log, an export or MCP: `registryauthdto` and the spec export leave it
  out, and a test fails if one does not.
- An AWS error is the caller's error, worded: "the AWS credentials of <name>
  were refused: <reason>".
- **Every use goes through it**, and `RegistryAuth.GenerateAuthHeader` refuses a
  credential of kind ECR, so that a path missed fails in tests rather than with
  an expired token in production.

### Renewing what Swarm keeps

A system scheduled job, built as `ssl-renewal` is:

- a global setting `registry-auth-renewal` - its schedule and its notification -
  and a scheduled job of type `registry-auth-renewal` targeting it, made by
  `InitDefaults` with the other system settings;
- **an interval only, from 1 to 10 hours, 6 by default**: no cron, whose gaps
  vary - the token rule needs the longest gap - and nothing past 10 hours, for
  which no 12-hour token is fresh enough;
- retries, a run's history and Run Now as every scheduled job has them, and a
  **notification when a run fails**: a renewal that keeps failing is apps that
  cannot be started on another node within 12 hours;
- in the spec export, as a singleton, as `ssl-renewal` is.

Besides its schedule it runs for one credential when its keys are saved, and at
start when its last run is older than its interval - HivePaaS down for a while
may have let Swarm's copies age.

A run:

1. for each active ECR credential, of every project, a token by the rule above;
2. the Swarm services that pull with it: apps whose `imageSource.registryAuth`
   is it, and repository apps and functions whose `pushToRegistry` is it - their
   image is in that registry;
3. each service updated with the new `EncodedRegistryAuth` and **its spec
   otherwise unchanged**.

**Checked on a local swarm (Docker 29.8):** a service updated three times with
its spec as it was and a different `X-Registry-Auth` each time kept its task -
no restart. The registry auth is not part of what makes a task dirty. Two
things the same check showed:

- the **first** update of a service created by the Docker CLI, its spec sent
  back as inspected and with no registry auth at all, did restart its task:
  the round trip itself changed something in a spec the CLI had written. The
  services renewal updates are HivePaaS's, created and updated through the API;
  phase 3 checks on one of them that a renewal leaves its tasks alone, and the
  renewal compares the spec it sends with the one it read.
- `docker service inspect` does not show `PullOptions`: the stored credential
  cannot be read back, so renewal does not try to compare it - it sends the
  fresh one.

A service whose update fails is retried at the next run, and listed on the
credential's page with the error. With no ECR credential, a run does nothing:
no AWS call, no update.

**Installations that exist already** get the setting and its job too:
`InitDefaults` runs at the first installation and when a system setting is
first read, which an upgrade does not guarantee; it is run at start as well, so
that a missing default is made whatever is opened.

### API and dashboard

- `registryauthdto` create/update/get carry `kind` and `ecr` (the secret key
  write-only, masked on read). Validation: region in AWS's form, a key id, a
  secret; `username`, `password` refused for ECR.
- **Test connection** gets a token and pings the registry's `/v2/` with it;
  its answer fills `Address`.
- The credential's form: **Kind** - *Username and password* / *Amazon ECR*;
  for ECR: **Region**, **Access Key ID**, **Secret Access Key**, **Role ARN**
  (optional), **Registry ID** (optional). The page shows when the token was
  last got and the renewal's last result.
- **System → Registry Auth Renewal**, beside SSL Renewal: the interval, the
  notification, Run Now, the last runs; turning it off with an ECR credential
  in use warns that their apps cannot be pulled on another node after 12
  hours.
- `make gen-swag`; the dashboard's types follow.

### A longer password, for Google's JSON key

A Google Artifact Registry credential that does not expire is a service
account's JSON key: user `_json_key_base64`, password the key file in base64,
about 3 KB. The password limit goes from 100 characters to 8 KB, in the API
and the dashboard's field; the docs say how. Nothing else is needed: it is a
username and a password as today's are.

### Elsewhere

- **Spec export/import** (`specservice`): the kind and region exported, the
  keys treated as the password is today.
- **MCP**: `list_registry_auths` shows the kind; nothing else changes.
- **Docs**: the registry credentials page - ECR, the IAM policy it needs
  (`ecr:GetAuthorizationToken`, and `ecr:BatchGetImage`,
  `ecr:GetDownloadUrlForLayer` to pull; the push actions to push to it).

## Phases

0. **Spike**: done for the restart question (above). Left: one real
   `GetAuthorizationToken` answer, to pin the decoding - it needs an AWS
   account.
1. **The longer password**, for Google's JSON key: on its own, small, now.
2. **Backend**: the kind and its fields, the token service and its rule, the
   token's own update and lock, every use moved onto it, `GenerateAuthHeader`
   refusing ECR; tests with a fake ECR client - a token decoded, stored, used
   while it has the interval and an hour left, got again after, an AWS error
   worded, never answered.
3. **Renewal**: the setting and its job, `InitDefaults` at start, the run on save
   and at start, the services found and updated with their tasks left alone,
   failures kept for the page; tests against a fake Docker, and one on a real
   service of HivePaaS's that its task keeps its id.
4. **API, dashboard, export, MCP, docs.**
5. Later, the same `Kind` for Google Artifact Registry and Azure ACR when they
   are used with short-lived tokens.

## Risks

- **A restart on renewal**: none when only the auth changes (checked); a spec
  that round-trips differently would restart, which phase 3 guards against.
- **AWS keys in HivePaaS**: a pull-only IAM user is what the docs ask for; the
  key is encrypted at rest like every secret, never answered by the API.
- **Clock skew** on the hosts makes AWS refuse signed calls; the error says so.
- **A node that pulls between expiry and renewal**: none while the rule holds -
  every token handed over lives past the next run. A run that fails is
  notified, and retried; two missed runs in a row are an expired copy.
- **Writing a token into a setting** that a person is editing: its own update,
  without the version, keeps their edit.
