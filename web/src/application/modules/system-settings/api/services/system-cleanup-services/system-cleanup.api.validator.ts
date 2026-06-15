import { type AxiosResponse } from "axios";
import { z } from "zod";
import { SystemCleanupSettingsEntitySchema } from "~/system-settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    SystemCleanup_ClearBuildCache_Res,
    SystemCleanup_ClearRepoCache_Res,
    SystemCleanup_Execute_Res,
    SystemCleanup_FindOne_Res,
    SystemCleanup_FindRepoCache_Res,
    SystemCleanup_UpdateOne_Res,
} from "./system-cleanup.api.contracts";

const FindOneSchema = z.object({
    data: SystemCleanupSettingsEntitySchema,
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

const RepoCacheInfoSchema = z.object({
    totalFiles: z.number(),
    totalSizeBytes: z.number(),
});

const FindRepoCacheSchema = z.object({
    data: RepoCacheInfoSchema,
    meta: BaseMetaApiSchema.nullish(),
});

const ClearRepoCacheResultSchema = z.object({
    filesDeleted: z.number(),
    spaceReclaimed: z.number(),
});

const ClearRepoCacheSchema = z.object({
    data: ClearRepoCacheResultSchema,
    meta: BaseMetaApiSchema.nullish(),
});

const ClearBuildCacheWireResultSchema = z.object({
    filesDeleted: z.number(),
    spaceReclaimed: z.number(),
});

const ClearBuildCacheSchema = z.object({
    data: ClearBuildCacheWireResultSchema,
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

    execute = (response: AxiosResponse): SystemCleanup_Execute_Res => {
        return parseApiResponse({ response, schema: ExecuteSchema });
    };

    findRepoCache = (response: AxiosResponse): SystemCleanup_FindRepoCache_Res => {
        return parseApiResponse({ response, schema: FindRepoCacheSchema });
    };

    clearRepoCache = (response: AxiosResponse): SystemCleanup_ClearRepoCache_Res => {
        return parseApiResponse({ response, schema: ClearRepoCacheSchema });
    };

    clearBuildCache = (response: AxiosResponse): SystemCleanup_ClearBuildCache_Res => {
        const { data, meta } = parseApiResponse({ response, schema: ClearBuildCacheSchema });

        return {
            data: {
                cachesDeleted: data.filesDeleted,
                spaceReclaimed: data.spaceReclaimed,
            },
            meta,
        };
    };
}
