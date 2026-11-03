import { z } from "zod";
import { TraefikServiceSettingsEntitySchema } from "~/system-settings/module-shared/schemas";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    TraefikServiceSettings_FindOne_Res,
    TraefikServiceSettings_UpdateOne_Res,
} from "./traefik-service-settings.api.contracts";

const FindOneSchema = z.object({
    data: TraefikServiceSettingsEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class TraefikServiceSettingsApiValidator {
    findOne = (response: ApiHttpResponse): TraefikServiceSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data, meta };
    };

    updateOne = (response: ApiHttpResponse): TraefikServiceSettings_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
