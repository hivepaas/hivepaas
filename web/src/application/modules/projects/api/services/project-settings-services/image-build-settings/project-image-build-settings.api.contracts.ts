import type { ProjectImageBuildRepoCacheInfo, ProjectImageBuildSettings } from "~/projects/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type ProjectImageBuildSettings_FindOne_Req = ApiRequestBase<{
    projectID: string;
}>;
export type ProjectImageBuildSettings_FindOne_Res = ApiResponseBase<ProjectImageBuildSettings>;

export type ProjectImageBuildSettings_UpdateOne_Payload = Pick<
    ProjectImageBuildSettings,
    "resources" | "sources" | "noCache" | "noVerbose" | "updateVer"
>;

export type ProjectImageBuildSettings_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: ProjectImageBuildSettings_UpdateOne_Payload;
}>;
export type ProjectImageBuildSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectImageBuildSettings_FindRepoCache_Req = ApiRequestBase<{
    projectID: string;
}>;
export type ProjectImageBuildSettings_FindRepoCache_Res = ApiResponseBase<ProjectImageBuildRepoCacheInfo>;

export type ProjectImageBuildSettings_ClearRepoCache_Req = ApiRequestBase<{
    projectID: string;
}>;
export type ProjectImageBuildSettings_ClearRepoCache_Res = ApiResponseBase<{ type: "success" }>;
