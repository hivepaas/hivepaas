# The Build Context Goes to the Build Node

A deployment from a repository checks the source out in the app's own
temporary directory, then builds the image on a build node. When the build node
is not the app's node, the build is run by that node's agent, and the agent is
given only the **path** of the checkout (`checkout_dir`). The directory is in
the app's container, on another node: the agent finds nothing there, and the
build fails.

This design sends the checkout itself. The app packs it and streams it to the
agent inside the build call; the agent unpacks it in its own temporary
directory and builds from there. Backward compatibility is not kept.

---

## Decisions

1. **The app keeps checking out.** The checkout, its credentials and the
   repository cache stay where they are; only the result travels.
2. **The whole checkout is sent, not filtered by `.dockerignore`.** The `.git`
   directory is already removed at the end of the checkout, so what is left is
   the files git tracks. Buildx on the agent applies `.dockerignore` as it
   always has, and the Dockerfile generator scans the source on the agent, so
   it must see every file.
3. **One call.** `ImageBuild` is replaced by `ImageBuildFromSource`, a stream
   in both directions: the source travels in the call that builds it, and lives
   as long as that call.
4. **The agent's copy is in a dated temporary directory**,
   `/tmp/hivepaas/<YYYY-MM-DD>/<random>`, as the app's already is, so old days
   can be removed whole.
5. **A build on the app's own node does not change**: no packing, no transfer.

## 1. The call

```proto
service ImageBuildService {
  rpc ImageBuildFromSource(stream ImageBuildMsg) returns (stream ImageBuildResp);
}

message ImageBuildMsg {
  oneof value {
    ImageBuildReq req = 1;   // first, once
    bytes source_chunk = 2;  // then the packed source, in order
  }
}
```

- The old `ImageBuild` rpc is removed. The new one has a name of its own so
  that an agent of an older image answers `Unimplemented` instead of reading
  the new messages as the old request.
- `ImageBuildReq` loses `checkout_dir` and `temp_dir` (fields 10 and 11,
  reserved): the agent chooses its own.
- The app sends the request, then the chunks, then closes its side. The agent
  answers with log frames and the result, as today.
- The source is a **tar stream compressed with zstd** at its fastest level,
  cut in chunks of 1 MiB. It is written to the stream as it is packed: no
  archive file on the app's disk.

## 2. The app's side

In `repoDeployStepImageBuild`, on the agent path only:

- pack `CheckoutDir` and send it: regular files with their mode, directories,
  and symbolic links as links. Anything else (devices, sockets) is skipped;
- log, in the deployment's log, the number of files and the size sent, and the
  time it took;
- when packing itself fails, that is the error the deployment gets: the agent
  only sees the call canceled, and drops what it received;
- the fallback stays: an error within the first 10 seconds builds on the
  current node instead. The checkout is still on the app, so nothing is lost.

An agent that does not know the new call (an older image) answers
`Unimplemented` at once, and the fallback covers it, with a warning in the
deployment's log. App and agent are updated together.

## 3. The agent's side

- **Where:** `fileutil.CreateTempDir(base.BaseTempDirDefault, "*", 0)`, which
  gives `/tmp/hivepaas/<YYYY-MM-DD>/<random>`, mode `0700`. The source is
  unpacked in `<that>/checkout`.
- **Unpacking** reads the stream as it arrives. It refuses, and fails the
  build, on:
  - an entry whose path is absolute or leaves the directory (`..`);
  - an entry written through a symbolic link unpacked earlier;
  - an entry type other than a file, a directory or a symbolic link;
  - an entry whose place is already taken (a duplicate in the stream).

  A symbolic link's target is kept as it is: a link pointing outside is only a
  name, and buildx does not follow links out of its context.
- **Then the build that exists today**, with that directory as `CheckoutDir`:
  preparing or generating the Dockerfile, buildx, the push. Nothing in
  `imagebuildservice` changes.
- **Afterwards** the directory is removed: on success, on failure, and when the
  call is cancelled or the connection drops.

## 4. Temporary directories that are left behind

An agent killed during a build leaves its directory. Two sweeps remove it:

- **when the agent starts**, everything under `/tmp/hivepaas` that is a dated
  directory goes: no build is running yet;
- **at each node cleanup** (when the cluster cleanup is enabled), dated
  directories older than 3 days go, the rule the app's own cleanup applies
  (`sysCleanupTempFiles`). The two share one
  function, `fileutil.RemoveDatedTempDirs(baseDir, before)`.

Directories under `/tmp/hivepaas` that are not dates (`backup-repos`) are not
touched.

## 5. What this does not change

- **The checkout and the repository cache**, on the app, on the project's
  volume.
- **Which node builds**, and the waiting for a build slot.
- **The build itself**, the image tags, the push to a registry.
- **Cancelling a running build** when the user cancels the deployment: still
  not done (`repoDeployOnCommand`), a separate piece of work.

## 6. Testing

**Go, packing and unpacking**, on real temporary directories:
- a tree with a nested directory, an executable file, an empty directory and a
  symbolic link arrives the same;
- an entry with `../`, an absolute path, a file written through a symbolic
  link, and a device entry are each refused, and nothing is written outside.

**Go, the call**, over a real gRPC connection:
- the agent's build is given a directory holding the source the app sent, and
  the logs and result come back;
- the directory is gone after a build that succeeds, one that fails, and one
  whose caller cancels half way;
- a source larger than one chunk arrives whole.

**Go, the sweeps:** dated directories older than the threshold go, newer ones
and non-dated ones stay.

**Live, on the Linux server:** a deployment from a repository with the build
node set to another node, with a Dockerfile in the repository and with a
generated one.

## Later

- **Cancelling the build** on the agent when the deployment is cancelled.
- **Filtering by `.dockerignore`** before sending, if a repository with large
  tracked files that are ignored makes the transfer matter.
