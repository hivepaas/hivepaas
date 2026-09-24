# Global-mode swarm services: what breaks

The logging collector runs one task per node, which is `swarm.ServiceMode.Global`.
Every service HivePaaS creates today is `Replicated`, so this asks whether any
code that reads a replica count will meet a service that has none.

**Answer: no.** Every path that reads a replica count reaches its service
through an app record or a label filter, and the logging services are neither
apps nor labelled as temporary. No guards are needed.

## Safe, and why

| Path | Why it never sees the collector |
|---|---|
| `clustercleanupserviceimpl/cleanup_temp_containers.go:68` | The only `ServiceList` caller in the repo. It filters on label `hivepaas.temp.resource`, which the logging services do not carry. |
| `appserviceimpl/update_status.go` | `SetAppRunning`, `stopApp` and `startApp` all take `app.ServiceID` from an app record and inspect that one service. Nothing enumerates. |
| `apppreviewserviceimpl/create_preview_settings.go`, `clone_db_apps_settings.go` | Operate on the swarm service of an app being previewed, reached from the app. |
| `appcloneserviceimpl/clone_3_swarm_service.go`, `clone_4_volumes.go` | Same: source and destination are app services. |
| `cmd/internal/docker.go:91` | The docker event listener subscribes with `type=node` and `event=create`. Service events are not delivered at all. |

## Needs a guard

None.

## What this does constrain

The spec HivePaaS sends must set exactly one mode. The daemon rejects a spec
carrying both `Global` and `Replicated`, and a `Replicated` left filled in
alongside `Global` "just in case" is how that happens. `toSwarmServiceSpec` sets
one or the other and a test asserts the unused one stays nil.

## If this changes

The reasoning rests on nothing enumerating swarm services without a filter. A
future `ServiceList` call with no label filter, followed by a read of
`Spec.Mode.Replicated.Replicas`, would panic on the collector. Searching for
`ServiceList(` is enough to check.

## Revised 2026-09-24: the collector is an app

The collector is now an app in the hidden `hivepaas` project, so the premise above - that the
logging services are neither apps nor labelled - no longer holds, and the paths that reach a
service through an app record do meet it:

| Path | What it does with a global service |
|---|---|
| `appserviceimpl/update_status.go` | Disabling an app, or stopping it, scaled it to zero by switching it to `Replicated{0}` - which swarm refuses for a global service (`service mode change is not allowed`, measured). A global service is stopped instead by the constraint `node.id==hivepaas-app-stopped`, which no node satisfies - a node id never contains a hyphen - and started again by removing it (`appservice.StopGlobalService`/`StartGlobalService`). Placement keeps it across a recompute, because it keeps every constraint it does not manage. |
| `appsettingsuc/service_settings_update.go` | Changing an app's mode recreates its service; the stopped state of the one being replaced is dropped with it (`appservice.ForgetStoppedState`). |
| `containerexecserviceimpl/exec.go` | A terminal opens in any running task, the same answer a replicated app with several replicas gets. |

The `ServiceList(` search above still holds for services reached without an app.

