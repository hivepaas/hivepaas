import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type HivePaaSAppSecret_UpdateOne_Payload = {
    currentSecret?: string;
    newSecret: string;
};

export type HivePaaSAppSecret_UpdateOne_Req = ApiRequestBase<{
    payload: HivePaaSAppSecret_UpdateOne_Payload;
}>;

export type HivePaaSAppSecret_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
