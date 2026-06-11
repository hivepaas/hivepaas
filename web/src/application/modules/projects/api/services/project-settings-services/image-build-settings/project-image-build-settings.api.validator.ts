import { type AxiosResponse } from "axios";
import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    ProjectImageBuildSettings_FindOne_Res,
    ProjectImageBuildSettings_FindRepoCache_Res,
} from "./project-image-build-settings.api.contracts";

const ImageBuildResourcesSchema = z
    .object({
        cpus: z.number().optional(),
        mem: z.string().optional(),
        memSwap: z.string().optional(),
        shmSize: z.string().optional(),
    })
    .nullish();

const ImageBuildSourcesSchema = z
    .object({
        repoCache: z.boolean().optional(),
    })
    .nullish();

const ImageBuildSettingsSchema = z.object({
    id: z.string(),
    type: z.string(),
    name: z.string(),
    kind: z.string().optional(),
    status: z.nativeEnum(ESettingStatus),
    inherited: z.boolean().optional(),
    availableInProjects: z.boolean().optional(),
    default: z.boolean().optional(),
    updateVer: z.number(),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date(),
    expireAt: z.coerce.date().nullish(),
    resources: ImageBuildResourcesSchema,
    sources: ImageBuildSourcesSchema,
    noCache: z.boolean().optional(),
    noVerbose: z.boolean().optional(),
});

const FindOneSchema = z.object({
    data: ImageBuildSettingsSchema,
    meta: BaseMetaApiSchema.nullable(),
});

const RepoCacheInfoSchema = z.object({
    totalFiles: z.number(),
    totalSizeBytes: z.number(),
});

const FindRepoCacheSchema = z.object({
    data: RepoCacheInfoSchema,
    meta: BaseMetaApiSchema.nullable(),
});

export class ProjectImageBuildSettingsApiValidator {
    findOne = (response: AxiosResponse): ProjectImageBuildSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });

        return {
            data: {
                ...data,
                expireAt: data.expireAt ?? null,
                resources: {
                    cpus: data.resources?.cpus,
                    mem: data.resources?.mem,
                    memSwap: data.resources?.memSwap,
                    shmSize: data.resources?.shmSize,
                },
                sources: {
                    repoCache: data.sources?.repoCache ?? false,
                },
                noCache: data.noCache ?? false,
                noVerbose: data.noVerbose ?? false,
            },
            meta,
        };
    };

    findRepoCache = (response: AxiosResponse): ProjectImageBuildSettings_FindRepoCache_Res => {
        return parseApiResponse({ response, schema: FindRepoCacheSchema });
    };
}
