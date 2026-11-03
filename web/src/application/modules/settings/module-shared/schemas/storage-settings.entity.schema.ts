import { z } from "zod";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

const NamedObjectSchema = z.object({
    id: z.string(),
    name: z.string(),
});

const StorageBindSettingsSchema = z.object({
    enabled: z.boolean().optional(),
    baseDirs: z
        .array(z.string())
        .nullish()
        .transform(value => value ?? []),
    subpathTemplate: z.string().optional(),
    subpathRequired: z.string().optional(),
});

const StorageVolumeSettingsSchema = z.object({
    enabled: z.boolean().optional(),
    volumes: z
        .array(NamedObjectSchema)
        .nullish()
        .transform(value => value ?? []),
    subpathTemplate: z.string().optional(),
    subpathRequired: z.string().optional(),
});

const StorageClusterVolumeSettingsSchema = z.object({
    enabled: z.boolean().optional(),
    volumes: z
        .array(NamedObjectSchema)
        .nullish()
        .transform(value => value ?? []),
    subpathTemplate: z.string().optional(),
    subpathRequired: z.string().optional(),
});

const StorageTmpfsSettingsSchema = z.object({
    enabled: z.boolean().optional(),
    maxSize: z.string().optional(),
});

export const StorageSettingsEntitySchema = SettingsBaseEntitySchema.extend({
    bindSettings: StorageBindSettingsSchema.nullish().transform(value => value ?? undefined),
    volumeSettings: StorageVolumeSettingsSchema.nullish().transform(value => value ?? undefined),
    clusterVolumeSettings: StorageClusterVolumeSettingsSchema.nullish().transform(value => value ?? undefined),
    tmpfsSettings: StorageTmpfsSettingsSchema.nullish().transform(value => value ?? undefined),
});
