import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SystemTask, SystemTaskJobName, SystemTaskStatus } from "~/system-status/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type SystemTaskLogFrameType = "in" | "out" | "err" | "warn" | "debug";

export interface SystemTaskLogFrame {
    type: SystemTaskLogFrameType;
    data: string;
    ts: Date | null;
}

export type SystemTasks_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
    jobName?: SystemTaskJobName;
    targetId?: string[];
    status?: SystemTaskStatus[];
}>;

export type SystemTasks_FindManyPaginated_Res = ApiResponsePaginated<SystemTask>;

export type SystemTasks_FindOneById_Req = ApiRequestBase<{
    taskID: string;
}>;

export type SystemTasks_FindOneById_Res = ApiResponseBase<SystemTask>;

export type SystemTasks_GetLogs_Req = ApiRequestBase<{
    taskID: string;
    follow?: boolean;
    tail?: number;
    since?: Date;
    duration?: string;
    timestamps?: boolean;
}>;

export type SystemTasks_GetLogs_Res = ApiResponseBase<SystemTaskLogFrame[]>;

export type SystemTasks_Cancel_Req = ApiRequestBase<{
    taskID: string;
}>;

export type SystemTasks_Cancel_Res = ApiResponseBase<{
    canceled: boolean;
}>;
