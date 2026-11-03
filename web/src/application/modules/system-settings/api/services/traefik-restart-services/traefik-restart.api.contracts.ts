import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type TraefikRestart_Execute_Req = ApiRequestBase<Record<string, never>>;

export type TraefikRestart_Execute_Res = ApiResponseBase<{ type: "success" }>;
