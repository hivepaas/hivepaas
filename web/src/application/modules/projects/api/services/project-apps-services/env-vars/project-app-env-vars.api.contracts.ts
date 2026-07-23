import {
    type ProjectAppEnvVar,
    type ProjectBuildtimeEnvVar,
    type ProjectRuntimeEnvVar,
    type ProjectSharedEnvVar,
} from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

export type EnvVarWireItem = {
    key: string;
    value: string;
    isLiteral: boolean;
};

/**
 * Find one project app env vars
 */
export type ProjectAppEnvVars_FindOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
}>;

export type ProjectAppEnvVars_FindOne_Res = ApiResponseBase<
    ProjectAppEnvVar & {
        inheritedBuildtimeEnvVars: ProjectBuildtimeEnvVar[];
        inheritedRuntimeEnvVars: ProjectRuntimeEnvVar[];
    }
>;

/**
 * Update project app env vars
 */
export type ProjectAppEnvVars_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    updateVer: number;
    buildtime: ProjectBuildtimeEnvVar[];
    runtime: ProjectRuntimeEnvVar[];
    shared: ProjectSharedEnvVar[];
}>;

export type ProjectAppEnvVars_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Compute project app env vars
 */
export type ProjectAppEnvVars_Compute_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    buildtimeEnvVars?: EnvVarWireItem[];
    runtimeEnvVars?: EnvVarWireItem[];
    sharedEnvVars?: EnvVarWireItem[];
}>;

export type ProjectAppEnvVars_Compute_Res = ApiResponseBase<{ key: string; value: string }[]>;
