import type { PaginationState, SortingState } from "@infrastructure/data";
import type { AppScheduledJobs_Command_Payload } from "~/projects/api/services/project-apps-services/scheduled-jobs";
import type { AppDataFile } from "~/projects/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type AppDataFiles_UploadLocal_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    fileKind: string;
    files: File[];
}>;

export type AppDataFiles_UploadLocal_Res = ApiResponseBase<{
    files: AppDataFile[];
}>;

export type AppDataFiles_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    fileKind: string;
    filePath: string;
    storageID: string;
    bucket?: string;
}>;

export type AppDataFiles_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type AppDataFiles_FindManyPaginated_Res = ApiResponsePaginated<AppDataFile>;

export type AppDataFiles_GetDownloadUrl_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    dataFileID: string;
}>;

export type AppDataFiles_GetDownloadUrl_Res = ApiResponseBase<{
    url: string;
}>;

export type AppDataFiles_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    dataFileID: string;
    deletePermanently: boolean;
}>;

export type AppDataFiles_DeleteOne_Res = ApiResponseBase<{
    type: "success";
}>;

export type AppDataFiles_CreateOne_Res = ApiResponseBase<{
    id: string;
}>;

export type AppDataFiles_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    dataFileID: string;
}>;

export type AppDataFiles_FindOneById_Res = ApiResponseBase<AppDataFile>;

/** What a data file is loaded into: a command reading it on its stdin, and the passphrase of one saved encrypted. */
export type AppDataFiles_Load_Payload = {
    command: AppScheduledJobs_Command_Payload;
    passphrase?: string;
};

export type AppDataFiles_Load_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    dataFileID: string;
    payload: AppDataFiles_Load_Payload;
}>;

export type AppDataFiles_Load_Res = ApiResponseBase<{ taskId: string }>;
