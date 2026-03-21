import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingNotification } from "~/settings/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

/**
 * Notifications
 */
export type Notifications_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;
export type Notifications_FindManyPaginated_Res = ApiResponsePaginated<SettingNotification>;

export type Notifications_FindOneById_Req = ApiRequestBase<{
    id: string;
}>;
export type Notifications_FindOneById_Res = ApiResponseBase<SettingNotification>;

export type Notifications_CreateOne_Req = ApiRequestBase<{
    payload: Omit<SettingNotification, "id" | "createdAt" | "updatedAt" | "expireAt" | "updateVer">;
}>;
export type Notifications_CreateOne_Res = ApiResponseBase<SettingNotification>;

export type Notifications_UpdateOne_Req = ApiRequestBase<{
    id: string;
    payload: Partial<Omit<SettingNotification, "id" | "createdAt" | "updatedAt" | "expireAt" | "updateVer">>;
}>;
export type Notifications_UpdateOne_Res = ApiResponseBase<SettingNotification>;

export type Notifications_DeleteOne_Req = ApiRequestBase<{
    id: string;
}>;
export type Notifications_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
