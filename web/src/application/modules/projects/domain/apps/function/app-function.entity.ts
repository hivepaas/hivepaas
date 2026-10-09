import type { SettingsBaseEntity } from "~/settings/domain";

/**
 * The kind category of an app that is a function.
 */
export const APP_CATEGORY_FUNCTION = "function";

export type FunctionFile = {
    path: string;
    content: string;
};

export type FunctionRepoCode = {
    repoType: string;
    repoUrl: string;
    repoRef: string;
    commitHash: string;
    credentials: SettingsBaseEntity | null;
    /** Whether a push to repoRef, received by a repo webhook, deploys the function. */
    autoDeploy: boolean;
};

/**
 * What a function is deployed from: its code, inline or in a repository, the
 * runtime that runs it, and the limits of one call.
 */
export type FunctionSource = {
    /** A runtime's id; one the dashboard does not know is kept as it is. */
    runtime: string;
    contract: string;
    entrypoint: { file: string; handler: string };
    code: {
        inline: { files: FunctionFile[] } | null;
        repo: FunctionRepoCode | null;
        dir: string;
    };
    systemPackages: string[];
    /** A duration, as "30s" or "2m". */
    timeout: string;
    maxConcurrency: number;
    /** A size, as "6MB". */
    maxBodySize: string;
    pushToRegistry: SettingsBaseEntity | null;
};

/**
 * The request a test run calls the function with. Its body is text.
 */
export type FunctionTestRequest = {
    method: string;
    path: string;
    query: Record<string, string[]>;
    headers: Record<string, string[]>;
    body: string;
};

/**
 * How a test run ended: as the runtime says (ok, error, timeout), or as
 * HivePaaS saw it.
 */
export type FunctionTestOutcome =
    "ok" | "error" | "timeout" | "libraries-failed" | "not-loaded" | "bad-request" | "killed" | "no-result";

export type FunctionTestRunResult = {
    /** A FunctionTestOutcome, or an outcome a later backend adds. */
    outcome: string;
    status: number;
    headers: Record<string, string[]>;
    /** The response's body, decoded from base64. */
    body: Uint8Array;
    bodyTruncated: boolean;
    requestId: string;
    durationMs: number;
    logs: string;
    logsTruncated: boolean;
    error: string;
    exitCode: number;
    librariesBuilt: boolean;
    librariesLog: string;
    lockFiles: FunctionFile[];
};
