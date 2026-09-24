# A registry HivePaaS runs for itself

**Status:** draft, awaiting review

## 1. What this is for

An image built on one node has to be pullable on every other node. Today that means an
external registry - Docker Hub, GHCR, a private one - with an account the operator creates
by hand, credentials they paste into a registry-auth setting, and a storage bill or a quota
that is somebody else's rule. An operator who does not want any of that has no answer.

The second half of the problem is the one that makes people avoid self-hosting a registry at
all: it fills up. HivePaaS pushes one tag per build
([calcBuildImageTags](../../../hivepaas_app/service/imagebuildservice/imagebuildserviceimpl/helper.go)
names them `<address>/<username>/<app key>:<commit sha7>`), so a repository gains a tag every
deploy and loses none. A registry nobody prunes grows until the disk does not.

This spec covers **one** registry: a system registry, global, provisioned and owned by
HivePaaS, with cleanup that runs by itself. A project that wants its own registry - a
different account model, a different retention, a registry for something other than HivePaaS
builds - provisions one from the catalogue (`registry.yaml`, `zot.yaml`) like any other app.
Nothing here changes that path, and nothing here is in its way.

## 2. Decisions

| Question | Decision | Why |
|---|---|---|
| Which registry? | zot | Its garbage collection runs online. distribution reclaims bytes only through `registry garbage-collect`, which needs the registry read-only - a maintenance window inside a PaaS that deploys on every push |
| Where does it run? | An app in the hidden `hivepaas` project | Routing, SSL and its renewal, deploy, logs, resources and the app screen all work because it is an ordinary app. Every app in that project is a system app already, and the project is filtered out of the dashboard's project lists |
| How is that app created? | `appprovisionservice.ProvisionApp` with a `Configure` written in Go | The same call the template usecase makes, without the catalogue. The configuration has to be exact and has to move with the backend that depends on it; a catalogue template is pinned separately and may be edited |
| Storage | The operator picks at provision: a local volume (default) or S3 | A single-node install wants neither an S3 account nor its latency. A cluster that reschedules wants no node pinning |
| Node pinning | Falls out of the volume | `ClusterVolume` already carries `NodeID`/`NodeLabel` and the placement constraint is derived from it. S3 means no volume and no constraint |
| Deduplication | On with a volume, off with S3 | zot refuses to start with `dedupe: true` and S3 unless a remote cache (DynamoDB) is configured. With S3 the same layer in two repositories is stored twice |
| Cleanup | zot's own retention plus its online GC, driven by two numbers | Measured to reclaim real bytes with no read-only window. No scheduled job, no API client, no lock, nothing to go wrong at 3am. §7 keeps the door open for a keep-set job later |
| Docker media types | `compat: ["docker2s2"]`, always | Without it a daemon on the classic image store gets 415 on the manifest after uploading every blob. Measured, §3 |
| The credential | One managed `registry-auth` setting in global scope, created with the app | It is a registry auth; making it one means the build path, the deploy path and the app settings screen already know what to do with it |
| Behind Cloudflare | Warn, detect what can be detected, and offer a real test | The proxy is not something HivePaaS can turn off. What it can do is tell the operator before they find out from a failed 500 MB push |
| Disabling it | Removes nothing | The images are on a volume or in a bucket the operator paid for. Deleting an app is done in the app screen, deliberately, not as a side effect of a checkbox |

## 3. What was measured

A throwaway spike on `ghcr.io/project-zot/zot:v2.1.21`, Docker Engine 29.7.2, filesystem and
S3 (MinIO) backends. The design rests on these numbers, so they are recorded here.

| Question | Result |
|---|---|
| Push from the docker daemon | containerd image store pushes an OCI index (with an attestation manifest): accepted, `docker pull` round-trips. Classic store's `application/vnd.docker.distribution.manifest.v2+json`: **415** unless `compat: ["docker2s2"]` is set, and then v2s2 manifests and v2s2 manifest lists both work |
| Local state with S3 | `meta.db` and `cache.db` in `rootDirectory` are a derived cache. Deleting both and restarting, and starting a second container with an empty directory against the same bucket, both served catalog, tags, manifests and `docker pull` at once. No volume, no pinning |
| `dedupe: true` + S3 | zot refuses to start: "dedupe set to true with remote storage and database, but no remote database configured" |
| `DELETE /v2/<repo>/manifests/<digest>` | 202, tag gone at once. Bytes reclaimed at the next GC pass: 47240 KB -> 23848 KB, 106 s later on a volume; 23656 KB -> 4096 KB, 22 s later on S3 |
| Overwriting a tag, with no API call | The displaced manifest became untagged and was collected: 43424 KB -> 23864 KB, ~74 s later |
| Age retention | `pushedWithin: "2m"` removed the tag once it aged out and kept a tag matched by a second rule. Six 8 MB images aged out together: 71028 KB -> 23988 KB, and the empty repository disappeared from `/v2/_catalog` |
| Count retention | `mostRecentlyPushedCount: 2` kept the two newest of four tags: 16072 KB -> 10168 KB |
| `pulledWithin` | Protects only images that were actually pulled. An image nobody pulled is outside the window, not inside it |
| Pushes during GC | Six builds pushed while passes were running: six succeeded, none failed. There is no read-only window |
| Cadence | One generator pass visits each repository once, a few seconds apart, and the next pass starts about `gcInterval` after the previous finished. A repository therefore waits up to ~2x `gcInterval`. `gcDelay` shields blobs younger than itself, so it must exceed the slowest push |

