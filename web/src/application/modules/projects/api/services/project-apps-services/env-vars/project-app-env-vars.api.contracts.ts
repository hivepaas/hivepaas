import {
    type ProjectAppEnvVar,
    type ProjectBuildtimeEnvVar,
    type ProjectRuntimeEnvVar,
    type ProjectSharedEnvVar,
} from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

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
