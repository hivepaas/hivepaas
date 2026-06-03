import { type PaginationState, type SortingState } from "@infrastructure/data";
import { type ProjectAppBase, type ProjectAppDetails } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase, type ApiResponsePaginated } from "@infrastructure/api";

/**
 * Find many project apps paginated
 */
export type ProjectApps_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
    env?: string;
    getStats?: boolean;
}>;

export type ProjectApps_FindManyPaginated_Res = ApiResponsePaginated<ProjectAppDetails>;

/**
 * Create project app
 */
export type ProjectApps_CreateOne_Req = ApiRequestBase<
    {
        projectID: string;
    } & Pick<ProjectAppBase, "name" | "env" | "note" | "tags">
>;

export type ProjectApps_CreateOne_Res = ApiResponseBase<{
    id: string;
}>;

/**
 * Find one project app by id
 */
export type ProjectApps_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
}>;

export type ProjectApps_FindOneById_Res = ApiResponseBase<ProjectAppDetails>;

/**
 * Delete project app
 */
export type ProjectApps_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
}>;

export type ProjectApps_DeleteOne_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Update project app
 */
export type ProjectApps_UpdateOne_Req = ApiRequestBase<
    {
        projectID: string;
        appID: string;
        updateVer: number;
    } & Partial<
        Omit<
            ProjectAppDetails,
            "id" | "key" | "env" | "createdAt" | "updatedAt" | "userAccesses" | "stats" | "updateVer"
        >
    >
>;

export type ProjectApps_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Deploy project app
 */
export type ProjectApps_Deploy_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
}>;

export type ProjectApps_Deploy_Res = ApiResponseBase<{
    deploymentId: string;
}>;

/**
 * Restart project app
 */
export type ProjectApps_Restart_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
}>;

export type ProjectApps_Restart_Res = ApiResponseBase<{
    type: "success";
}>;
