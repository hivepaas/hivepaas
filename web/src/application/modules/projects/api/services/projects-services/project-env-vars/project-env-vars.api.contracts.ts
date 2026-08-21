import { type ProjectBuildtimeEnvVar, type ProjectEnvVar, type ProjectRuntimeEnvVar } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

/**
 * Find one project env vars
 */
export type ProjectEnvVars_FindOne_Req = ApiRequestBase<{
    projectID: string;
    /** When set to a concrete env name, uses /projects/{id}/{env}/env-vars */
    env?: string;
}>;

export type ProjectEnvVars_FindOne_Res = ApiResponseBase<ProjectEnvVar>;

/**
 * Update project env vars
 */
export type ProjectEnvVars_UpdateOne_Req = ApiRequestBase<
    {
        projectID: string;
        /** When set to a concrete env name, uses /projects/{id}/{env}/env-vars */
        env?: string;
        updateVer: number;
        buildtime: ProjectBuildtimeEnvVar[];
        runtime: ProjectRuntimeEnvVar[];
    }
>;

export type ProjectEnvVars_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Compute project env vars
 */
export type ProjectEnvVars_Compute_Req = ApiRequestBase<{
    projectID: string;
    /** When set to a concrete env name, uses /projects/{id}/{env}/env-vars/compute */
    env?: string;
    buildtimeEnvVars?: { key: string; value: string; isLiteral: boolean }[];
    runtimeEnvVars?: { key: string; value: string; isLiteral: boolean }[];
}>;

export type ProjectEnvVars_Compute_Res = ApiResponseBase<{ key: string; value: string }[]>;
