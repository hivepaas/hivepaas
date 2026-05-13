import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingImService } from "~/settings/domain";

import type { EImServiceKind, ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ImService_Webhook_Payload = {
    webhook: string;
};

export type ImService_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ImService_FindManyPaginated_Res = ApiResponsePaginated<SettingImService>;

export type ImService_FindOneById_Req = ApiRequestBase<{
    id: string;
}>;

export type ImService_FindOneById_Res = ApiResponseBase<SettingImService>;

export type ImService_CreateOne_Payload = {
    availableInProjects: boolean;
    default: boolean;
    name: string;
    kind: EImServiceKind;
    slack: ImService_Webhook_Payload | null;
    discord: ImService_Webhook_Payload | null;
};

export type ImService_CreateOne_Req = ApiRequestBase<{
    payload: ImService_CreateOne_Payload;
}>;

export type ImService_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type ImService_UpdateOne_Payload = ImService_CreateOne_Payload & {
    updateVer: number;
};

export type ImService_UpdateOne_Req = ApiRequestBase<{
    id: string;
    payload: ImService_UpdateOne_Payload;
}>;

export type ImService_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ImService_UpdateStatus_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    availableInProjects?: boolean;
    default?: boolean;
};

export type ImService_UpdateStatus_Req = ApiRequestBase<{
    id: string;
    payload: ImService_UpdateStatus_Payload;
}>;

export type ImService_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type ImService_DeleteOne_Req = ApiRequestBase<{
    id: string;
}>;

export type ImService_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;

export type ImService_TestSendMsg_Payload = ImService_CreateOne_Payload & {
    testMsg: string;
};

export type ImService_TestSendMsg_Req = ApiRequestBase<{
    payload: ImService_TestSendMsg_Payload;
}>;

export type ImService_TestSendMsg_Res = ApiResponseBase<{ type: "success" }>;
