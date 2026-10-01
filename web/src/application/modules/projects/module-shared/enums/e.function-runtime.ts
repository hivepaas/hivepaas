/**
 * The runtimes a function runs on, as HivePaaS releases their images.
 */
export const EFunctionRuntime = {
    Node24: "node24",
    Python313: "python313",
    Go127: "go127",
} as const;

export type EFunctionRuntime = (typeof EFunctionRuntime)[keyof typeof EFunctionRuntime];

export const ALL_FUNCTION_RUNTIMES: EFunctionRuntime[] = [
    EFunctionRuntime.Node24,
    EFunctionRuntime.Python313,
    EFunctionRuntime.Go127,
];

export const FUNCTION_RUNTIME_LABELS: Record<EFunctionRuntime, string> = {
    [EFunctionRuntime.Node24]: "Node.js 24",
    [EFunctionRuntime.Python313]: "Python 3.13",
    [EFunctionRuntime.Go127]: "Go 1.27",
};

/**
 * Where the runtime finds the handler when a function does not say: the file
 * (for Go, the package's directory) and the handler's name.
 */
export const FUNCTION_RUNTIME_DEFAULT_ENTRYPOINT: Record<EFunctionRuntime, { file: string; handler: string }> = {
    [EFunctionRuntime.Node24]: { file: "index.js", handler: "default" },
    [EFunctionRuntime.Python313]: { file: "main.py", handler: "handler" },
    [EFunctionRuntime.Go127]: { file: ".", handler: "Handle" },
};
