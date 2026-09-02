import type { PaginationState, SortingState } from "@infrastructure/data";
import type { BackupRepo_UpdateStatus_Payload } from "~/settings/api/services/backup-repo-services";
import type { SettingBackupRepo } from "~/settings/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectBackupRepo_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectBackupRepo_FindManyPaginated_Res = ApiResponsePaginated<SettingBackupRepo>;

export type ProjectBackupRepo_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    id: string;
}>;

export type ProjectBackupRepo_FindOneById_Res = ApiResponseBase<SettingBackupRepo>;

export type ProjectBackupRepo_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    id: string;
    payload: BackupRepo_UpdateStatus_Payload;
}>;

export type ProjectBackupRepo_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectBackupRepo_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    id: string;
}>;

export type ProjectBackupRepo_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
