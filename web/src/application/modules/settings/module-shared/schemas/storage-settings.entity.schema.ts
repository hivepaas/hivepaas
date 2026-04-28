import { z } from "zod";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

const NamedObjectSchema = z.object({
    id: z.string(),
    name: z.string(),
});

const StorageBindSettingsSchema = z.object({
    enabled: z.boolean().optional(),
    baseDirs: z.array(z.string()).optional(),
    baseSubpath: z.string().optional(),
    appsMustUseSubPaths: z.boolean().optional(),
});

const StorageVolumeSettingsSchema = z.object({
    enabled: z.boolean().optional(),
    volumes: z.array(NamedObjectSchema).optional(),
    baseSubpath: z.string().optional(),
    appsMustUseSubPaths: z.boolean().optional(),
});

const StorageClusterVolumeSettingsSchema = z.object({
    enabled: z.boolean().optional(),
    volumes: z.array(NamedObjectSchema).optional(),
    baseSubpath: z.string().optional(),
    appsMustUseSubPaths: z.boolean().optional(),
});

const StorageTmpfsSettingsSchema = z.object({
    enabled: z.boolean().optional(),
    maxSize: z.number().optional(),
});

export const StorageSettingsEntitySchema = SettingsBaseEntitySchema.extend({
    bindSettings: StorageBindSettingsSchema.optional(),
    volumeSettings: StorageVolumeSettingsSchema.optional(),
    clusterVolumeSettings: StorageClusterVolumeSettingsSchema.optional(),
    tmpfsSettings: StorageTmpfsSettingsSchema.optional(),
});
