import type { PaginationState, SortingState } from "@infrastructure/data";
import type { AppScheduledJob, AppScheduledJobTask, AppScheduledJobTaskLogFrame } from "~/projects/domain";
import type {
    EAppScheduledJobArgSeparator,
    EAppScheduledJobTaskPriority,
    EAppScheduledJobType,
} from "~/projects/module-shared/enums";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type AppScheduledJobs_RefObject_Payload = {
    id: string;
};

export type AppScheduledJobs_Schedule_Payload = {
    cronExpr?: string;
    interval?: string;
    initialTime?: Date;
    endTime?: Date;
};

export type AppScheduledJobs_EnvVar_Payload = {
    key: string;
    value: string;
};

export type AppScheduledJobs_CommandArg_Payload = {
    use: boolean;
    name: string;
    value: string;
};

export type AppScheduledJobs_CommandArgGroup_Payload = {
    enabled: boolean;
    exportEnv: string;
    separator: EAppScheduledJobArgSeparator;
    args: AppScheduledJobs_CommandArg_Payload[];
};

export type AppScheduledJobs_CommandConsoleSize_Payload = {
    width: number;
    height: number;
};

export type AppScheduledJobs_Command_Payload = {
    runInShell?: string;
    command: string;
    script: string;
    workingDir?: string;
    consoleSize: AppScheduledJobs_CommandConsoleSize_Payload;
    tty: boolean;
    envVars?: AppScheduledJobs_EnvVar_Payload[];
    argGroups?: AppScheduledJobs_CommandArgGroup_Payload[];
};

export type AppScheduledJobs_Notification_Payload = {
    successUseDefault: boolean;
    success?: AppScheduledJobs_RefObject_Payload;
    failureUseDefault: boolean;
    failure?: AppScheduledJobs_RefObject_Payload;
};

export type AppScheduledJobs_CommandOutput_SaveToFile_Payload = {
    fileName: string;
    filePath?: string;
    fileKind?: string;
    storage?: AppScheduledJobs_RefObject_Payload;
    bucket?: string;
    compressionFormat: string;
    encryptionFormat: string;
    encryptionSecret?: string;
};

export type AppScheduledJobs_CommandOutput_PipeToApp_Payload = {
    targetApp: AppScheduledJobs_RefObject_Payload;
    command: AppScheduledJobs_Command_Payload;
};

export type AppScheduledJobs_CommandOutput_Payload = {
    enabled: boolean;
    saveToFile?: AppScheduledJobs_CommandOutput_SaveToFile_Payload | null;
    pipeToApp?: AppScheduledJobs_CommandOutput_PipeToApp_Payload | null;
};

export type AppScheduledJobs_Upsert_Payload = {
    availableInProjects: boolean;
    default: boolean;
    name: string;
    jobType: EAppScheduledJobType;
    schedule: AppScheduledJobs_Schedule_Payload;
    app: AppScheduledJobs_RefObject_Payload;
    priority: EAppScheduledJobTaskPriority;
    maxRetry?: number;
    retryDelay?: string;
    retryDelayIncr?: string;
    retryBackoff: boolean;
    retryDelayMax?: string;
    timeout?: string;
    controlDisabled: boolean;
    command: AppScheduledJobs_Command_Payload;
    commandOutput?: AppScheduledJobs_CommandOutput_Payload;
    notification: AppScheduledJobs_Notification_Payload;
};

export type AppScheduledJobs_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type AppScheduledJobs_FindManyPaginated_Res = ApiResponsePaginated<AppScheduledJob>;

export type AppScheduledJobs_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    scheduledJobID: string;
}>;

export type AppScheduledJobs_FindOneById_Res = ApiResponseBase<AppScheduledJob>;

export type AppScheduledJobs_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    payload: AppScheduledJobs_Upsert_Payload;
}>;

export type AppScheduledJobs_CreateOne_Res = ApiResponseBase<{
    id: string;
}>;

export type AppScheduledJobs_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    scheduledJobID: string;
    payload: AppScheduledJobs_Upsert_Payload & {
        updateVer: number;
    };
}>;

export type AppScheduledJobs_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;

export type AppScheduledJobs_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    scheduledJobID: string;
    payload: {
        updateVer: number;
        status: ESettingStatus;
        expireAt: Date | null;
        availableInProjects: boolean;
        default: boolean;
    };
}>;

export type AppScheduledJobs_UpdateStatus_Res = ApiResponseBase<{
    type: "success";
}>;

export type AppScheduledJobs_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    scheduledJobID: string;
}>;

export type AppScheduledJobs_DeleteOne_Res = ApiResponseBase<{
    type: "success";
}>;

export type AppScheduledJobs_RunNow_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    scheduledJobID: string;
}>;

export type AppScheduledJobs_RunNow_Res = ApiResponseBase<{
    task: {
        id: string;
    };
}>;

export type AppScheduledJobTasks_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    scheduledJobID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type AppScheduledJobTasks_FindManyPaginated_Res = ApiResponsePaginated<AppScheduledJobTask>;

export type AppScheduledJobTasks_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    scheduledJobID: string;
    taskID: string;
}>;

export type AppScheduledJobTasks_FindOneById_Res = ApiResponseBase<AppScheduledJobTask>;

export type AppScheduledJobTasks_GetLogs_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    scheduledJobID: string;
    taskID: string;
    follow?: boolean;
    tail?: number;
    since?: Date;
    duration?: string;
}>;

export type AppScheduledJobTasks_GetLogs_Res = ApiResponseBase<AppScheduledJobTaskLogFrame[]>;

export type AppScheduledJobTasks_Cancel_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    scheduledJobID: string;
    taskID: string;
}>;

export type AppScheduledJobTasks_Cancel_Res = ApiResponseBase<{
    type: "success";
}>;
