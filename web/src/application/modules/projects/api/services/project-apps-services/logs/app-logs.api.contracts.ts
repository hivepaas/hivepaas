import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type AppLogFrameType = "in" | "out" | "err" | "warn" | "debug";

export interface AppLogFrame {
    type: AppLogFrameType;
    data: string;
    ts: Date | null;
}

export interface AppLogTask {
    id: string;
}

export type AppLogs_GetInfo_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
}>;

export type AppLogs_GetInfo_Res = ApiResponseBase<{
    enabled: boolean;
    tasks: AppLogTask[];
    /** Whether stored logs can be shown, and if not why. */
    history: AppLogHistoryInfo;
}>;

export type AppLogs_GetLogs_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    taskId?: string;
    follow?: boolean;
    tail?: number;
    since?: Date;
    duration?: string;
    timestamps?: boolean;
}>;

export type AppLogs_GetLogs_Res = ApiResponseBase<AppLogFrame[]>;

export type AppLogHistoryReason =
    | "disabled"
    | "apps-not-collected"
    | "no-query-endpoint"
    | "driver-unreadable"
    | "identity-missing";

export interface AppLogHistoryInfo {
    available: boolean;
    reason: AppLogHistoryReason | null;
    /**
     * How far back stored logs reach, as a duration such as `30d`. Undefined
     * when the backend is not one HivePaaS keeps, and the depth is then
     * unknown rather than unlimited.
     */
    retention?: string;
}

export type AppLogs_GetHistory_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    start?: Date;
    /** A Date, or a `nextEnd` passed back verbatim - it has nanoseconds a Date would lose. */
    end?: Date | string;
    limit?: number;
    search?: string;
    /** Reads `search` as a regular expression. Slower: the backend reads it row
     *  by row instead of from its index, and a malformed one is rejected. */
    regex?: boolean;
    matchCase?: boolean;
    levels?: string[];
    streams?: string[];
}>;

export type AppLogs_GetHistory_Res = ApiResponseBase<{
    /** Oldest first. */
    logs: AppLogFrame[];
    truncated: boolean;
    /** What to pass as `end` for the page before this one; null on the last page. */
    nextEnd: string | null;
}>;

/** The ranges a function's metrics are asked for. */
export type FunctionMetricsRange = "1h" | "6h" | "24h" | "7d";

export type AppLogs_GetFunctionMetrics_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    range: FunctionMetricsRange;
}>;

/** A set of calls: how many, how many failed, how many answered 4xx and 5xx,
 *  and the handler's duration in milliseconds - null without a call. */
export interface FunctionMetricsCounts {
    calls: number;
    failed: number;
    errors4xx: number;
    errors5xx: number;
    p50: number | null;
    p95: number | null;
    p99: number | null;
}

export interface FunctionMetricsPoint extends FunctionMetricsCounts {
    time: string;
}

/** The calls of one method and path. */
export interface FunctionMetricsPath extends FunctionMetricsCounts {
    method: string;
    path: string;
}

export type AppLogs_GetFunctionMetrics_Res = ApiResponseBase<{
    /** False when the logs cannot be read; reason says why. */
    available: boolean;
    reason: AppLogHistoryReason | null;
    range: FunctionMetricsRange;
    start: string | null;
    end: string | null;
    stepSeconds: number;
    /** The logs are kept for less than the range: it starts where they do. */
    clamped: boolean;
    totals: FunctionMetricsCounts | null;
    byOutcome: Record<string, number>;
    /** One point per step, oldest first. */
    series: FunctionMetricsPoint[];
    /** The most called methods and paths, the most called first. */
    byPath: FunctionMetricsPath[];
}>;
