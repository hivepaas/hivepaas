import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingBackupRepo } from "~/settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type BackupRepo_Retention_Payload = {
    keepLast?: number;
    keepHourly?: number;
    keepDaily?: number;
    keepWeekly?: number;
    keepMonthly?: number;
};

export type BackupRepo_CreateOne_Payload = {
    name: string;
    engine: string;
    importExisting?: boolean;
    description?: string;
    cloudStorage?: { id: string };
    volume?: { id: string };
    storagePrefix?: string;
    password?: string;
    compression?: string;
    packSize?: string;
    retention?: BackupRepo_Retention_Payload;
    inheritable?: boolean;
    default?: boolean;
};

export type BackupRepo_CreateOne_Req = ApiRequestBase<{
    payload: BackupRepo_CreateOne_Payload;
}>;

export type BackupRepo_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type BackupRepo_UpdateOne_Payload = {
    updateVer: number;
    name: string;
    description?: string;
    compression?: string;
    packSize?: string;
    retention?: BackupRepo_Retention_Payload;
    inheritable?: boolean;
    default?: boolean;
};

export type BackupRepo_UpdateOne_Req = ApiRequestBase<{
    id: string;
    payload: BackupRepo_UpdateOne_Payload;
}>;

export type BackupRepo_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

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

export type BackupRepo_UpdatePassword_Payload = {
    updateVer: number;
    currentPassword: string;
    newPassword: string;
    inheritable?: boolean;
    default?: boolean;
};

export type BackupRepo_UpdatePassword_Req = ApiRequestBase<{
    id: string;
    payload: BackupRepo_UpdatePassword_Payload;
}>;

export type BackupRepo_UpdatePassword_Res = ApiResponseBase<{ type: "success" }>;

export type BackupRepo_Cleanup_Req = ApiRequestBase<{
    id: string;
}>;

export type BackupRepo_Cleanup_Res = ApiResponseBase<{ type: "success" }>;

export type BackupRepo_Sync_Req = ApiRequestBase<{
    id: string;
}>;

export type BackupRepo_Sync_Res = ApiResponseBase<{ type: "success" }>;
