import { type ProjectEnvVar } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

/**
 * Find one project app env vars
 */
export type ProjectAppEnvVars_FindOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
}>;

export type ProjectAppEnvVars_FindOne_Res = ApiResponseBase<ProjectEnvVar>;

/**
 * Update project app env vars
 */
export type ProjectAppEnvVars_UpdateOne_Req = ApiRequestBase<
    {
        projectID: string;
        appID: string;
        updateVer: number;
    } & ProjectEnvVar
>;

export type ProjectAppEnvVars_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;