## 4. The setting

A new `base.SettingTypeRegistry = "registry"`, global scope, one per system, stored the way
every other setting is: `entity/setting_registry.go` with a parser registered through
`registerSettingParser`, plus `entity/setting_registry_migration.go` for the version bump
convention.

```go
type RegistrySettings struct {
    Enabled bool                  `json:"enabled,omitempty"`
    Type    base.RegistryType     `json:"type,omitempty"`    // "zot" today
    Managed bool                  `json:"managed"`           // defaults true in New

    // Domain is the address images are named with, and it is required to enable the
    // registry: a docker daemon speaks to a registry over HTTPS at a name, so a registry
    // without one is a registry nothing can push to.
    Domain string `json:"domain,omitempty"`

    Storage RegistryStorage `json:"storage"`
    Cleanup RegistryCleanup `json:"cleanup"`

    // MemoryLimit is the app's limit, default 512MB, minimum 256MB.
    MemoryLimit unit.DataSize `json:"memoryLimit,omitempty"`

    // App and RegistryAuth are what provisioning created, written back after it ran.
    // They are how Apply finds its own work again, and how the dashboard links to it.
    AppID          string `json:"appId,omitempty"`
    RegistryAuthID string `json:"registryAuthId,omitempty"`
}

type RegistryStorage struct {
    // Type is decided at provision and refused afterwards: see §10.
    Type base.RegistryStorageType `json:"type,omitempty"` // "volume" (default) | "s3"

    // VolumeID is a cluster-volume setting, for Type "volume".
    VolumeID string `json:"volumeId,omitempty"`

    // CloudStorageID is a cloud-storage setting, for Type "s3". Bucket and prefix come
    // from it; nothing about S3 is duplicated here.
    CloudStorageID string `json:"cloudStorageId,omitempty"`
}

type RegistryCleanup struct {
    // Enabled defaults to true in New: a registry that never prunes is the problem this
    // feature exists to solve.
    Enabled bool `json:"enabled"`

    // Mode names how the keep set is decided. "policy" is the only value today; §7
    // explains what a second value would be for.
    Mode base.RegistryCleanupMode `json:"mode,omitempty"`

    // KeepLast keeps this many of the newest tags of every repository, default 10.
    KeepLast int `json:"keepLast,omitempty"`
    // KeepDays keeps every tag pushed within this many days, default 30.
    KeepDays int `json:"keepDays,omitempty"`
}
```

No password lives here. The account's password is in the managed `registry-auth` setting, in
its `EncryptedField`, and its bcrypt hash is in the app's htpasswd secret. `New()` writes
every defaulted field unconditionally, for the reason
[setting_logging.go](../../../hivepaas_app/entity/setting_logging.go) states: with
`omitempty`, a stored `false` is dropped from the JSON and the default survives the
unmarshal as a silent `true`.

`base.SettingTypeRegistry` also has to be named in
[specmodel.singletonBlockNames](../../../hivepaas_app/service/specservice/specmodel/singleton.go)
as `"registry"`, or `TestEverySettingTypeIsClassifiedExactlyOnce` fails. It is a singleton:
one per installation, read through `GetSingle`. A global-scope spec bundle therefore carries
the configuration, while the app itself is not in any bundle, because `selectProjects` skips
the `hivepaas` project. That is the right split, and it puts one requirement on spec import
when it is built: it may write the setting, but an app is provisioned only by `Apply`, which
runs when somebody saves.

The endpoints follow logging exactly - `GET /system/settings/registry` and
`PUT /system/settings/registry` in
[router_system.go](../../../hivepaas_app/interface/api/server/router_system.go), a
`systemsettings/registryuc` built on `UpdateUniqueSetting`, and a `registryservice` whose
`Apply` makes the cluster match what was saved.

