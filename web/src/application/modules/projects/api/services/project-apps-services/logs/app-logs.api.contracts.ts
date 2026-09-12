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
