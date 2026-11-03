import type { PaginationState, SortingState } from "@infrastructure/data";
import type { ProjectCommandPipe } from "~/projects/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectCommandPipe_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectCommandPipe_FindManyPaginated_Res = ApiResponsePaginated<ProjectCommandPipe>;

export type ProjectCommandPipe_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    id: string;
}>;

export type ProjectCommandPipe_FindOneById_Res = ApiResponseBase<ProjectCommandPipe>;

export type ProjectCommandPipe_CreateOne_Payload = {
    inheritable: boolean;
    default: boolean;
    name: string;
    sourceCommand: { id: string };
    targetCommand: { id: string };
};

export type ProjectCommandPipe_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    payload: ProjectCommandPipe_CreateOne_Payload;
}>;

export type ProjectCommandPipe_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type ProjectCommandPipe_CreateFromTemplate_Payload = {
    commandType: string;
    commandKind: string;
};

export type ProjectCommandPipe_CreateFromTemplate_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    payload: ProjectCommandPipe_CreateFromTemplate_Payload;
}>;

export type ProjectCommandPipe_CreateFromTemplate_Res = ProjectCommandPipe_CreateOne_Res;

export type ProjectCommandPipe_UpdateOne_Payload = ProjectCommandPipe_CreateOne_Payload & {
    updateVer: number;
};

export type ProjectCommandPipe_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    id: string;
    payload: ProjectCommandPipe_UpdateOne_Payload;
}>;

export type ProjectCommandPipe_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectCommandPipe_UpdateStatus_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    inheritable?: boolean;
    default?: boolean;
};

export type ProjectCommandPipe_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    id: string;
    payload: ProjectCommandPipe_UpdateStatus_Payload;
}>;

export type ProjectCommandPipe_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectCommandPipe_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    id: string;
}>;

export type ProjectCommandPipe_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
