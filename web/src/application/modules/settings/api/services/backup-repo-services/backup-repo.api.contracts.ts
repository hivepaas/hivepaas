import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingBackupRepo } from "~/settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type BackupRepo_UpdateStatus_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    inheritable?: boolean;
    default?: boolean;
};

export type BackupRepo_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type BackupRepo_FindManyPaginated_Res = ApiResponsePaginated<SettingBackupRepo>;

export type BackupRepo_FindOneById_Req = ApiRequestBase<{
    id: string;
}>;

export type BackupRepo_FindOneById_Res = ApiResponseBase<SettingBackupRepo>;

export type BackupRepo_UpdateStatus_Req = ApiRequestBase<{
    id: string;
    payload: BackupRepo_UpdateStatus_Payload;
}>;

export type BackupRepo_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type BackupRepo_DeleteOne_Req = ApiRequestBase<{
    id: string;
}>;

export type BackupRepo_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
