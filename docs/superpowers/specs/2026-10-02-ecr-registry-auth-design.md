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
credential that does not expire (a service account's JSON key; an admin user or
a service principal). ECR does not. This spec does ECR, on a shape the other two
can take later.

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
  saved; `Username` is `AWS`; `Password` stays empty - the token is never
  stored in the setting.
- The secret access key is an `EncryptedField`, masked as a password is.
- Keys on the host (the EC2 instance's role, through the SDK's default chain)
  are a later option: they make the credential depend on which node asks, and
  the app and worker are not always on EC2.
- A setting migration is not needed: `kind` empty is today's behaviour.

### Getting a token: one service

`registryauthservice` (new, or in `registryservice`) answers what Docker takes:

```go
// AuthConfig is the credential as Docker takes it: a username and password
// credential as stored; an ECR one with a token got for it, fresh for at
// least an hour.
AuthConfig(ctx context.Context, setting *entity.Setting) (*registry.AuthConfig, error)
// AuthHeader is AuthConfig, encoded as Docker's X-Registry-Auth.
AuthHeader(ctx context.Context, setting *entity.Setting) (string, error)
```

- For ECR it calls `ecr.GetAuthorizationToken` (aws-sdk-go-v2's `service/ecr`;
  the SDK is in the module already, for S3), decodes the token -
  base64 of `AWS:<password>` - and keeps it **in memory, per process**, keyed by
  the setting's id and `UpdateVer`, until an hour before `expiresAt`. A
  singleflight makes concurrent deploys ask once.
- Not cached in Redis: the token is a credential, a fetch is one AWS call, and
  an app, a worker and an updater each asking once every 11 hours is nothing.
- An AWS error is the deploy's error, worded: "the AWS credentials of
  <name> were refused: <reason>".
- **Every use above goes through it**, and `RegistryAuth.GenerateAuthHeader`
  refuses a credential of kind ECR, so that a path missed fails in tests rather
  than with an expired token in production.

### Renewing what Swarm keeps

A task `registry-auth-renew` runs every 6 hours (a system scheduled job, like
`ssl-renewal`), and when an ECR credential is saved:

1. for each active ECR credential, a fresh token;
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
  phase 1 checks on one of them that a renewal leaves its tasks alone, and the
  renewal compares the spec it sends with the one it read.
- `docker service inspect` does not show `PullOptions`: the stored credential
  cannot be read back, so renewal does not try to compare it - it sends the
  fresh one.

A service whose update fails is retried at the next run, and listed on the
credential's page with the error.

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
- `make gen-swag`; the dashboard's types follow.

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
1. **Backend**: the kind and its fields, the token service with its cache, every
   use moved onto it, `GenerateAuthHeader` refusing ECR; tests with a fake ECR
   client - a token decoded, cached, renewed an hour before it expires, an AWS
   error worded.
2. **Renewal**: the system job and its run on save, the services found and
   updated, failures kept for the page; tests against a fake Docker.
3. **API, dashboard, export, MCP, docs.**
4. Later, the same `Kind` for Google Artifact Registry and Azure ACR when they
   are used with short-lived tokens.

## Risks

- **A restart on renewal**: none when only the auth changes (checked); a spec
  that round-trips differently would restart, which phase 1 guards against.
- **AWS keys in HivePaaS**: a pull-only IAM user is what the docs ask for; the
  key is encrypted at rest like every secret, never answered by the API.
- **Clock skew** on the hosts makes AWS refuse signed calls; the error says so.
- **A node that pulls between expiry and renewal**: renewal at 6 hours leaves
  6 hours of margin on a 12-hour token.
