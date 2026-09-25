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
    env: string;
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
    env: string;
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
    env: string;
    appID: string;
    buildtimeEnvVars?: EnvVarWireItem[];
    runtimeEnvVars?: EnvVarWireItem[];
    sharedEnvVars?: EnvVarWireItem[];
}>;

export type ProjectAppEnvVars_Compute_Res = ApiResponseBase<{ key: string; value: string }[]>;

export type EnvLinkTarget = {
    id: string;
    key: string;
    name: string;
    category: string;
    engine: string;
};

export type EnvLinkVar = {
    key: string;
    value: string;
    description: string;
};

export type EnvLinkGroup = {
    id: string;
    title: string;
    description: string;
    recommended: boolean;
    warnings: string[];
    vars: EnvLinkVar[];
};

export type ProjectAppEnvVars_FindLinkTargets_Req = ApiRequestBase<{ projectID: string; env: string; appID: string }>;
export type ProjectAppEnvVars_FindLinkTargets_Res = ApiResponseBase<EnvLinkTarget[]>;

export type ProjectAppEnvVars_FindLinkSuggestions_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    targetAppID: string;
}>;
export type ProjectAppEnvVars_FindLinkSuggestions_Res = ApiResponseBase<{
    target: EnvLinkTarget;
    groups: EnvLinkGroup[];
}>;
