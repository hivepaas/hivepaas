import { type AxiosResponse } from "axios";
import { z } from "zod";
import { SystemBackupRepoCleanupSettingsEntitySchema } from "~/system-settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    SystemBackupRepoCleanup_Execute_Res,
    SystemBackupRepoCleanup_FindOne_Res,
    SystemBackupRepoCleanup_UpdateOne_Res,
} from "./system-backup-repo-cleanup.api.contracts";

const FindOneSchema = z.object({
    data: SystemBackupRepoCleanupSettingsEntitySchema,
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

export class SystemBackupRepoCleanupApiValidator {
    findOne = (response: AxiosResponse): SystemBackupRepoCleanup_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data, meta };
    };

    updateOne = (response: AxiosResponse): SystemBackupRepoCleanup_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    execute = (response: AxiosResponse): SystemBackupRepoCleanup_Execute_Res => {
        const { data, meta } = parseApiResponse({ response, schema: ExecuteSchema });
        return { data, meta };
    };
}
