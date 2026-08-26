import { type PaginationState, type SortingState } from "@infrastructure/data";
import { type ProjectBaseEntity, type ProjectDetailsEntity } from "~/projects/domain";
import type { EProjectEnvStatus, EProjectStatus } from "~/projects/module-shared/enums";

import {
    type ApiRequestBase,
    type ApiResponseBase,
    type ApiResponsePaginated,
    type OpenApiConstant,
} from "@infrastructure/api";

/**
 * Find many projects paginated
 */
export type Projects_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type Projects_FindManyPaginated_Res = ApiResponsePaginated<ProjectBaseEntity>;

/**
 * Create project
 */
export type Projects_CreateOne_Req = ApiRequestBase<Pick<ProjectBaseEntity, "name" | "note" | "envs" | "tags">>;

export type Projects_CreateOne_Res = ApiResponseBase<{
    id: string;
}>;

/**
 * Find one project by id
 */
export type Projects_FindOneById_Req = ApiRequestBase<{
    projectID: string;
}>;

export type Projects_FindOneById_Res = ApiResponseBase<ProjectDetailsEntity>;

/**
 * Delete project
 */
export type Projects_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
}>;

export type Projects_DeleteOne_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Update project
 */
export type Projects_UpdateOne_Req = ApiRequestBase<
    {
        projectID: string;
        updateVer: number;
        owner?: { id: string };
    } & Partial<
        Omit<
            ProjectDetailsEntity,
            "id" | "key" | "status" | "createdAt" | "updatedAt" | "updateVer" | "owner" | "userAccesses"
        >
    > & {
            status?: OpenApiConstant<EProjectStatus> | EProjectStatus;
        }
>;

export type Projects_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Update project status
 */
export type Projects_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    payload: {
        updateVer: number;
        status: OpenApiConstant<EProjectStatus> | EProjectStatus;
    };
}>;

export type Projects_UpdateStatus_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Update project photo
 */
export type ProjectPhotoPayload = { fileName: string; dataBase64: string } | { delete: true };

export type Projects_UpdatePhoto_Req = ApiRequestBase<{
    projectID: string;
    photo: ProjectPhotoPayload;
}>;

export type Projects_UpdatePhoto_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Update project env status
 */
export type Projects_UpdateEnvStatus_Req = ApiRequestBase<{
    projectID: string;
    envName: string;
    payload: {
        updateVer?: number;
        status: EProjectEnvStatus;
    };
}>;

export type Projects_UpdateEnvStatus_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Delete project env
 */
export type Projects_DeleteEnv_Req = ApiRequestBase<{
    projectID: string;
    envName: string;
}>;

export type Projects_DeleteEnv_Res = ApiResponseBase<{
    type: "success";
}>;
