import { z } from "zod";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { HivePaaSAppSecret_UpdateOne_Res } from "./hivepaas-app-secret.api.contracts";

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class HivePaaSAppSecretApiValidator {
    updateOne = (response: ApiHttpResponse): HivePaaSAppSecret_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
