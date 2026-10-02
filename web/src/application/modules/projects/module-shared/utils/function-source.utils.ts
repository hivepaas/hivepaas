import { type AppDeploymentSettings_UpdateOne_Req, type FunctionSourcePayload } from "~/projects/api/services";
import { APP_CATEGORY_FUNCTION, type FunctionFile, type FunctionMethod, type FunctionSource } from "~/projects/domain";

/**
 * Whether an app is a function: its kind's category says so.
 */
export function isFunctionApp(app: { category: string }): boolean {
    return app.category === APP_CATEGORY_FUNCTION;
}

/**
 * The directory HivePaaS keeps in a function's code, for its Dockerfile.
 */
export const FUNCTION_RESERVED_DIR = ".hivepaas";

const FUNCTION_FILE_PATH_PATTERN = /^[A-Za-z0-9._/-]+$/;

/**
 * Why a path cannot be a file of the function's code, or null when it can:
 * relative, plain, inside the function, not HivePaaS's, not taken.
 */
export function functionFilePathProblem(path: string, takenPaths: string[]): string | null {
    const parts = path.split("/");
    if (path === "" || !FUNCTION_FILE_PATH_PATTERN.test(path) || path.startsWith("/")) {
        return "Use letters, digits, '.', '_', '-' and '/'";
    }
    if (parts.some(part => part === "" || part === "." || part === "..")) {
        return "A path stays inside the function";
    }
    if (parts[0] === FUNCTION_RESERVED_DIR) {
        return `${FUNCTION_RESERVED_DIR} is HivePaaS's`;
    }
    if (takenPaths.includes(path)) {
        return "There is a file of that name";
    }
    return null;
}

export type CodeLanguage = "javascript" | "typescript" | "python" | "go" | "json" | "plain";

/**
 * The language a file of a function is highlighted as, by its extension.
 */
export function languageOfPath(path: string): CodeLanguage {
    const extension = path.slice(path.lastIndexOf(".") + 1).toLowerCase();
    switch (extension) {
        case "js":
        case "mjs":
        case "cjs":
            return "javascript";
        case "ts":
        case "mts":
        case "cts":
            return "typescript";
        case "py":
            return "python";
        case "go":
        case "mod":
        case "sum":
            return "go";
        case "json":
            return "json";
        default:
            return "plain";
    }
}

/**
 * A function's source as the API takes it back: the settings it refers to by
 * id, its inline files replaced when files are given.
 */
export function functionSourceToPayload(source: FunctionSource, files?: FunctionFile[]): FunctionSourcePayload {
    const inline = files ? { files } : source.code.inline;

    return {
        runtime: source.runtime,
        contract: source.contract,
        entrypoint: source.entrypoint,
        code: {
            ...(inline ? { inline: { files: inline.files } } : {}),
            ...(source.code.repo && !files
                ? {
                      repo: {
                          repoType: source.code.repo.repoType,
                          repoUrl: source.code.repo.repoUrl,
                          repoRef: source.code.repo.repoRef,
                          commitHash: source.code.repo.commitHash,
                          credentials: { id: source.code.repo.credentials?.id ?? "" },
                      },
                  }
                : {}),
            dir: files ? "" : source.code.dir,
        },
        systemPackages: source.systemPackages,
        timeout: source.timeout,
        maxConcurrency: source.maxConcurrency,
        maxBodySize: source.maxBodySize,
        pushToRegistry: { id: source.pushToRegistry?.id ?? "" },
    };
}

type UpdatePayload = AppDeploymentSettings_UpdateOne_Req["data"]["payload"];

/**
 * A function's deployment settings as the API takes them back, with its source
 * replaced: what saving the function's code or its settings sends.
 */
export function functionSettingsToPayload(settings: FunctionMethod, source: FunctionSourcePayload): UpdatePayload {
    const { notification } = settings;

    return {
        activeMethod: settings.activeMethod,
        command: settings.command ?? "",
        workingDir: settings.workingDir ?? "",
        preDeploymentCommand: settings.preDeploymentCommand ?? "",
        postDeploymentCommand: settings.postDeploymentCommand ?? "",
        notification: {
            successUseDefault: notification?.successUseDefault ?? true,
            ...(notification && !notification.successUseDefault && notification.success
                ? { success: { id: notification.success.id } }
                : {}),
            failureUseDefault: notification?.failureUseDefault ?? true,
            ...(notification && !notification.failureUseDefault && notification.failure
                ? { failure: { id: notification.failure.id } }
                : {}),
        },
        functionSource: source,
    };
}
