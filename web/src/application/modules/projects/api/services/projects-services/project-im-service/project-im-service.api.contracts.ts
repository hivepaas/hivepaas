import type { PaginationState, SortingState } from "@infrastructure/data";
import type {
    ImService_CreateOne_Payload,
    ImService_UpdateOne_Payload,
    ImService_UpdateStatus_Payload,
} from "~/settings/api/services/im-service-services";
import type { SettingImService } from "~/settings/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectImService_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectImService_FindManyPaginated_Res = ApiResponsePaginated<SettingImService>;

export type ProjectImService_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;

export type ProjectImService_FindOneById_Res = ApiResponseBase<SettingImService>;

export type ProjectImService_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: ImService_CreateOne_Payload;
}>;

export type ProjectImService_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type ProjectImService_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: ImService_UpdateOne_Payload;
}>;

export type ProjectImService_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectImService_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: ImService_UpdateStatus_Payload;
}>;

export type ProjectImService_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectImService_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;

export type ProjectImService_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
