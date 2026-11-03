import { z } from "zod";
import { SystemRegistryAuthRenewalSettingsEntitySchema } from "~/system-settings/module-shared/schemas";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    SystemRegistryAuthRenewal_Execute_Res,
    SystemRegistryAuthRenewal_FindOne_Res,
    SystemRegistryAuthRenewal_UpdateOne_Res,
} from "./system-registry-auth-renewal.api.contracts";

const FindOneSchema = z.object({
    data: SystemRegistryAuthRenewalSettingsEntitySchema,
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

export class SystemRegistryAuthRenewalApiValidator {
    findOne = (response: ApiHttpResponse): SystemRegistryAuthRenewal_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data, meta };
    };

    updateOne = (response: ApiHttpResponse): SystemRegistryAuthRenewal_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    execute = (response: ApiHttpResponse): SystemRegistryAuthRenewal_Execute_Res => {
        const { data, meta } = parseApiResponse({ response, schema: ExecuteSchema });
        return { data, meta };
    };
}
