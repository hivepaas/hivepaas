# What an app may do through the Docker API

An app given Docker API access in HivePaaS talks to a proxy, not to Docker. The proxy refuses
everything unless a rule below lets it through, and everything the app reaches is what it created
itself. The dashboard shows the same guide on an app's Docker API screen; the two are kept together
with `pkg/dockerproxy`, and change when it does.

## Always allowed - the core

Every app with Docker API access has these, whatever permissions are ticked.

**System**
- Checking the daemon is there, and which version it runs: `GET|HEAD /_ping`, `GET /version`.
- The node's info, without the cluster's: `GET /info`.

**Images**
- Pulling an image, and creating a container of it, only when it matches one of **Images**:
  `POST /images/create`.
- Listing and inspecting the images of the node: `GET /images/json`, `GET /images/{name}/json`.
- Importing an image from a tarball is refused.

**Its own containers**
- Creating a container, checked field by field: what reaches past the container is refused, the
  limits are applied, and it is labeled as the app's: `POST /containers/create`.
- Starting, watching, stopping and removing the containers it created, and seeing only those:
  `GET /containers/json` (only the app's own, and the app itself),
  `GET /containers/{id}/json|logs|stats|top`,
  `POST /containers/{id}/start|stop|kill|wait|restart|resize|attach`, `DELETE /containers/{id}`.
- `GET /events`, filtered to its own containers.

**Network**
- Every container joins the app's own network, `hp-dapi-<app>`. Asking for bridge or host lands
  there too.
- With **Env network** ticked, a container may also join the env's network, to reach the env's
  other apps.
- No network at all (`none`) is fine; sharing another container's network is refused.

**Storage**
- tmpfs mounts.
- A bind of a **Shared directory**, which becomes a mount of the app's own directory on its volume.
- A mount of a **Shared volume**'s name, which becomes the same mount as a bind of the directory it
  stands for. No volume of that name is created or used.
- Any other path of the node is refused.

**Limits**
- Containers: how many the app may have at once, running or not. One more is refused.
- Memory and CPUs: the most one container may ask for, and what it gets when it asks for none.
  Asking for more is refused.
- Processes: 1024 for a container that asks for no limit, and at most 4096.
- Unlimited swap is refused.

## Permissions

### exec - run commands inside the app's own containers, as `docker exec` does

- Allows starting a command in a running container the app created, attaching to its input and
  output, resizing its terminal and reading how it ended: `POST /containers/{id}/exec`,
  `POST /exec/{id}/start`, `POST /exec/{id}/resize`, `GET /exec/{id}/json`.
- Only in containers the app created. Fields allowed: User, Cmd, Env, WorkingDir, Tty,
  AttachStdin, AttachStdout, AttachStderr, DetachKeys, ConsoleSize. Privileged is refused.
- Without it: the app can start containers and read their logs, but every exec is refused with 403.
- Used by: the Gitea / Forgejo runner (every step of a job), Jenkins Docker agents, Coder.
- Risk: low. The command runs inside a container the app already controls.

### files - copy files into and out of the app's own containers, as `docker cp` does

- Allows `PUT|GET|HEAD /containers/{id}/archive`: upload a tar archive, download a path, stat it.
- Only containers the app created; what is written lands in the container's own filesystem or what
  it mounts, never a path of the node.
- Without it: files reach a container only through its image, a shared directory or a volume.
- Used by: the Gitea / Forgejo runner, to copy the workspace in and results out.
- Risk: low.

### volumes - keep volumes of the app's own, and mount them into its containers by name

- Allows `GET /volumes`, `POST /volumes/create`, `GET /volumes/{name}`, `DELETE /volumes/{name}`,
  and naming a volume in a container's mounts - one that does not exist yet is created for the app.
- Only the local driver, with no driver options. Every volume is labeled as the app's; any other
  volume is refused. A subpath must stay inside its volume.
- Volumes are not counted by the limits. They stay while the app has access, and are removed when
  access is turned off or the app is deleted.
- Without it: containers use tmpfs, the shared directories and the shared volumes only.
- Risk: medium - disk space; the volumes are caches that grow, and no limit bounds them.

### networks - keep networks of the app's own, and connect its containers to them

- Allows `GET /networks`, `POST /networks/create`, `GET|DELETE /networks/{id}`,
  `POST /networks/{id}/connect|disconnect`.
- Only a bridge on this node, with the default IPAM: no overlay, no driver options. An endpoint
  takes Aliases and DNSNames only, no static address. Connecting: only the app's own containers,
  and only to the app's network, the env network when allowed, or a network the app created.
- A network no container uses, older than an hour, is removed by the node's agent.
- Without it: every container joins the app's own network, and the env network when allowed.
- Risk: low.

### nestedSocket - give the app's containers the same Docker API

- Allows mounting the app's socket into a container: a bind of `/var/run/hivepaas/docker.sock`, or
  the volume `hp-dapi-sock-<app>`. What that container starts goes through the same proxy, under
  the same policy, and counts toward the same container limit.
- It is not Docker-in-Docker: no second daemon, nothing privileged.
- Without it: mounting the socket is refused, and a job that runs docker fails.
- Risk: medium - the code a job runs can start containers of any image the app may run.

## Never allowed, whatever is ticked

--privileged; added capabilities; devices and GPUs; the node's network, PID or IPC namespace;
security options and runtimes; sysctls; publishing ports; any path of the node beyond the shared
directories; volumes-from and links; docker build; swarm services and stacks; events of anything
but its own containers, prune and other calls across the whole daemon; another app's containers,
volumes or networks.
