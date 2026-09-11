# Volume lazy materialization

A volume's configuration lives in one node's docker daemon, so it cannot reach
the node that needs it. This moves the configuration into the volume's setting,
stops creating the docker volume up front, and makes the node pin something
swarm enforces instead of something HivePaaS merely records.

## The problem

`CreateVolume` calls `dockerManager.VolumeCreate`, which is a plain call to the
daemon HivePaaS is connected to - the manager node it runs on. Volumes created
with the `local` driver are node-scoped; there is no cluster-wide registry for
them. So a volume "pinned" to node A is in fact created on whatever node runs
HivePaaS, and `docker volume inspect` on node A does not find it.

What the pin does reach is the bind directory: `createBindDirectoryInNode` runs
`mkdir` on the pinned node through `nodeexecservice`. The directory is in the
right place; the volume describing it is not.

When a task then lands on a node that has neither, it does not fail. It silently
gets an empty place to write, by one of two mechanisms.

**Bind volumes.** `useBindMountIfAppropriate` rewrites a `local` volume whose
options carry `type=none` and a `device` into a plain bind mount at
`device/subpath`, with `BindOptions.CreateMountpoint = true`. The absolute path
travels to any node, but `CreateMountpoint` means a node missing the directory
creates an empty one and the app writes into it.

**Everything else** - nfs, tmpfs, custom drivers - stays a `TypeVolume` mount
naming a volume that does not exist on that node. Docker creates it with default
options and no driver config:

```
Name: hp-voltest-vol   Driver: local   Options: null
Mountpoint: /var/lib/docker/volumes/hp-voltest-vol/_data
```

For an NFS volume this is silent data loss: the app writes to local disk while
appearing healthy.

Both branches share a root cause. The authoritative description of a volume sits
in one node's docker, and nothing carries it to the others.

## What changes

### 1. The setting holds the full specification

`entity.ClusterVolume` gains the fields needed to rebuild a volume anywhere:

```go
type ClusterVolume struct {
	NodeID    string `json:"nodeId,omitempty"`
	NodeLabel string `json:"nodeLabel,omitempty"`

	// Managed says HivePaaS wrote the specification below and is responsible
	// for reproducing the volume on whatever node needs it. A volume that
	// arrived through discovery belongs to somebody else: it is mounted by
	// name and nothing about it is inferred.
	Managed    bool              `json:"managed,omitempty"`
	Driver     string            `json:"driver,omitempty"`
	DriverOpts map[string]string `json:"driverOpts,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}
