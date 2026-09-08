import type { AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { HivePaaSRequestInfo_FindOne_Res } from "./hivepaas-request-info.api.contracts";

const RequestInfoSchema = z.object({
    remoteAddr: z.string(),
    clientIp: z.string(),
    forwardedFor: z.array(z.string()).nullish(),
    headers: z.record(z.string()).nullish(),
    suggestedProxyHops: z.number(),
    explanation: z.string(),
});

const FindOneSchema = z.object({
    data: RequestInfoSchema,
    meta: BaseMetaApiSchema.nullish(),
});

export class HivePaaSRequestInfoApiValidator {
    findOne = (response: AxiosResponse): HivePaaSRequestInfo_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return {
            data: {
                remoteAddr: data.remoteAddr,
                clientIp: data.clientIp,
                forwardedFor: data.forwardedFor ?? [],
                headers: data.headers ?? {},
                suggestedProxyHops: data.suggestedProxyHops,
                explanation: data.explanation,
            },
            meta,
        };
    };
}
