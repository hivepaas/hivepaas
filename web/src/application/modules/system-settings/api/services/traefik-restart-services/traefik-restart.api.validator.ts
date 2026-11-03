import { z } from "zod";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { TraefikRestart_Execute_Res } from "./traefik-restart.api.contracts";

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class TraefikRestartApiValidator {
    execute = (response: ApiHttpResponse): TraefikRestart_Execute_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
