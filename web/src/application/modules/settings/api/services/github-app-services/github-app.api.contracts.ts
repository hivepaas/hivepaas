import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingGithubApp } from "~/settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type GithubApp_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;
export type GithubApp_FindManyPaginated_Res = ApiResponsePaginated<SettingGithubApp>;

export type GithubApp_FindOneById_Req = ApiRequestBase<{
    id: string;
}>;
export type GithubApp_FindOneById_Res = ApiResponseBase<SettingGithubApp>;

export type GithubApp_BasePayload = {
    name: string;
    clientId: string;
    clientSecret: string;
    organization: string;
    appId: number;
    installationId: number;
    privateKey: string;
    ssoEnabled: boolean;
};

export type GithubApp_CreateOne_Payload = GithubApp_BasePayload & {
    availableInProjects: boolean;
    default: boolean;
};
export type GithubApp_CreateOne_Req = ApiRequestBase<{
    payload: GithubApp_CreateOne_Payload;
}>;
export type GithubApp_CreateOne_Res = ApiResponseBase<{
    id: string;
    callbackURL?: string;
}>;

export type GithubApp_UpdateOne_Payload = GithubApp_CreateOne_Payload & {
    updateVer: number;
};
export type GithubApp_UpdateOne_Req = ApiRequestBase<{
    id: string;
    payload: GithubApp_UpdateOne_Payload;
}>;
export type GithubApp_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type GithubApp_UpdateStatus_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    availableInProjects?: boolean;
    default?: boolean;
};
export type GithubApp_UpdateStatus_Req = ApiRequestBase<{
    id: string;
    payload: GithubApp_UpdateStatus_Payload;
}>;
export type GithubApp_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type GithubApp_DeleteOne_Req = ApiRequestBase<{
    id: string;
}>;
export type GithubApp_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;

export type GithubApp_TestConnection_Req = ApiRequestBase<{
    payload: GithubApp_BasePayload;
}>;
export type GithubApp_TestConnection_Res = ApiResponseBase<{ type: "success" }>;

export type GithubApp_BeginManifestFlow_Payload = {
    name: string;
    org: string;
    ssoEnabled: boolean;
    availableInProjects: boolean;
    default: boolean;
};
export type GithubApp_BeginManifestFlow_Req = ApiRequestBase<{
    payload: GithubApp_BeginManifestFlow_Payload;
}>;
export type GithubApp_BeginManifestFlow_Res = ApiResponseBase<{
    redirectURL: string;
    settingId: string;
}>;

export type GithubApp_BeginReprovision_Payload = {
    name: string;
    updateVer: number;
};
export type GithubApp_BeginReprovision_Req = ApiRequestBase<{
    id: string;
    payload: GithubApp_BeginReprovision_Payload;
}>;
export type GithubApp_BeginReprovision_Res = ApiResponseBase<{
    redirectURL: string;
}>;

export type GithubApp_ListInstallations_Payload = GithubApp_BasePayload;
export type GithubApp_ListInstallations_Req = ApiRequestBase<{
    payload: GithubApp_ListInstallations_Payload;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;
export type GithubAppInstallation = {
    id: number;
    nodeId: string;
    appId: number;
    appSlug: string;
};
export type GithubApp_ListInstallations_Res = ApiResponsePaginated<GithubAppInstallation>;
