import type { PaginationState, SortingState } from "@infrastructure/data";
import type {
    Email_CreateOne_Payload,
    Email_UpdateOne_Payload,
    Email_UpdateStatus_Payload,
} from "~/settings/api/services/email-services";
import type { SettingEmail } from "~/settings/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectEmail_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectEmail_FindManyPaginated_Res = ApiResponsePaginated<SettingEmail>;

export type ProjectEmail_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;

export type ProjectEmail_FindOneById_Res = ApiResponseBase<SettingEmail>;

export type ProjectEmail_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: Email_CreateOne_Payload;
}>;

export type ProjectEmail_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type ProjectEmail_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: Email_UpdateOne_Payload;
}>;

export type ProjectEmail_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectEmail_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: Email_UpdateStatus_Payload;
}>;

export type ProjectEmail_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectEmail_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;

export type ProjectEmail_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
