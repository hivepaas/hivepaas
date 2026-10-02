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

/**
 * The language a Node.js function is written in. Node.js runs TypeScript by
 * removing its types as it loads a file.
 */
export const EFunctionLanguage = {
    JavaScript: "javascript",
    TypeScript: "typescript",
} as const;

export type EFunctionLanguage = (typeof EFunctionLanguage)[keyof typeof EFunctionLanguage];

export const FUNCTION_LANGUAGE_LABELS: Record<EFunctionLanguage, string> = {
    [EFunctionLanguage.JavaScript]: "JavaScript",
    [EFunctionLanguage.TypeScript]: "TypeScript",
};

/**
 * The runtimes whose functions may be written in either language.
 */
export const FUNCTION_RUNTIMES_WITH_LANGUAGES: EFunctionRuntime[] = [EFunctionRuntime.Node24];

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

/**
 * Whether the dashboard knows a runtime. A function keeps the id of a runtime a
 * newer HivePaaS added; the dashboard shows it, and leaves it as it is.
 */
export function isKnownFunctionRuntime(runtime: string): runtime is EFunctionRuntime {
    return (ALL_FUNCTION_RUNTIMES as string[]).includes(runtime);
}

/**
 * A runtime's name; a runtime the dashboard does not know goes by its id.
 */
export function functionRuntimeLabel(runtime: string): string {
    return isKnownFunctionRuntime(runtime) ? FUNCTION_RUNTIME_LABELS[runtime] : runtime;
}

/**
 * A runtime's default entrypoint; none for a runtime the dashboard does not know.
 */
export function functionRuntimeDefaultEntrypoint(runtime: string): { file: string; handler: string } | null {
    return isKnownFunctionRuntime(runtime) ? FUNCTION_RUNTIME_DEFAULT_ENTRYPOINT[runtime] : null;
}
