# Kopia Repository Server: A Repository on a Volume, Reached From Anywhere

A backup repository on a volume lives on one node. kopia reaches it through
the agent of that node, which runs its commands. Two things it cannot do yet:
take a stream - the agent's command protocol carries no stdin - and take a
directory from another node. The data backup job refuses neither when it is
saved; their runs fail, saying why. This design makes them run: for the time
of a backup, the agent of the repository's node runs a kopia repository
server, and kopia connects to it from where the data is.

It follows the data backup job (`2026-09-27-data-backup-job-design.md`), and
is the ground restore will stand on: a restore into a volume on another node
than its repository is the same session, read the other way.

---

## Decisions

1. **The server runs in the agent of the repository's node**, a process of its
   own, for one session. It lives as long as the stream that asked for it: a
   backend that closes the stream, is canceled or dies, stops it. Nothing
   outlives the backup.
2. **Only where it is needed.** A repository on cloud storage, and a volume on
   the node of its volume repository, are reached as today. The server serves
   the two cases that cannot be served otherwise.
3. **kopia's own access control.** Clients log in as `hivepaas@data-backup`,
   the identity every data backup snapshot is recorded under; the server
   refuses them any other.
4. **Resources.** Measured with kopia 0.23.1, 1 GiB of incompressible data: the
   server peaked near 1 GB of memory without a limit and near 400 MB with
   `GOMEMLIMIT=256MiB`, one core busy; the stream took about twice as long as
   one written directly. Nothing is used between sessions.

## 1. The agent's server

A new agent RPC, `RepoServer(stream RepoServerReq) returns (stream
RepoServerResp)`.

1. The backend opens the stream and sends the start request:
   - the repository's directory on the host, read under `/host`;
   - the repository's password;
   - the user and its password;
   - the address to listen on: the agent's own, as the backend reached it.
2. The agent makes a session directory, `/tmp/hivepaas/repo-servers/<session>/`,
   and does the following in it:
   1. connects to the repository on the filesystem, with the session's own
      config file;
   2. adds the user, or sets its password when the user exists
      (`kopia server user add` / `set`);
   3. starts `kopia server start` with:
      - a TLS certificate it generates;
      - that address and port 0, so the system picks a free port;
      - a control user, and a control password random for the session, in the
        environment;
      - `GOMEMLIMIT`, and its logs and cache in the session directory;
      - its own process group, so that stopping it stops everything it
        started.
3. Once the server listens, the agent answers with its port and the
   certificate's SHA-256 fingerprint, read from what kopia prints. The backend
   makes the URL, `https://<agent address>:<port>`. kopia's certificate names
   127.0.0.1 only; a client pinning the fingerprint connects by any address
   (checked with kopia 0.23.1).
4. The session ends in one of two ways:
   - **The stream ends** - the backend closes it, is canceled, or is gone. The
     agent stops the server, then removes the session directory.
   - **The server exits by itself.** The agent sends the error, with the tail
     of kopia's stderr, and ends the stream.
5. A server not listening within the start timeout is stopped. The stream ends
   with that error and the tail of stderr.

**Configuration** (`config/agent.go`, read by the agent):

```go
type AgentRepoServer struct {
    MemLimit     string        `toml:"mem_limit" env:"HP_AGENT_REPO_SERVER_MEM_LIMIT" default:"256MiB"`
    StartTimeout time.Duration `toml:"start_timeout" env:"HP_AGENT_REPO_SERVER_START_TIMEOUT" default:"30s"`
}
// Agent gains RepoServer AgentRepoServer `toml:"repo_server"`
```

## 2. The backend's session

`backupreposervice` gains:

```go
OpenRepoServer(ctx, db, *OpenRepoServerReq) (*RepoServerSession, error)
// OpenRepoServerReq: the repository (RepoTarget) and the user, e.g. hivepaas@data-backup.
// RepoServerSession: URL, Fingerprint, Username, Hostname, Password; Close() error.
```

