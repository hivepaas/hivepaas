import type { PaginationState, SortingState } from "@infrastructure/data";
import type { AppDataFile } from "~/projects/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type AppDataFiles_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type AppDataFiles_FindManyPaginated_Res = ApiResponsePaginated<AppDataFile>;

export type AppDataFiles_GetDownloadUrl_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    dataFileID: string;
}>;

export type AppDataFiles_GetDownloadUrl_Res = ApiResponseBase<{
    url: string;
}>;

export type AppDataFiles_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    dataFileID: string;
    deletePermanently: boolean;
}>;

export type AppDataFiles_DeleteOne_Res = ApiResponseBase<{
    type: "success";
}>;
