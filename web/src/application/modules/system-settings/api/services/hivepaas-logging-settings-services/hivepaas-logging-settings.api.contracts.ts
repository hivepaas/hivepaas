import type { HivePaaSLoggingEndpoint, HivePaaSLoggingSettings, HivePaaSLoggingStatus } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type HivePaaSLoggingSettings_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type HivePaaSLoggingSettings_FindOne_Res = ApiResponseBase<{
    settings: HivePaaSLoggingSettings;
    loggingStatus: HivePaaSLoggingStatus;
}>;

export type HivePaaSLoggingSettings_UpdateOnePayload = {
    updateVer: number;
    enabled: boolean;
    sources: { apps: boolean; hivepaas: boolean; traefikAccess: boolean; nodes: boolean };
    collector: { type: string; managed: boolean };
    backend: {
        type: string;
        managed: boolean;
        ingest?: HivePaaSLoggingEndpoint | null;
        query?: HivePaaSLoggingEndpoint | null;
        victoriaLogs?: {
            volume: { id: string };
            retention: string;
            maxDiskUsagePercent?: number;
            cpuLimit?: number;
            memoryLimit?: string;
        } | null;
    };
    forwards: { name: string; format?: string; endpoint: HivePaaSLoggingEndpoint }[];
    /**
     * Asked of this save, not stored by it: switching logging off, or handing the
     * backend to a store somebody else runs, takes an app down, and the apps have
     * no screen of their own to be removed from.
     */
    removeApp?: boolean;
    /** Deletes the stored logs with the backend. */
    removeStorage?: boolean;
};

export type HivePaaSLoggingSettings_UpdateOne_Req = ApiRequestBase<{
    payload: HivePaaSLoggingSettings_UpdateOnePayload;
}>;
export type HivePaaSLoggingSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
