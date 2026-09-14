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
            volumeSubpath?: string;
            retention: string;
            maxDiskUsagePercent?: number;
            cpuLimit?: number;
            memoryLimit?: string;
        } | null;
    };
    forwards: { name: string; format?: string; endpoint: HivePaaSLoggingEndpoint }[];
};

export type HivePaaSLoggingSettings_UpdateOne_Req = ApiRequestBase<{
    payload: HivePaaSLoggingSettings_UpdateOnePayload;
}>;
export type HivePaaSLoggingSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
