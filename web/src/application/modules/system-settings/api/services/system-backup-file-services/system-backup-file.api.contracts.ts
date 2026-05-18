import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SystemBackupFile } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type SystemBackupFile_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type SystemBackupFile_FindManyPaginated_Res = ApiResponsePaginated<SystemBackupFile>;

export type SystemBackupFile_FindOneById_Req = ApiRequestBase<{
    fileID: string;
}>;

export type SystemBackupFile_FindOneById_Res = ApiResponseBase<SystemBackupFile>;

export type SystemBackupFile_DownloadOne_Req = ApiRequestBase<{
    fileID: string;
}>;

export type SystemBackupFile_DownloadOne_Res = ApiResponseBase<{
    blob: Blob;
    filename?: string;
}>;
