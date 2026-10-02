# TypeScript functions

## Why

A Node.js function is written in JavaScript today. Node.js 24 runs TypeScript
itself: it removes the types from a `.ts` file as it loads it, with no build
step and no compiler in the image. A function can be written in TypeScript as
soon as HivePaaS lets its entrypoint be a `.ts` file.

This was checked on the released image
`function-runtime-node24:1.0.0` (Node.js 24.21.0): `hivepaas-runtime invoke`
with `HP_FN_ENTRYPOINT=index.ts` called a handler that imports another `.ts`
file and a type with `import type`, and answered 200; a file with an `enum`
exited 3, "the handler could not be loaded", the line at fault shown.

**In scope:** TypeScript on the `node24` runtime - the API, the dashboard, the
contract and its conformance suite. **Not in scope:** checking types; a build
step that compiles TypeScript; a package of the handler's types.

## What TypeScript is here

- **No new runtime.** A TypeScript function is a `node24` function whose
  entrypoint is a `.ts`, `.mts` or `.cts` file. The runtime, the image and the
  contract version do not change; no image is released for this.
- **Node.js removes the types, nothing else.** Only syntax that can be erased
  is accepted:
  - no `enum`, no `namespace` holding values, no parameter properties
    (`constructor(private x: number)`), no `import x = require(...)`, no
    decorators;
  - a file using one of them cannot be loaded: the call answers 500 with
    outcome `not_loaded`, the health check 503, and the error names the line.
- **Types are not checked.** A type error runs as JavaScript would.
- **Imports name the file with its extension:** `import { greet } from
  "./lib/util.ts"`. A type imported from another file is imported with
  `import type`: an import of a name that exists only as a type fails when the
  module loads.
- **The module system** is the file's: `.mts` is an ES module, `.cts`
  CommonJS, `.ts` follows `"type"` in `package.json`, as `.js` does.
- **Libraries are JavaScript.** Node.js does not run `.ts` files under
  `node_modules`; a library ships its JavaScript, as every published package
  does.
- `tsconfig.json` is not read: its `paths` and its compiler options do not
  apply. A function may keep one for its author's editor.

## The API

`node24`'s entrypoint may be a `.js`, `.mjs`, `.cjs`, `.ts`, `.mts` or `.cts`
file (`base.FunctionEntrypointExts`). Creating a function, saving its
settings and importing it from a spec bundle all check the entrypoint through
that one list. The default entrypoint stays `index.js`.

## The dashboard

**Create a function.** With Node.js 24, a **Language** choice: JavaScript (the
default) or TypeScript.

- TypeScript starts from its own template: `package.json` (`"type":
  "module"`), and `index.ts`, the handler with the request, context and
  response it is given typed in the file:

  ```ts
  // A function answers one request: it returns { status, headers, body }.
  // A body that is an object is sent as JSON. Node.js runs this file by
  // removing its types: they are not checked, and syntax that cannot be
  // removed - enum, namespace - is not allowed.
  interface Request {
      method: string;
      path: string;
      query: Record<string, string | undefined>;
      queryAll: Record<string, string[] | undefined>;
      headers: Record<string, string | undefined>;
      body: Buffer;
      text(): string;
      json(): unknown;
  }

  interface Context {
      requestId: string;
      deadline: number;
      signal: AbortSignal;
      log(...args: unknown[]): void;
  }

  interface Response {
      status?: number;
      headers?: Record<string, string | string[]>;
      body?: unknown;
  }

  export default async function (req: Request, ctx: Context): Promise<Response> {
      ctx.log(`${req.method} ${req.path}`);
      const name = req.query.name ?? "world";

      return { status: 200, body: { hello: name } };
  }
  ```

- The function is created with its entrypoint `index.ts`, from the template
  or from a repository; a repository's entrypoint is changed in the function's
  settings as any other.
- The choice is not stored: a function is TypeScript because its entrypoint
  is.

**The editor** highlights `.ts`, `.mts` and `.cts` files as TypeScript
(Prism's `typescript` grammar); it highlights them as JavaScript today.

**The function's settings** need nothing new: the entrypoint field takes
`index.ts`.

## The contract and its suite (function-runtimes)

- `CONTRACT.md`: `HP_FN_ENTRYPOINT` for Node.js is a `.js`, `.mjs`, `.cjs`,
  `.ts`, `.mts` or `.cts` file; a section "TypeScript" under "The handler"
  says what this spec's "What TypeScript is here" says, in a few lines.
- The conformance suite, for `node24`:
  - the fixture gains `typed.ts`, which imports a `.ts` file of its own and a
    type with `import type`; called through `serve` and `invoke` with
    `HP_FN_ENTRYPOINT=typed.ts`, it answers as the JavaScript fixture does;
  - `enum.ts`, a handler with an `enum`: `invoke` exits 3 and says the handler
    could not be loaded.
- `README.md`'s table says `node24` runs JavaScript and TypeScript.

The suite runs on the next tag of function-runtimes, so a later image cannot
lose TypeScript unnoticed.

## Testing

- The API accepts `index.ts`, `src/handler.mts`, `main.cts` for `node24`, and
  refuses `index.ts` for `python313`; the spec import's parity test covers the
  same list.
- Create Function with TypeScript sends the TypeScript template and the
  entrypoint `index.ts`; with JavaScript, what it sends today.
- `languageOfPath` gives `typescript` for `.ts`, `.mts`, `.cts`.
- The template is called with the `node24` image, through `invoke`, and
  answers `{"hello":"Ada"}` for `?name=Ada`, as the other templates were.
- The conformance tests above, on `node24`.
- The dashboard has no unit test runner: its part is checked by `lint:ci` and
  `build`, and in the browser by the user.

## Later

- Checking types in the build (`tsc --noEmit`), failing it on a type error.
- A package of the handler's types (`@hivepaas/function`), in place of the
  template's interfaces.
- TypeScript's `--experimental-transform-types`, for `enum` and its kind, once
  Node.js makes it stable.
