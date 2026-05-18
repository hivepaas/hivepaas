import { type AxiosResponse } from "axios";
import { z } from "zod";
import { SystemBackupSettingsEntitySchema } from "~/system-settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { SystemBackup_FindOne_Res, SystemBackup_UpdateOne_Res } from "./system-backup.api.contracts";

const FindOneSchema = z.object({
    data: SystemBackupSettingsEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class SystemBackupApiValidator {
    findOne = (response: AxiosResponse): SystemBackup_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data, meta };
    };

    updateOne = (response: AxiosResponse): SystemBackup_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
