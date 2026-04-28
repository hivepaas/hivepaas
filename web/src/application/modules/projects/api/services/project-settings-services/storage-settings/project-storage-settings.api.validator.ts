import { type AxiosResponse } from "axios";
import { z } from "zod";
import { StorageSettingsEntitySchema } from "~/settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { ProjectStorageSettings_FindOne_Res } from "./project-storage-settings.api.contracts";

const FindOneSchema = z.object({
    data: StorageSettingsEntitySchema,
    meta: BaseMetaApiSchema.nullable(),
});

export class ProjectStorageSettingsApiValidator {
    findOne = (response: AxiosResponse): ProjectStorageSettings_FindOne_Res => {
        return parseApiResponse({ response, schema: FindOneSchema });
    };
}
