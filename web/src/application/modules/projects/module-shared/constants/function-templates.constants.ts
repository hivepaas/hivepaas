import { type FunctionFile } from "~/projects/domain";

import { EFunctionLanguage, EFunctionRuntime } from "../enums";

/**
 * The code a new function starts from, per runtime: a handler that answers
 * `?name=Ada` with `{"hello":"Ada"}`. Each was built and called with its
 * runtime's image.
 */
export const FUNCTION_TEMPLATES: Record<EFunctionRuntime, FunctionFile[]> = {
    [EFunctionRuntime.Node24]: [
        {
            path: "package.json",
            content: `{
  "type": "module"
}
`,
        },
        {
            path: "index.js",
            content: `// A function answers one request: it returns { status, headers, body }.
// A body that is an object is sent as JSON.
export default async function (req, ctx) {
    ctx.log(\`\${req.method} \${req.path}\`);
    const name = req.query.name ?? "world";

    return { status: 200, body: { hello: name } };
}
`,
        },
    ],
    [EFunctionRuntime.Python313]: [
        {
            path: "main.py",
            content: `# A function answers one request: it returns a dict of status, headers, body.
# A body that is a dict is sent as JSON.
def handler(req, ctx):
    ctx.log(f"{req.method} {req.path}")
    name = req.query.get("name", "world")

    return {"status": 200, "body": {"hello": name}}
`,
        },
    ],
    [EFunctionRuntime.Go127]: [
        {
            path: "go.mod",
            content: `module example.com/function

go 1.27
`,
        },
        {
            path: "handler.go",
            content: `package function

import (
\t"context"

\t"github.com/hivepaas/function-runtimes/hivepaas"
)

// Handle answers one request.
func Handle(ctx context.Context, req *hivepaas.Request) (*hivepaas.Response, error) {
\thivepaas.Log(ctx, "%s %s", req.Method, req.Path)
\tname := req.Query.Get("name")
\tif name == "" {
\t\tname = "world"
\t}

\treturn hivepaas.JSON(200, map[string]string{"hello": name})
}
`,
        },
    ],
};

/**
 * A Node.js function in TypeScript: Node.js removes the types as it loads the
 * file, so the handler's request, context and response are typed in it.
 */
export const FUNCTION_TYPESCRIPT_TEMPLATE: FunctionFile[] = [
    {
        path: "package.json",
        content: `{
  "type": "module"
}
`,
    },
    {
        path: "index.ts",
        content: `// A function answers one request: it returns { status, headers, body }.
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
    ctx.log(\`\${req.method} \${req.path}\`);
    const name = req.query.name ?? "world";

    return { status: 200, body: { hello: name } };
}
`,
    },
];

/**
 * What a new function starts from: its files and its entrypoint, empty for the
 * runtime's default. TypeScript is for the runtimes that take either language.
 */
export function functionTemplateOf(
    runtime: EFunctionRuntime,
    language: EFunctionLanguage,
): { files: FunctionFile[]; entrypoint: string } {
    if (language === EFunctionLanguage.TypeScript && runtime === EFunctionRuntime.Node24) {
        return { files: FUNCTION_TYPESCRIPT_TEMPLATE, entrypoint: "index.ts" };
    }
    return { files: FUNCTION_TEMPLATES[runtime], entrypoint: "" };
}
