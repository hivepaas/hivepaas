import type { AttentionItem } from "~/home/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type HomeAttention_FindAll_Req = ApiRequestBase<Record<string, never>>;
export type HomeAttention_FindAll_Res = ApiResponseBase<AttentionItem[]>;