## 5. Provisioning

> Since 2026-09-24 the parts of this that any system app needs - where it lives, how it is
> created, how it is removed - are `systemappservice`, which the logging stack uses too.

`Apply` is called after every save and does nothing that is already done. It is the only
thing that writes the app, so a failed save leaves work the next save retries.

1. **Validate.** A domain, which is required; a volume or cloud-storage setting that exists
   and is readable in global scope; `KeepLast >= 1`, `KeepDays >= 1`; memory at or above
   256MB.
   With `Enabled: false` nothing below runs, and nothing is torn down (§10).

   The volume also has to be one the registry's app can see. A volume reaches an app only
   when it is marked inheritable - that is the rule `applyAppFilter` enforces - and the
   registry's app lives in a project of its own, so a volume that is not shared with apps is
   refused here, by name, rather than failing later inside the mount build as "Volume not
   found". The dashboard offers only the volumes that qualify (§13).
2. **The account.** Username is fixed: `hivepaas`. The password is generated once, 32 bytes
   of base62, and stored in a managed `registry-auth` setting in global scope whose
   `Address` is the domain. Its id goes into `RegistryAuthID`. On later saves the password
   is left alone; §11 is how it changes.
3. **The htpasswd.** bcrypt of that password, written as the app's `ZOT_HTPASSWD` secret,
   mounted at `/etc/zot/htpasswd`. HivePaaS hashes it itself (`golang.org/x/crypto/bcrypt`),
   so nothing asks the operator to run `htpasswd` by hand the way the catalogue template
   does.
4. **The configuration.** §6, rendered from the settings, written as the app's config file
   and mounted at `/etc/zot/config.json`.
5. **The app.** If `AppID` is empty, `ProvisionApp` into the `hivepaas` project, environment
   `default` - the key [project_sync.go](../../../hivepaas_app/service/projectservice/projectserviceimpl/project_sync.go)
   gives a system app that names no environment - with app key `registry`, a new
   `base.HivepaasRegistryKey` beside `HivepaasAppKey` and the rest, and the image
   `ghcr.io/project-zot/zot:v2.1.21` pinned in the backend. It carries the secret and the
   config file above, `/var/lib/registry` mounted from the chosen volume when storage is a
   volume, routing on port 5000 with the domain and `forceHttps`, the memory limit, and the
   healthcheck disabled - the image is one static binary with no shell to run a check with.
   The app id is written back into the setting.
6. **The first deployment.** `ProvisionApp` creates the service with a placeholder image and
   returns the deployment and certificate tasks **unscheduled**: a task row can be picked up
   only once the transaction it was written in has committed, and `Apply` runs inside the
   caller's. They travel back in the response, and the usecase schedules them after the
   commit, the way the template usecase does. Without that the registry sits on the
   placeholder image for ever.
7. **Later saves.** The config file and the settings are rewritten and the app redeployed.
   A swarm config object is immutable, so a changed configuration is a new object and a
   service update: the registry restarts, which §10 says what to do about.

The S3 credentials are read from the cloud-storage setting and passed to zot as
`storageDriver.accesskey`/`secretkey` inside the config file, which is a swarm config -
readable by anyone who can read the service spec, which is already true of every secret
HivePaaS mounts this way. They are not written into the registry setting.

## 6. The configuration zot gets

```jsonc
{
  "distSpecVersion": "1.1.1",
  "storage": {
    "rootDirectory": "/var/lib/registry",     // with S3: a scratch path, see §3
    "dedupe": true,                            // false with S3: zot refuses otherwise
    "gc": true,
    "gcDelay": "2h",                           // must exceed the slowest push
    "gcInterval": "1h",                        // a repository waits up to ~2h, §3
    "retention": {                             // only when cleanup.enabled
      "dryRun": false,
      "policies": [{
        "repositories": ["**"],
        "deleteUntagged": true,
        "deleteReferrers": true,
        "keepTags": [
          {"patterns": [".*"], "mostRecentlyPushedCount": 10},   // cleanup.keepLast
          {"patterns": [".*"], "pushedWithin": "720h"},          // cleanup.keepDays
          {"patterns": [".*"], "pulledWithin": "720h"}           // cleanup.keepDays
        ]
      }]
    },
    "storageDriver": {                         // only with S3
      "name": "s3", "rootdirectory": "/zot", "bucket": "...", "region": "...",
      "regionendpoint": "...", "secure": true, "forcepathstyle": true,
      "accesskey": "...", "secretkey": "..."
    }
  },
  "http": {
    "address": "0.0.0.0",
    "port": "5000",
    "compat": ["docker2s2"],
    "auth": {"htpasswd": {"path": "/etc/zot/htpasswd"}}
  },
  "log": {"level": "info"},
  "extensions": {"search": {"enable": true}, "ui": {"enable": true}}
}
```

