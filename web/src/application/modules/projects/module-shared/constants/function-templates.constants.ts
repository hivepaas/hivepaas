import { type FunctionFile } from "~/projects/domain";

import { EFunctionRuntime } from "../enums";

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
