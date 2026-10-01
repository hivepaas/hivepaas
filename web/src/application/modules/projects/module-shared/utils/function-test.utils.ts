import type { FunctionFile, FunctionTestRequest } from "~/projects/domain";

/**
 * A test run's request from what the panel holds: a path with its query, and
 * headers written one per line as "name: value".
 */
export function buildTestRequest(
    method: string,
    pathWithQuery: string,
    headerLines: string,
    body: string,
): FunctionTestRequest {
    const trimmed = pathWithQuery.trim();
    const queryStart = trimmed.indexOf("?");
    const rawPath = queryStart < 0 ? trimmed : trimmed.slice(0, queryStart);
    // A query can hold "?" itself: it is everything after the first one.
    const rawQuery = queryStart < 0 ? "" : trimmed.slice(queryStart + 1);
    const query: Record<string, string[]> = {};
    for (const [name, value] of new URLSearchParams(rawQuery)) {
        (query[name] ??= []).push(value);
    }

    const headers: Record<string, string[]> = {};
    for (const line of headerLines.split("\n")) {
        const separator = line.indexOf(":");
        if (separator <= 0) {
            continue;
        }
        const name = line.slice(0, separator).trim().toLowerCase();
        if (name !== "") {
            (headers[name] ??= []).push(line.slice(separator + 1).trim());
        }
    }

    return {
        method,
        path: rawPath.startsWith("/") ? rawPath : `/${rawPath}`,
        query,
        headers,
        body: method === "GET" || method === "HEAD" ? "" : body,
    };
}

/**
 * A response's body to show: JSON indented, text as it is, or how many bytes
 * of something else.
 */
export function describeBody(body: Uint8Array, headers: Record<string, string[]>): { text: string; isJson: boolean } {
    if (body.length === 0) {
        return { text: "", isJson: false };
    }
    let text: string;
    try {
        text = new TextDecoder("utf-8", { fatal: true }).decode(body);
    } catch {
        return { text: `${body.length} bytes that are not text`, isJson: false };
    }
    const contentType = headers["content-type"]?.[0] ?? "";
    if (contentType.includes("json")) {
        try {
            return { text: JSON.stringify(JSON.parse(text), null, 2), isJson: true };
        } catch {
            return { text, isJson: false };
        }
    }
    return { text, isJson: false };
}

/**
 * The code with the lock files a run made: each added, or replacing the file
 * of its name.
 */
export function withLockFiles(files: FunctionFile[], lockFiles: FunctionFile[]): FunctionFile[] {
    const lockPaths = new Set(lockFiles.map(file => file.path));
    return [...files.filter(file => !lockPaths.has(file.path)), ...lockFiles];
}
