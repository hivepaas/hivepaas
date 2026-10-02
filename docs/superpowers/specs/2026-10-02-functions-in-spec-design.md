# Functions in a spec

## Why

A function cannot leave the installation it was made on: spec export leaves it
out (`FUNCTION_SKIPPED`), and import and templates refuse it. Moving a project
to another server, or keeping a copy of it, loses every function and the jobs
that call them.

This makes a function travel through a spec bundle like any app, its code as
files of its own in the bundle, and come back through the same checks creating
a function makes.

**In scope:** export and import of functions, inline or from a repository.
**Not in scope:** an app template holding a function - templates still refuse
one; and the dashboard's export and import screens beyond what they already
show.

## What travels

A function is an app, so what an app carries a function carries: its kind
(`function`), its deployment source (`deployment.source`, with `activeMethod:
function` and `functionSource`), routing, env vars, secrets, config files,
scheduled jobs - `function-invoke` ones among them - and what the swarm service
records. References inside it (git credentials, the registry the image is pushed
to, a job's app) are written as for any app: a path in the bundle, or an
external reference.

A function from a repository carries no code: its `functionSource.code.repo`
says where the code is, as a repository deployment does.

## The code: files of their own

A function's inline code is not written inside the env document. Each file is a
file of the bundle, under the function's directory:

```
projects/<project>/envs/<env>.yaml
projects/<project>/envs/<env>/functions/<app>/index.js
projects/<project>/envs/<env>/functions/<app>/package.json
projects/<project>/envs/<env>/functions/<app>/lib/util.js
```

and the function's source names the directory instead of listing the files:

```yaml
deployment:
  source:
    activeMethod: function
    functionSource:
      runtime: node24
      entrypoint: {file: index.js, handler: default}
      code:
        inline:
          filesFrom: projects/shop/envs/prod/functions/report
      timeout: 30s
      maxConcurrency: 16
      maxBodySize: 6mb
```

- `<app>` is the app's key in the env document; a file's path below it is the
  file's path in the function's code, with its directories.
- `filesFrom` is relative to the bundle's root, and is always the function's
  own directory: `projects/<project>/envs/<env>/functions/<app>`.
- The manifest's `files` lists every one of these files, as it lists every
  payload file.
- A lock file the code holds (`package-lock.json`, `requirements.lock`,
  `go.sum`) is a file like the others.

So the code reads, diffs and is edited as files in a repository that keeps the
bundle unpacked, and the env document stays the size of its settings.

### Writing

The documents are built as today, the function's files inside its source.
Only when the bundle's files are written is each inline function's code taken
out: its files become bundle files, and its `inline` becomes `filesFrom`. Every
other step - the report, the reference paths, an import's comparison of a
bundle with what exists - works on documents whose code is inline, so it does
not change.

### Reading

When a bundle is read, before anything else looks at it, each source with
`filesFrom` gets its files back: every bundle file under that directory becomes
a file of the code, `filesFrom` is dropped, and the source reads as an exported
inline function did. From then on import sees what it would have seen with the
code inline.

A bundle is refused as unreadable - nothing is planned - when:

- `filesFrom` is not `projects/<project>/envs/<env>/functions/<app>` for the
  document and the app it is in;
- the directory holds no file;
- a file's path below it is not a function's file path: letters, digits, `.`,
  `_`, `-` and `/`, no `.` or `..` part, nothing under `.hivepaas`;
- a file is not UTF-8 text.

A file under `functions/` that no source names is ignored, as any file the
reader does not know is today. The limits of inline code - 1 MB, 100 files - are
the function's source's, checked with it (below), not the bundle's.

## Import

### A new function

An app whose kind is `function` is created as a function, through what creating
one checks, and each failure is a blocked issue of the plan on that app:

- the kind and the source agree: kind `function` with `activeMethod: function`
  and a `functionSource`, and no other kind with either;
- the source, normalized and validated as `POST …/apps/function` does it: the
  runtime, the contract, the entrypoint and handler, the inline code's paths and
  limits or the repository, the Debian packages, the timeout, the concurrency,
  the body size;
- the build source, as creating a function checks it: the repository answers at
  its ref with its credentials, and a cluster of several nodes has a registry
  to push the image to.

The function is created with its routing at its runtime's port, as any function
(`fixFunctionRouting`), and its container is fixed by its deployment, as any
function's. Its first deployment follows the import's existing option to deploy
the apps it creates.

### An existing app

- A function onto a function: its settings are updated as any app's; its source
  is normalized and validated as above, and it is deployed when its source
  changed, as any app is.
- A function onto an app that is not one, or an app onto a function: refused,
  as a blocked issue on that app. An app does not change its kind, here as in
  its settings.

### The gap this closes

Today the check that refuses a function's settings runs only when an app is
built from a document, not on the path an import writes settings by. A bundle
written by hand can make a function past every check. With this change, every
imported setting that makes an app a function - its kind, its deployment source
- goes through the checks above, whichever path writes it.

### Jobs that call a function

A `function-invoke` job is an app job of its function, exported and imported
with it; its app is the function, mapped as any job's app is. A function's call
is refused, as when it is saved, where its app is not a function.

## Issue codes

- `FUNCTION_SKIPPED` is no longer written by export.
- New blocked codes on import: `FUNCTION_SOURCE_INVALID` (the source's
  validation failed; the detail names the field and the reason),
  `FUNCTION_KIND_MISMATCH` (kind and source disagree), `APP_KIND_CHANGED` (a
  function onto another kind, or the reverse), `FUNCTION_BUILD_SOURCE` (the
  repository or the registry check failed), `FUNCTION_CALL_ON_APP` (a
  `function-invoke` job on an app that is no function).

## Testing

- Export of an inline function writes its files under its directory and
  `filesFrom` in its source, and lists them in the manifest; a repository
  function's source is written as it is.
- A bundle read back gives the inline source with the same files, paths and
  contents: export then import of the same env plans no change for the
  function.
- A bundle is refused for a `filesFrom` outside the function's directory, an
  empty directory, an unsafe path and a file that is not UTF-8.
- Import creates a function from an inline source and from a repository source,
  with its routing at its port and its first deployment.
- Import blocks a source that does not validate, a kind that disagrees with the
  source, a kind change on an existing app, and a missing registry on a cluster
  of several nodes.
- A `function-invoke` job travels with its function and calls it on the target.
- A hand-made bundle whose deployment source is a function on a non-function app
  is blocked, through the path an import writes settings by.

## Later

- An app template holding a function.

## Changes after the plan

Writing the plan (`docs/superpowers/plans/2026-10-02-functions-in-spec.md`)
settled what follows:

- the repository of an imported function is not reached at import: the
  credentials it needs may be created by the same import, so at plan time there
  is nothing to reach it with; as for a repository app imported today, its first
  build says when it cannot be. `FUNCTION_BUILD_SOURCE` is the registry a
  cluster of several nodes needs;
- the source's checks are written for the entity in the spec service and held to
  the API's by a test, since a service does not import a DTO; the path, handler
  and entrypoint rules they share are in `base`.
