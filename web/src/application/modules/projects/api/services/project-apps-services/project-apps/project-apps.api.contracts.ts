import { type PaginationState, type SortingState } from "@infrastructure/data";
import { type ProjectAppBase, type ProjectAppDetails } from "~/projects/domain";
import type { EProjectAppStatus } from "~/projects/module-shared/enums";

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
    env: string;
    appID: string;
    getStats?: boolean;
}>;

export type ProjectApps_FindOneById_Res = ApiResponseBase<ProjectAppDetails>;

/**
 * Delete project app
 */
export type ProjectApps_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    env: string;
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
        env: string;
        appID: string;
        updateVer: number;
    } & Partial<
        Omit<
            ProjectAppDetails,
            | "id"
            | "key"
            | "env"
            | "status"
            | "createdAt"
            | "updatedAt"
            | "userAccesses"
            | "stats"
            | "parentApp"
            | "updateVer"
        >
    > & {
            status?: EProjectAppStatus;
        }
>;

export type ProjectApps_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Update project app status
 */
export type ProjectApps_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    payload: {
        updateVer: number;
        status: EProjectAppStatus;
    };
}>;

export type ProjectApps_UpdateStatus_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Deploy project app
 */
export type ProjectApps_Deploy_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    noCache?: boolean;
}>;

export type ProjectApps_Deploy_Res = ApiResponseBase<{
    deploymentId: string;
}>;

/**
 * Restart project app
 */
export type ProjectApps_Restart_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
}>;

export type ProjectApps_Restart_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Set project app running status
 */
export type ProjectApps_SetRunning_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    running: boolean;
}>;

export type ProjectApps_SetRunning_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Update project app photo
 */
export type ProjectAppPhotoPayload =
    | { fileName: string; dataBase64: string }
    | { fileName: string; isPresetIcon: true }
    | { delete: true };

export type ProjectApps_UpdatePhoto_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    photo: ProjectAppPhotoPayload;
}>;

export type ProjectApps_UpdatePhoto_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Detect project app photo
 */
export type ProjectApps_DetectPhoto_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
}>;

export type ProjectApps_DetectPhoto_Res = ApiResponseBase<{
    url: string;
}>;
