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
    appID: string;
}>;

export type AppLogs_GetInfo_Res = ApiResponseBase<{
    tasks: AppLogTask[];
}>;

export type AppLogs_GetToken_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
}>;

export type AppLogs_GetToken_Res = ApiResponseBase<{
    token: string;
}>;

export type AppLogs_GetLogs_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    taskId?: string;
    follow?: boolean;
    tail?: number;
    since?: Date;
    duration?: string;
    timestamps?: boolean;
}>;

export type AppLogs_GetLogs_Res = ApiResponseBase<AppLogFrame[]>;
