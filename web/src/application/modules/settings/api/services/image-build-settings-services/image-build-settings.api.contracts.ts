import type {
    ImageBuildRepoCacheClearResult,
    ImageBuildRepoCacheInfo,
    ImageBuildSettings,
    ImageBuildWorkerNode,
} from "~/settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type ImageBuildSettings_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type ImageBuildSettings_FindOne_Res = ApiResponseBase<ImageBuildSettings>;

export type ImageBuildSettings_NamedObject_Payload = Pick<ImageBuildWorkerNode, "id">;

export type ImageBuildSettings_UpdateOne_Payload = {
    updateVer: number;
    availableInProjects?: boolean;
    default?: boolean;
    workers: {
        nodes: ImageBuildSettings_NamedObject_Payload[];
        nodeLabels: string[];
        maxParallelism: number;
    };
    resources: ImageBuildSettings["resources"];
    sources: ImageBuildSettings["sources"];
    noCache: boolean;
    noVerbose: boolean;
};

export type ImageBuildSettings_UpdateOne_Req = ApiRequestBase<{
    payload: ImageBuildSettings_UpdateOne_Payload;
}>;
export type ImageBuildSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ImageBuildSettings_FindRepoCache_Req = ApiRequestBase<Record<string, never>>;
export type ImageBuildSettings_FindRepoCache_Res = ApiResponseBase<ImageBuildRepoCacheInfo>;

export type ImageBuildSettings_ClearRepoCache_Req = ApiRequestBase<Record<string, never>>;
export type ImageBuildSettings_ClearRepoCache_Res = ApiResponseBase<ImageBuildRepoCacheClearResult>;