- **Where it opens.** It opens the stream to the agent of the node of the
  repository's volume. A volume repository pinned to
  no node cannot have a server, and the run fails, saying so.
- **The user's password** is an HMAC-SHA256 of the repository's password, keyed
  by the user's name.
  - It is the same in every session, so two sessions on one repository at
    once do not log each other out.
  - It never leaves the backend and the agent.
- **A new storage kind.** The engine gains a storage of the kind "server". Its
  client connects with the following, from a config file of its own in `/tmp`,
  removed when the session closes:
  - `kopia repository connect server --url=<url>`;
  - `--server-cert-fingerprint=<fingerprint>`;
  - `--override-username` and `--override-hostname` from the user.

**Where each backup runs**, decided in `BackupStream` and `BackupDirectory`:

| Repository | Source | kopia runs |
|---|---|---|
| cloud storage | any | as today |
| volume on node N | a volume on N | as today, through the agent of N |
| volume on node N | a command's output | in the backend, connected to a server on N |
| volume on node N | a volume on node M ≠ N | through the agent of M, connected to a server on N |

- **The two `ERR_NOT_IMPLEMENTED` answers go:** a stream into a volume
  repository, and a directory on another node than its volume repository.
- **Order in a session:**
  1. The snapshot is taken.
  2. The session closes.
  3. The repository's snapshot list is synced, directly on node N, as today.
- **Directly on the repository's node, not through the server:**
  - deleting a failed command's snapshot;
  - the sync;
  - `backup-repo-cleanup`.

  These need the repository's owner, who sees every snapshot and runs
  maintenance.

**The run's log** says the following, through a `Progress` callback on the
backup requests:
- "Starting the repository server on node X";
- "Repository server ready at <url>";
- "Repository server stopped".

**Errors** fail the run with what went wrong:
- the node's agent unreachable;
- the server not starting, with kopia's words;
- the server or its agent gone during the backup - the error says the server
  stopped, and why.

A snapshot left half made is deleted by its run's tag, `hivepaas.run:<task
id>`, as today. Canceling a run, or a timeout, ends the stream, and with it the
server. Two sessions on one repository run two servers; kopia takes several
writers.

**Restore.** `OpenRepoServer` takes the user; it is not bound to the data
backup. Restore, a spec of its own, opens a session and runs `kopia snapshot
restore` from the node that receives the data.

**Dashboard.** No change: the data backup form never refused these cases.

## 3. Testing

- **Go, the agent:**
  - the arguments of `kopia server start`: the address, the port, TLS, the
    session's config file, and `GOMEMLIMIT` in its environment;
  - `server user add`, and `set` when the user exists;
  - the URL and the fingerprint read from the server's output;
  - with a fake process, the process stopped and the session directory removed
    when the stream ends;
  - the start timeout, with kopia's stderr in the error.
- **Go, the backend:**
  - the table of §2, a case each;
  - the user's password: stable, and different for another repository;
  - the "server" storage's connect command;
  - the session closed when the backup fails too;
  - the old `ERR_NOT_IMPLEMENTED` answers gone.
- **Go, with the kopia binary** (skipped without it): a real server on a
  filesystem repository, and then:
  - a stream and a directory snapshotted through it as `hivepaas@data-backup`;
  - another identity refused with `access denied`;
  - no process left once it is stopped.
- **Live, on the Linux server:**
  1. `pg_dump` of a Postgres app into a volume repository.
  2. A volume on a worker into a volume repository on another node.
  3. Two runs into one repository at once.
  4. A run canceled half way: no `kopia server` process left on the node, and no
     half-made snapshot.
  5. The agent of the repository's node restarted during a run: the run fails,
     saying why, and its retry succeeds.
  6. The agent's memory during a run, near the limit.
  7. A snapshot opened with the kopia CLI on the node.

## Not in this version

- Restore.
- A server shared between sessions, or kept running.
- Serving repositories on cloud storage through a server.
- A limit on concurrent servers on a node.
