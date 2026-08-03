import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export interface TraefikConfigOptionsStartupCommand {
    logLevel: string;
    accessLog: boolean;
    http3: boolean;
    fastProxy: boolean;
    args: string[];
}

export interface TraefikConfigOptions {
    startupCommand: TraefikConfigOptionsStartupCommand;
}

export type TraefikConfigOptions_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type TraefikConfigOptions_FindOne_Res = ApiResponseBase<TraefikConfigOptions>;

export type TraefikConfigOptions_UpdateOne_Payload = {
    startupCommand: TraefikConfigOptionsStartupCommand;
};

export type TraefikConfigOptions_UpdateOne_Req = ApiRequestBase<{
    payload: TraefikConfigOptions_UpdateOne_Payload;
}>;
export type TraefikConfigOptions_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
