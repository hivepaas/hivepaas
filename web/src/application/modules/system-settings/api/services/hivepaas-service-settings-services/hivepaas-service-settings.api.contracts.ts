import type { HivePaaSServiceSettings } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type HivePaaSServiceSettings_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type HivePaaSServiceSettings_FindOne_Res = ApiResponseBase<HivePaaSServiceSettings>;

export type HivePaaSServiceSettings_UpdateOne_Payload = {
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

export type HivePaaSServiceSettings_UpdateOne_Req = ApiRequestBase<{
    payload: HivePaaSServiceSettings_UpdateOne_Payload;
}>;
export type HivePaaSServiceSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
