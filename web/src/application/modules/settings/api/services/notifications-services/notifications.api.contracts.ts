import type { PaginationState, SortingState } from "@infrastructure/data";
import type { NotificationEntity } from "~/settings/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

/**
 * Notifications
 */
export type Notifications_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;
export type Notifications_FindManyPaginated_Res = ApiResponsePaginated<NotificationEntity>;

export type Notifications_FindOneById_Req = ApiRequestBase<{
    id: string;
}>;
export type Notifications_FindOneById_Res = ApiResponseBase<NotificationEntity>;

export type Notifications_CreateOne_Req = ApiRequestBase<{
    payload: Omit<NotificationEntity, "id" | "createdAt" | "updatedAt" | "expireAt" | "updateVer">;
}>;
export type Notifications_CreateOne_Res = ApiResponseBase<NotificationEntity>;

export type Notifications_UpdateOne_Req = ApiRequestBase<{
    id: string;
    payload: Partial<Omit<NotificationEntity, "id" | "createdAt" | "updatedAt" | "expireAt" | "updateVer">>;
}>;
export type Notifications_UpdateOne_Res = ApiResponseBase<NotificationEntity>;

export type Notifications_DeleteOne_Req = ApiRequestBase<{
    id: string;
}>;
export type Notifications_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
