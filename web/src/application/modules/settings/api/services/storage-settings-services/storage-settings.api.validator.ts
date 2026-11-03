import { z } from "zod";
import { StorageSettingsEntitySchema } from "~/settings/module-shared/schemas";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { StorageSettings_FindOne_Res } from "./storage-settings.api.contracts";

const FindOneSchema = z.object({
    data: StorageSettingsEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

export class StorageSettingsApiValidator {
    findOne = (response: ApiHttpResponse): StorageSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneSchema,
        });

        return { data, meta };
    };
}
