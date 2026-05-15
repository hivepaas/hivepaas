import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingEmail } from "~/settings/domain";

import type { EEmailKind, ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type Email_HTTP_Method = "GET" | "POST" | "PUT";

export type Email_SMTP_Payload = {
    host: string;
    port: number;
    username: string;
    displayName: string;
    password: string;
    ssl: boolean;
};

export type Email_HTTP_FieldMapping_Payload = {
    fromAddress: string;
    fromName: string;
    toAddress: string;
    toAddresses: string;
    subject: string;
    content: string;
    password: string;
};

export type Email_HTTP_Payload = {
    endpoint: string;
    method: Email_HTTP_Method;
    contentType: string;
    headers: Record<string, string>;
    fieldMapping: Email_HTTP_FieldMapping_Payload;
    username: string;
    displayName: string;
    password: string;
};

export type Email_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type Email_FindManyPaginated_Res = ApiResponsePaginated<SettingEmail>;

export type Email_FindOneById_Req = ApiRequestBase<{
    id: string;
}>;

export type Email_FindOneById_Res = ApiResponseBase<SettingEmail>;

export type Email_CreateOne_Payload = {
    availableInProjects: boolean;
    default: boolean;
    name: string;
    kind: EEmailKind;
    smtp: Email_SMTP_Payload | null;
    http: Email_HTTP_Payload | null;
};

export type Email_CreateOne_Req = ApiRequestBase<{
    payload: Email_CreateOne_Payload;
}>;

export type Email_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type Email_UpdateOne_Payload = Email_CreateOne_Payload & {
    updateVer: number;
};

export type Email_UpdateOne_Req = ApiRequestBase<{
    id: string;
    payload: Email_UpdateOne_Payload;
}>;

export type Email_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type Email_UpdateStatus_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    availableInProjects?: boolean;
    default?: boolean;
};

export type Email_UpdateStatus_Req = ApiRequestBase<{
    id: string;
    payload: Email_UpdateStatus_Payload;
}>;

export type Email_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type Email_DeleteOne_Req = ApiRequestBase<{
    id: string;
}>;

export type Email_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;

export type Email_TestSendMail_Payload = Omit<Email_CreateOne_Payload, "availableInProjects" | "default"> & {
    testRecipient: string;
    testSubject: string;
    testContent: string;
};

export type Email_TestSendMail_Req = ApiRequestBase<{
    payload: Email_TestSendMail_Payload;
}>;

export type Email_TestSendMail_Res = ApiResponseBase<{ type: "success" }>;
