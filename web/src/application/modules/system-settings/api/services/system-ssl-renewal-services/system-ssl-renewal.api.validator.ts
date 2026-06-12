import { type AxiosResponse } from "axios";
import { z } from "zod";
import { SystemSslRenewalSettingsEntitySchema } from "~/system-settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    SystemSslRenewal_Execute_Res,
    SystemSslRenewal_FindOne_Res,
    SystemSslRenewal_UpdateOne_Res,
} from "./system-ssl-renewal.api.contracts";

const FindOneSchema = z.object({
    data: SystemSslRenewalSettingsEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

const ExecuteSchema = z.object({
    data: z.object({
        task: z.object({
            id: z.string(),
        }),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

export class SystemSslRenewalApiValidator {
    findOne = (response: AxiosResponse): SystemSslRenewal_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data, meta };
    };

    updateOne = (response: AxiosResponse): SystemSslRenewal_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    execute = (response: AxiosResponse): SystemSslRenewal_Execute_Res => {
        const { data, meta } = parseApiResponse({ response, schema: ExecuteSchema });
        return { data, meta };
    };
}
