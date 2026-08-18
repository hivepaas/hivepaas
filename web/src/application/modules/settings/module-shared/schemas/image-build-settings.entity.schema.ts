import { z } from "zod";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

const NamedObjectSchema = z.object({
    id: z.string(),
    name: z.string(),
});

const ImageBuildWorkerSettingsSchema = z
    .object({
        nodes: z.array(NamedObjectSchema).nullish(),
        nodeLabels: z.array(z.string()).nullish(),
        maxParallelism: z.number().nullish(),
    })
    .nullish();

const ImageBuildResourceSettingsSchema = z
    .object({
        cpus: z.number().optional(),
        mem: z.string().optional(),
        memSwap: z.string().optional(),
        shmSize: z.string().optional(),
    })
    .nullish();

const ImageBuildSourceSettingsSchema = z
    .object({
        repoCache: z.boolean().optional(),
    })
    .nullish();

export const ImageBuildSettingsEntitySchema = SettingsBaseEntitySchema.extend({
    workers: ImageBuildWorkerSettingsSchema,
    resources: ImageBuildResourceSettingsSchema,
    sources: ImageBuildSourceSettingsSchema,
    noCache: z.boolean().optional(),
    noVerbose: z.boolean().optional(),
});

export const ImageBuildRepoCacheInfoSchema = z.object({
    totalFiles: z.number(),
    totalSizeBytes: z.number(),
});

export const ImageBuildRepoCacheClearResultSchema = z.object({
    filesDeleted: z.number(),
    spaceReclaimed: z.number(),
});
