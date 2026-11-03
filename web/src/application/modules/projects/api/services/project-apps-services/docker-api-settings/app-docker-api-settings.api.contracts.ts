import type { AppDockerApiMode, AppDockerApiSettings } from "~/projects/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type AppDockerApiSettings_FindOne_Req = ApiRequestBase<{ projectID: string; env: string; appID: string }>;
export type AppDockerApiSettings_FindOne_Res = ApiResponseBase<AppDockerApiSettings>;

/** The whole new state. With enabled off the server keeps what access allowed and reads nothing else. */
export type AppDockerApiSettings_UpdatePayload = {
    enabled: boolean;
    mode: AppDockerApiMode;
    images: string[];
    sharedDirs: string[];
    sharedVolumes: Record<string, string>;
    networks: string[];
    allow: string[];
    /** Left out where the default is kept. */
    limits: { containers?: number; memory?: string; cpus?: number };
    updateVer: number;
};

export type AppDockerApiSettings_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    payload: AppDockerApiSettings_UpdatePayload;
}>;
export type AppDockerApiSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