```

The whole struct is written once, at creation, and never changed.

### 2. Node pinning becomes immutable

Pinning is decided when the volume is created and cannot be updated afterwards.
Update keeps only `inheritable` and `default`, which is what the note already in
`UpdateVolume` describes.

This is what makes everything else simple. An immutable specification cannot go
stale, so a mount spec derived from it is correct for as long as the volume
exists, and a placement constraint derived from it never has to be recomputed
because the volume moved.

It removes work added in `deb5d974`:

- `volumedto.UpdateVolumeReq.NodeID`, `.NodeLabel`, and `Pinning()`
- the pinning branch in `UpdateVolume`'s `PrepareUpdate`
- the update pinning tests in `volumedto/create_test.go`
- dashboard: the pair in `toVolumeUpdatePayload`, `PINNING_CHANGE_WARNING`, and
  the `warnOnPinningChange` prop; the pinning fieldset joins the read-only group
  on the update form

The warning that a pinning change moves no data goes away with it, because that
situation can no longer arise.

### 3. Creation stops writing to docker

`CreateVolume` records the setting and creates the bind directory on the pinned
node, as it does today. It no longer calls `VolumeCreate`.

`CreateProjectDefaultVolume` in `volumeserviceimpl` is the other creation site
and changes the same way: it builds the same driver options it builds today and
stores them in the setting instead of handing them to docker. Its directory is
created through `MakeSubDirInHost` before the volume is recorded, which is
unchanged.

Nothing is lost by deferring. `docker volume create` with the `local` driver
records a specification and touches no filesystem - a volume naming a device
that does not exist is created without complaint:

```
docker volume create --opt type=none --opt device=/does/not/exist --opt o=bind
→ created
```

So the eager call never validated anything. The first honest check of a bind
volume is the mount, and that happens when a task starts either way.

The duplicate-name check moves from `VolumeInspect` to the settings table.

### 4. New volumes are named by their setting ID

A new volume's docker name is its `Setting.ID` (a ULID), recorded in
`Setting.RefID` the way a docker-assigned name is today. Because `RefID` already
carries the docker-side identity, existing volumes keep the names they have and
nothing needs renaming or migrating.

Human-readable identity moves into volume labels, which survive the trip to
another node - a volume materialized from a mount spec carries the labels the
spec declared:

```
Options: {'device': '/var/lib/hp-t-bind', 'o': 'bind', 'type': 'none'}
Labels:  {'hivepaas.volume.name': 'pgdata'}
```

so `docker volume ls --filter label=hivepaas.project=<key>` stays usable.

### 5. Mount specs are built from the setting

`useBindMountIfAppropriate` reads `DriverOpts` from the setting instead of
`dockerVol.Options`. The logic is unchanged; only the source moves, and that is
what frees it from the manager node's docker.

`buildDockerMount` fills `VolumeOptions.DriverConfig` from the setting for
`TypeVolume` mounts when `Managed` is set. The plumbing already exists - it is
populated from the client request today and simply never fed from the volume's
own configuration.

With the driver config in the mount spec, docker materializes the volume
correctly on whichever node the task lands, with no HivePaaS involvement:

```
task reads: du-lieu-that-trong-vm     ← the data that was already there
host sees:  /var/lib/hp-t-bind/from-task   ← the task's write, in the real directory
```

and when the device is genuinely missing the task fails loudly instead of
inventing a place to write:

```
starting container failed: failed to mount local volume:
mount /var/lib/hp-t-bind:... no such file or directory
```

`prepareUpdatingAppStorageSettings` no longer calls `VolumeListByIDs`, so there
is no `dockerVol` to be nil.

`CreateMountpoint: true` stays. For a volume with no pin - the claim that every
node reaches the same data - creating the directory is the correct behaviour.
Pinned volumes are covered by the constraint below.

### 6. The pin becomes a placement constraint

A service mounting a pinned volume gets `node.id==<id>` or
`node.labels.<k>==<v>` added to its placement constraints, so a task can only be
scheduled where its data is.

`placementservice` already has the mechanism: constraints HivePaaS sets are
recorded in the `hivepaas.app.placementConstraints` label and recomputed on each
apply, which preserves constraints the operator set by hand. The volume-derived
constraint joins `newHivepaasConstraints` and inherits that behaviour.

It is the first *required* constraint HivePaaS emits; every existing one is an
exclusion. Two volumes pinned to different nodes therefore produce a set no node
satisfies. That is a real contradiction rather than a subtle misconfiguration,
so saving the storage settings is refused, naming both volumes, instead of
leaving swarm with a task that can never be scheduled.

Constraints are applied wherever HivePaaS writes an app's service spec through
`placementservice`: creating the app, deploying it (image or repo build), and
saving its storage settings. The storage case is folded into the same
`ServiceUpdate` that writes the mounts - that update reschedules the tasks, so a
constraint arriving one update later would let a task land on a node with none
of the data in the meantime.

Saving an app's *placement* settings is not one of those moments, despite the
name: `appplacementsettingsuc.UpdateAppPlacementSettings` stores the setting and
writes no service spec, so that change reaches swarm at the app's next deploy.

There is no sweep across existing apps.
Sweeping would move a task that is running on the wrong node today, and the node
it is running on may be where its data actually accumulated; correcting that
belongs to a moment the operator is already looking at that app. A diagnostic
reports the apps whose running node differs from the pin of a volume they mount,
so the mismatch is visible without being acted on automatically.

### 7. Sync inverts

Today docker is where volumes are discovered and the setting is the copy. After
this change the setting is authoritative, so **absence from docker no longer
means the volume was deleted** - "never materialized" is the normal state of a
freshly created volume. `TestSyncVolumesMarksVanishedVolumesDeleted` changes
accordingly.

Sync keeps two jobs:

- **Backfill.** A setting with no recorded `Driver` gets one read from docker on
  the manager node, with `Managed` set, unless the docker volume is a cluster
  volume. A node-scoped volume that HivePaaS already holds a setting for is one
  it is already responsible for; recording the specification only makes that
  responsibility executable on more than one node. The specification is copied
  from the volume as it exists, so stamping it back is a no-op where it came
  from and faithful everywhere else.
- **Discovery.** Volumes created outside HivePaaS still get settings, and the
  setting starts with `Managed` false and no `Driver` recorded. That lasts one
  sync pass. Backfill, above, cannot tell a setting discovery just wrote from
  one that predates this change - both have an empty `Driver` - so the next
  sync backfills the one discovery just created the same as any other,
  stamping on the specification and setting `Managed: true`, unless the volume
  is a swarm cluster volume. That is intended, not a gap: the specification
  backfill records is copied from the volume as it exists, so stamping it back
  is faithful to what is actually there, and it is what lets the volume be
  rebuilt correctly on another node instead of docker silently creating an
  empty default one in its place.

The rule the sync already follows is unchanged and now covers one more field:
nothing HivePaaS did not author is overwritten.

### 8. Deletion

A volume can now exist on several nodes. `DeleteVolume` keeps removing the
manager node's copy and already ignores a not-found, which is the common case
now that volumes materialize lazily.

The existing cluster cleanup does not do the rest. `clustercleanupservice`
calls `dockerManager.VolumePrune(ctx, true)`
(`clustercleanupserviceimpl/cleanup_cluster.go:100`), and that `true` is
`anonymousOnly`: the wrapper in `services/docker/volume.go` adds the `all=true`
filter only when `anonymousOnly` is false, and docker's own client documents
that filter as controlling whether named volumes are pruned at all - by
default, only anonymous ones are. Every HivePaaS volume is named: `RefID` is a
ULID for one created on this branch, or the pre-existing docker name for an
older one. So the prune that already runs on every node reclaims none of them,
deleted or not.

That means a volume deleted after materializing on more than one node leaves
copies behind on every node but the manager - a gap this change introduces,
since before lazy materialization a volume only ever existed on the manager in
the first place, so there was nothing else to orphan. The cost differs by
kind: for a bind volume it is only a stale pointer, since the data lives in
the bind target directory, not in the volume, so nothing is lost but an unused
docker object; for a plain local volume with no bind, nfs or tmpfs options,
the data is in the volume itself, under `/var/lib/docker/volumes`, and is
wasted disk left behind on that node.

This is recorded as a known gap, not closed here. Closing it needs removal
addressed to a specific node, which is a capability that already exists in a
different form: the same `nodeexecservice` path `createBindDirectoryInNode`
uses today to create a bind directory on the pinned node can reach a node to
remove a volume from it too.

It must not be closed by flipping `VolumePrune`'s `anonymousOnly` argument to
false instead. That would prune every unused named volume on every node, not
just the copies a deleted setting left behind - including a volume whose
setting still exists but whose service happens to be scaled to zero right now.
For a bind volume that is harmless, the same as above. For a plain local
volume it is not: an operator's data would be destroyed for the crime of being
briefly unused, turning today's disk-space leak into permanent data loss.

## What is deliberately not changing

**HivePaaS still cannot create CSI cluster volumes, and this does not add that.**
A volume is a cluster volume only if the create request carries
`ClusterVolumeSpec`, and nothing in the codebase sets it - the field appears only
in a response DTO, in `VolumeUpdate`'s signature, and in vendored code. Typing a
CSI driver name into the driver field produces a node-scoped volume through that
plugin, not a cluster volume. Cluster volumes enter HivePaaS only through
discovery, and `Managed: false` keeps them mounted by name exactly as they are
today.

**Volumes created outside HivePaaS keep working as they do now.** They are
mounted by name, with no driver config inferred on their behalf.

## Compatibility

Existing volumes keep their docker names through `RefID`; only new ones are named
by ULID. Existing settings gain their specification through backfill on the next
sync, after which they behave like new ones. Until then they mount by name, which
is current behaviour.

The one behavioural change operators will notice is that an app mounting a pinned
volume becomes constrained to that node the next time its service is updated. An
app that was running elsewhere stops moving freely - which is the point, but it
is a change to scheduling that a running cluster will feel.

## Testing

- mount specs built from a setting, for bind, nfs, and tmpfs
- `DriverConfig` present for `Managed` volumes and absent for discovered ones
- placement constraints generated from a pinned volume; contradictory pins refused
  when storage settings are saved
- sync does not mark an unmaterialized setting deleted
- backfill records a specification once and never overwrites one already there
- creation writes no docker volume, and a duplicate name is refused from the
  settings table
