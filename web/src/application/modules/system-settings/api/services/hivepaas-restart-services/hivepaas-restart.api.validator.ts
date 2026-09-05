import type { AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { HivePaaSRestart_Execute_Res } from "./hivepaas-restart.api.contracts";

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class HivePaaSRestartApiValidator {
    execute = (response: AxiosResponse): HivePaaSRestart_Execute_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
