import type { AxiosResponse } from "axios";
import { z } from "zod";
import { HivePaaSServiceSettingsEntitySchema } from "~/system-settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    HivePaaSServiceSettings_FindOne_Res,
    HivePaaSServiceSettings_UpdateOne_Res,
} from "./hivepaas-service-settings.api.contracts";

const FindOneSchema = z.object({
    data: HivePaaSServiceSettingsEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class HivePaaSServiceSettingsApiValidator {
    findOne = (response: AxiosResponse): HivePaaSServiceSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data, meta };
    };

    updateOne = (response: AxiosResponse): HivePaaSServiceSettings_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
