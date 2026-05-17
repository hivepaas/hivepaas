import type { PaginationState, SortingState } from "@infrastructure/data";
import type { ProjectNotification } from "~/projects/domain";
import type {
    Notification_CreateOne_Payload,
    Notifications_UpdateOne_Payload,
    Notifications_UpdateStatus_Payload,
} from "~/settings/api/services/notifications-services";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectNotification_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectNotification_FindManyPaginated_Res = ApiResponsePaginated<ProjectNotification>;

export type ProjectNotification_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;

export type ProjectNotification_FindOneById_Res = ApiResponseBase<ProjectNotification>;

export type ProjectNotification_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: Notification_CreateOne_Payload;
}>;
export type ProjectNotification_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type ProjectNotification_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: Notifications_UpdateOne_Payload;
}>;
export type ProjectNotification_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectNotification_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: Notifications_UpdateStatus_Payload;
}>;
export type ProjectNotification_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectNotification_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;
export type ProjectNotification_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
