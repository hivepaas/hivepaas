import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type HivePaaSRestart_Execute_Payload = {
    restartMainApp: boolean;
    restartDbApp: boolean;
    restartCacheApp: boolean;
    restartWorkers: boolean;
    restartAgents: boolean;
};

export type HivePaaSRestart_Execute_Req = ApiRequestBase<{
    payload: HivePaaSRestart_Execute_Payload;
}>;

export type HivePaaSRestart_Execute_Res = ApiResponseBase<{ type: "success" }>;
