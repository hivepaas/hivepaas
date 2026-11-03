import { z } from "zod";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { HivePaaSProject_FindOne_Res } from "./hivepaas-project.api.contracts";

const FindOneSchema = z.object({
    data: z.object({
        id: z.string(),
        name: z.string(),
        key: z.string(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

export class HivePaaSProjectApiValidator {
    findOne = (response: ApiHttpResponse): HivePaaSProject_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data, meta };
    };
}
