import { z } from "zod";
import { SystemBackupRepoCleanupSettingsEntitySchema } from "~/system-settings/module-shared/schemas";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

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
    findOne = (response: ApiHttpResponse): SystemBackupRepoCleanup_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data, meta };
    };

    updateOne = (response: ApiHttpResponse): SystemBackupRepoCleanup_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    execute = (response: ApiHttpResponse): SystemBackupRepoCleanup_Execute_Res => {
        const { data, meta } = parseApiResponse({ response, schema: ExecuteSchema });
        return { data, meta };
    };
}
