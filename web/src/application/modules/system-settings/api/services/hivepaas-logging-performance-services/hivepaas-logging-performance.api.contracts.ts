import type { LoggingPerformance, LoggingPerformanceCapacityChoice } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type HivePaaSLoggingPerformance_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type HivePaaSLoggingPerformance_FindOne_Res = ApiResponseBase<LoggingPerformance>;

export type HivePaaSLoggingPerformance_UpdateOnePayload = {
    /** The logging settings' version, which these are saved with. */
    updateVer: number;
    enabled: boolean;
    /** The nodes that run OBI, each once. */
    nodes: { id: string; capacity: LoggingPerformanceCapacityChoice }[];
};

export type HivePaaSLoggingPerformance_UpdateOne_Req = ApiRequestBase<{
    payload: HivePaaSLoggingPerformance_UpdateOnePayload;
}>;
export type HivePaaSLoggingPerformance_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
