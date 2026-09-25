import { type PaginationState, type SortingState } from "@infrastructure/data";
import type { AppSettingMount, AppSettingMountSources } from "~/projects/domain";
import type { EProjectSecretStatus } from "~/projects/module-shared/enums";

import { type ApiRequestBase, type ApiResponseBase, type ApiResponsePaginated } from "@infrastructure/api";

type AppScope = {
    projectID: string;
    env: string;
    appID: string;
};

export type AppSettingMountFilePayload = {
    part: string;
    path: string;
    uid: string;
    gid: string;
    mode: string;
};

export type AppSettingMountPayload = {
    name: string;
    inheritable: boolean;
    sourceID: string;
    files: AppSettingMountFilePayload[];
};

export type AppSettingMounts_FindManyPaginated_Req = ApiRequestBase<
    AppScope & { pagination?: PaginationState; sorting?: SortingState; search?: string }
>;
export type AppSettingMounts_FindManyPaginated_Res = ApiResponsePaginated<AppSettingMount>;

export type AppSettingMounts_FindOneById_Req = ApiRequestBase<AppScope & { settingMountID: string }>;
export type AppSettingMounts_FindOneById_Res = ApiResponseBase<AppSettingMount>;

export type AppSettingMounts_FindSources_Req = ApiRequestBase<AppScope>;
export type AppSettingMounts_FindSources_Res = ApiResponseBase<AppSettingMountSources>;

export type AppSettingMounts_CreateOne_Req = ApiRequestBase<AppScope & AppSettingMountPayload>;
export type AppSettingMounts_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type AppSettingMounts_UpdateOne_Req = ApiRequestBase<
    AppScope & AppSettingMountPayload & { settingMountID: string; updateVer: number }
>;
export type AppSettingMounts_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type AppSettingMounts_UpdateStatus_Req = ApiRequestBase<
    AppScope & { settingMountID: string; updateVer: number; status: EProjectSecretStatus }
>;
export type AppSettingMounts_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type AppSettingMounts_DeleteOne_Req = ApiRequestBase<AppScope & { settingMountID: string }>;
export type AppSettingMounts_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