`search` is not decoration: `/v2/_zot/ext/search` answers, authenticated, with each
repository's name, size, last update and newest tag, which is what a storage page in the
dashboard would read and what a keep-set job would read later. `ui` is the browsable
interface on the same domain, behind the same account.

One thing to say out loud in the dashboard: `/v2/_zot/ext/mgmt` answers **without
authentication** with the version, the build flags and which authentication methods are on.
Nothing private, but it is public on a public domain.

## 7. Cleanup

The two numbers are a union, not a sequence: a tag survives if any rule keeps it. That is
zot's own behaviour, measured in §3.

- `KeepLast` -> `mostRecentlyPushedCount`. The newest N tags of every repository stay, however
  old they are. Since a repository is one app, this is "the last N builds of every app".
- `KeepDays` -> `pushedWithin` **and** `pulledWithin`. Everything pushed in the window stays,
  and so does anything a node actually pulled in the window - which is exactly the image a
  long-running service fetched when it was last rescheduled.
- `deleteUntagged: true` always. A tag that a new build displaces takes its old manifest out
  of use, and the bytes come back without anybody deleting anything.
- Everything else is collected, and the bytes are reclaimed within about two `gcInterval`s.

Worked through: an app deployed twice a day with `KeepLast: 10, KeepDays: 30` keeps its last
ten builds and every build of the last month - about 60 tags - and drops the rest. An app
nobody has deployed for a year keeps its ten newest builds, one of which is the image it is
running.

**The hole**, stated plainly: an app that has not been deployed for longer than `KeepDays`
**and** has had more than `KeepLast` builds since the one it is running - which requires
building without deploying - can have its running image collected. The service keeps running,
because the image is already on its node; a reschedule onto a node without it fails to pull.
`pulledWithin` covers the common version of this, since a reschedule pulls. The rest is what
`Cleanup.Mode` exists for: a later `"keep-set"` mode where HivePaaS reads the image of every
live service spec and writes those exact tags into `keepTags` as patterns, or deletes
everything outside the set itself. It changes one field in a struct that already has a place
for it, and nothing else in this design.

## 8. The domain, and the proxy

Traffic to a registry is layer upon layer of blobs. Through Cloudflare's proxy it leaves the
cluster, crosses to Cloudflare, comes back, and meets an upload limit on the way - 100 MB per
request on the free plan, which a single layer passes without trying. The registry's DNS
record has to be **DNS-only**.

HivePaaS cannot see somebody's DNS settings, so it does two things:

- **Detect what is visible.** After the domain is set, request `https://<domain>/v2/` and
  read the response headers: `cf-ray`, `server: cloudflare`, and the generic `via` are proxy
  evidence. Resolving the domain and comparing the addresses with the cluster's node
  addresses is a second, weaker signal - a load balancer is not a proxy - so it is worth a
  sentence in the warning, never a refusal. Detection is advisory: it explains, it does not
  block.
- **Offer a real test.** A "Test a large push" action uploads ~150 MB of nothing in particular
  to `/v2/hivepaas-selftest/blobs/uploads/` through the public domain, then cancels the upload
  with a `DELETE` on the upload URL; anything it leaves behind is an unfinished upload, which
  zot's GC removes on its next pass - measured, an abandoned push left no lasting bytes. The
  backend does this over plain HTTP, with no docker and no build, and reports what came back:
  202 all the way through means the path is clear, 413 means something in front of the
  registry has a body limit, and a timeout means what it says. This is the only answer that is
  not a guess.

## 9. Using it

Because the credential is an ordinary global `registry-auth` setting, everything downstream
already works:

- An app's deployment settings name it in `pushToRegistry`, the same field any other registry
  auth goes in. When the system registry exists, the dashboard offers it as the default choice
  for a new app; apps that already have a choice keep it.
- Builds push `<domain>/hivepaas/<app key>:<sha7>`.
- Nodes pull with the credential swarm carries in the service spec
  (`options.EncodedRegistryAuth`, set in
  [image_deploy_apply_svc.go](../../../hivepaas_app/service/appdeploymentservice/appdeploymentserviceimpl/image_deploy_apply_svc.go)),
  which is how every private registry already works here.

## 10. Changing it afterwards

