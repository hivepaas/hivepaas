import type { AxiosResponse } from "axios";
import { z } from "zod";
import { LocalPaaSServiceSettingsEntitySchema } from "~/system-settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    LocalPaaSServiceSettings_FindOne_Res,
    LocalPaaSServiceSettings_UpdateOne_Res,
} from "./localpaas-service-settings.api.contracts";

const FindOneSchema = z.object({
    data: LocalPaaSServiceSettingsEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class LocalPaaSServiceSettingsApiValidator {
    findOne = (response: AxiosResponse): LocalPaaSServiceSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data, meta };
    };

    updateOne = (response: AxiosResponse): LocalPaaSServiceSettings_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
