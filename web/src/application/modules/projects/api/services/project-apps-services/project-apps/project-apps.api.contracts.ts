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

export interface ProjectApps_CopyToggle {
    copy: boolean;
}

export interface ProjectApps_CopySslCertRef {
    id: string;
    name: string;
}

export interface ProjectApps_CopyPreparedDomainSetting {
    sourceDomain: string;
    targetDomain: string;
    sourceSslCert: ProjectApps_CopySslCertRef | null;
    targetSslCert: ProjectApps_CopySslCertRef | null;
}

export interface ProjectApps_CopyDomainSettingPayload {
    sourceDomain: string;
    targetDomain: string;
    sourceSslCert: { id: string };
    targetSslCert: { id: string };
}

export interface ProjectApps_CopyHttpSettings<TDomainSetting> {
    copy: boolean;
    copyDomainSettings: TDomainSetting[];
}

export interface ProjectApps_CopyBase<TDomainSetting> {
    sourceName: string;
    targetName: string;
    sourceEnv: string;
    targetEnv: string;
    sourceStatus: EProjectAppStatus;
    targetStatus: EProjectAppStatus;
    copyConfigFiles: ProjectApps_CopyToggle;
    copyDeploymentSettings: ProjectApps_CopyToggle;
    copyEnvVars: ProjectApps_CopyToggle;
    copyHealthChecks: ProjectApps_CopyToggle;
    copyHttpSettings: ProjectApps_CopyHttpSettings<TDomainSetting>;
    copySchedJobs: ProjectApps_CopyToggle;
    copySecrets: ProjectApps_CopyToggle;
    updateVer: number;
}

export type ProjectApps_PrepareCopy_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
}>;

export type ProjectApps_PrepareCopy_Res = ApiResponseBase<ProjectApps_CopyBase<ProjectApps_CopyPreparedDomainSetting>>;

export type ProjectApps_Copy_Req = ApiRequestBase<
    {
        projectID: string;
        appID: string;
    } & ProjectApps_CopyBase<ProjectApps_CopyDomainSettingPayload>
>;

export type ProjectApps_Copy_Res = ApiResponseBase<{
    id: string;
}>;

/**
 * Find one project app by id
 */
export type ProjectApps_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    getStats?: boolean;
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
            "id" | "key" | "env" | "createdAt" | "updatedAt" | "userAccesses" | "stats" | "parentApp" | "updateVer"
        >
    >
>;

export type ProjectApps_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Update project app status
 */
export type ProjectApps_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
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

/**
 * Set project app running status
 */
export type ProjectApps_SetRunning_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    running: boolean;
}>;

export type ProjectApps_SetRunning_Res = ApiResponseBase<{
    type: "success";
}>;
