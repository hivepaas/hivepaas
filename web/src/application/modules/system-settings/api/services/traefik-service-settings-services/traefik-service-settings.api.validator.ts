import type { AxiosResponse } from "axios";
import { z } from "zod";
import { TraefikServiceSettingsEntitySchema } from "~/system-settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

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
    findOne = (response: AxiosResponse): TraefikServiceSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data, meta };
    };

    updateOne = (response: AxiosResponse): TraefikServiceSettings_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
