import type { AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { TraefikRestart_Execute_Res } from "./traefik-restart.api.contracts";

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class TraefikRestartApiValidator {
    execute = (response: AxiosResponse): TraefikRestart_Execute_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
