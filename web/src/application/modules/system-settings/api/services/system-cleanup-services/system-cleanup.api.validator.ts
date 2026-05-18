import { type AxiosResponse } from "axios";
import { z } from "zod";
import { SystemCleanupSettingsEntitySchema } from "~/system-settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { SystemCleanup_FindOne_Res, SystemCleanup_UpdateOne_Res } from "./system-cleanup.api.contracts";

const FindOneSchema = z.object({
    data: SystemCleanupSettingsEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class SystemCleanupApiValidator {
    findOne = (response: AxiosResponse): SystemCleanup_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data, meta };
    };

    updateOne = (response: AxiosResponse): SystemCleanup_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
