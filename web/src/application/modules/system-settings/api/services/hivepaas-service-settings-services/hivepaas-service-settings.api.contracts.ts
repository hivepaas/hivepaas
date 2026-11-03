import type { HivePaaSServiceSettings, SettingsPendingChange } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type HivePaaSServiceSettings_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type HivePaaSServiceSettings_FindOne_Res = ApiResponseBase<HivePaaSServiceSettings>;

export type HivePaaSServiceSettings_UpdateOne_Payload = {
    updateVer: number;

    /**
     * How long a proxy settings change may stay unconfirmed before it is undone,
     * as a Go duration string ("5m"). Ignored when the request leaves the proxy
     * settings alone, since nothing else here can lock the caller out.
     */
    confirmWindow?: string;
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
    periodicSettings: {
        baseInterval: string;
        batchSize: number;
    };
    proxySettings: {
        proxyProvider: string;
        trustedIPs: string[];
        proxyHops: number;
    };
};

export type HivePaaSServiceSettings_UpdateOne_Req = ApiRequestBase<{
    payload: HivePaaSServiceSettings_UpdateOne_Payload;
}>;
// Null unless the request changed the proxy settings - those are the only ones
// put on trial.
export type HivePaaSServiceSettings_UpdateOne_Res = ApiResponseBase<{
    pendingChange: SettingsPendingChange | null;
}>;

export type HivePaaSServiceSettings_ConfirmChange_Req = ApiRequestBase<{ changeId: string }>;
export type HivePaaSServiceSettings_ConfirmChange_Res = ApiResponseBase<{ type: "success" }>;

export type HivePaaSServiceSettings_RevertChange_Req = ApiRequestBase<{ changeId: string }>;
export type HivePaaSServiceSettings_RevertChange_Res = ApiResponseBase<{
    reverted: boolean;
    reason: string | null;
}>;