| Change | What happens |
|---|---|
| Cleanup numbers, memory | Config file rewritten, service updated, a restart of a few seconds. A push landing in that window fails and the build reports it; the operator retries |
| Domain | Routing updated, the registry auth's `Address` rewritten. Images already pushed keep their old names: they resolve while the old DNS record does, and after that the apps have to be rebuilt. The dashboard says so before saving |
| Storage type, volume, or bucket | **Refused.** Nothing copies the images from one to the other, and a registry that silently forgot everything it held is worse than an error message. Changing it means provisioning a new registry and rebuilding |
| `Enabled: false` | The app keeps running and nothing is deleted. HivePaaS stops offering the registry for new apps and stops managing the app's configuration. Deleting the app, and with it the volume, is done in the app screen |

## 11. Rotating the password

An explicit action, not a schedule. It generates a new password, writes it into the registry
auth, and appends a second line to the htpasswd secret - zot accepts both, so the old
credential keeps working. The old line is removed on the next rotation, or after a grace
period the dashboard names.

This grace period is not cosmetic. A swarm service carries the credential it was deployed
with; a service that is not redeployed still presents the old password when a node reschedules
it. The grace period is how long the operator has to redeploy their apps. The dashboard says
which apps still hold the old credential, by listing the apps whose last deployment predates
the rotation.

## 12. When it goes wrong

| Situation | What the operator sees |
|---|---|
| The registry app is down | Builds fail at the push step with the registry's error. Deploys of images a node already has keep working; a reschedule that needs a pull does not |
| The volume's node is gone | The app cannot start, because the volume pins it there. This is the cost of the volume option, and the provisioning screen says it in one line |
| Disk full | Pushes fail with 500 from zot. The storage page shows what each repository holds, from the search extension |
| S3 credentials wrong or bucket missing | zot starts and fails on the first request. Provisioning validates the cloud-storage setting first, so the common case is caught before the app exists |
| A push during a GC pass | Nothing. Measured: six of six succeeded, and `gcDelay` at 2h shields any blob a push is still working on |

## 13. The dashboard

A page under system settings, beside Logging.

Before it exists, one card: what it gives (images pullable on every node, cleanup that runs by
itself), what it asks for (a domain, and either disk on one node or an S3 bucket), and a
button.

The form is short. The domain, with the DNS-only sentence under it - not after a failure, but
while the operator is typing. Storage as two cards through the shared `OptionCardGroup`, a
volume picker or a cloud-storage picker underneath, and the sentence that says what each one
costs. The volume picker lists only volumes that are shared with apps, because only those
reach the registry's app; when none qualifies it says where to make one instead of offering an
empty list: a volume pins the registry to its node, S3 does not and stores shared layers twice.
Memory. Then cleanup: a switch and two numbers, with a line underneath that reads back what
they mean - "Keeps the last 10 builds of every app, and everything from the past 30 days" -
because two numbers in a form are not a policy anybody can picture.

Once it is running the page also carries the registry's state: a link to the app, what it
holds (from the search extension), the credential as a link to its registry-auth setting, the
"Test a large push" button of §8, and "Rotate password", which opens a warning modal in the
house style - the title, a separator, then what will happen to apps that are not redeployed.
The storage cards are disabled with the reason from §10, rather than hidden.

## 14. Testing

- **Entity.** Defaults from `New()` survive a round trip, including the ones that must not
  carry `omitempty`; the migration bumps versions the way the other settings' do.
- **Configuration rendering.** Golden files: volume and S3, cleanup on and off, so that
  `dedupe` is never true with S3, `compat` is never absent, and the retention rules carry the
  numbers from the settings.
- **Validation.** A missing domain, a volume in another scope, `KeepLast: 0`, a storage type
  changed after provisioning - each is refused with the message the dashboard shows.
- **Apply.** With a stubbed provisioner: first save creates the app, the auth and the secret;
  second save with the same settings creates nothing; a changed cleanup number rewrites the
  config file and redeploys; `Enabled: false` leaves everything alone.
- **End to end**, on the development cluster: enable with a volume, build an app with
  `pushToRegistry` pointing at it, check the pushed tag over the public domain, redeploy the
  app from the pushed image, rotate the password and confirm the old credential still pulls.
- Not in CI: zot itself. The spike is the evidence that the configuration this code writes is
  a configuration zot accepts, and the golden files are what keep it that way.

## 15. Later

- `Cleanup.Mode: "keep-set"`, §7.
- A storage page: what each app's repository holds, and the oldest tag, from the search
  extension.
- `distribution` as a second `Type`, for an operator who wants it and accepts its maintenance
  window. The catalogue template already provisions one at project level.
- Per-project credentials with zot's `accessControl`, once there is a reason for a project to
  hold a credential of its own.
- More than one replica, which needs the remote cache zot asks for.
