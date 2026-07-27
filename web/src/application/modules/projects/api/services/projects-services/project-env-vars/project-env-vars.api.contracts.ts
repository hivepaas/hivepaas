import { type ProjectEnvVar } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

/**
 * Find one project env vars
 */
export type ProjectEnvVars_FindOne_Req = ApiRequestBase<{
    projectID: string;
}>;

export type ProjectEnvVars_FindOne_Res = ApiResponseBase<ProjectEnvVar>;

/**
 * Update project env vars
 */
export type ProjectEnvVars_UpdateOne_Req = ApiRequestBase<
    {
        projectID: string;
        updateVer: number;
    } & ProjectEnvVar
>;

export type ProjectEnvVars_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Compute project env vars
 */
export type ProjectEnvVars_Compute_Req = ApiRequestBase<{
    projectID: string;
    buildtimeEnvVars?: { key: string; value: string; isLiteral: boolean }[];
    runtimeEnvVars?: { key: string; value: string; isLiteral: boolean }[];
}>;

export type ProjectEnvVars_Compute_Res = ApiResponseBase<{ key: string; value: string }[]>;
