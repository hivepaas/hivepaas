import type { LocalPaaSServiceSettings } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type LocalPaaSServiceSettings_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type LocalPaaSServiceSettings_FindOne_Res = ApiResponseBase<LocalPaaSServiceSettings>;

export type LocalPaaSServiceSettings_UpdateOne_Payload = {
    updateVer: number;
    appSettings: {
        replicas: number;
    };
    workerSettings: {
        replicas: number;
        concurrency: number;
        runWorkerInMainApp: boolean;
    };
    taskSettings: {
        taskCheckInterval: string;
        taskCreateInterval: string;
    };
    healthcheckSettings: {
        baseInterval: string;
    };
};

export type LocalPaaSServiceSettings_UpdateOne_Req = ApiRequestBase<{
    payload: LocalPaaSServiceSettings_UpdateOne_Payload;
}>;
export type LocalPaaSServiceSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
