import { z } from "zod";

import { parseDataSizeToBytes } from "@application/shared/utils";

export const AppConfigKindSettingsDatabaseFormSchema = z.object({
    dbName: z.string().max(100, "Database name must be at most 100 characters"),
    username: z.string().max(100, "Username must be at most 100 characters"),
    password: z.string(),
    rootPassword: z.string(),
    sslMode: z.enum(["disable", "prefer", "require", "verify-ca", "verify-full"]),
    sslCert: z
        .object({
            id: z.string(),
            name: z.string().optional(),
        })
        .nullish(),
    tlsPassthrough: z.boolean(),
});

export const AppConfigKindSettingsCacheFormSchema = z.object({
    password: z.string(),
    maxMemory: z.string().refine(val => !val || parseDataSizeToBytes(val) !== null, {
        message: "Invalid data size format (e.g. 256mb, 1gb)",
    }),
    evictionRule: z.string().max(100, "Eviction rule must be at most 100 characters"),
    persistenceMode: z.string().max(100, "Persistence mode must be at most 100 characters"),
    sslCert: z
        .object({
            id: z.string(),
            name: z.string().optional(),
        })
        .nullish(),
});

export const AppConfigKindSettingsStorageFormSchema = z.object({
    keyId: z.string().max(100, "Key ID must be at most 100 characters"),
    secret: z.string(),
    bucket: z.string().max(100, "Bucket must be at most 100 characters"),
    region: z.string().max(100, "Region must be at most 100 characters"),
});

export const AppConfigKindSettingsFormSchema = z.object({
    category: z.enum(["webapp", "database", "cache", "storage"]),
    engine: z.string().min(1, "Engine is required").max(100, "Engine must be at most 100 characters"),
    port: z
        .number({ invalid_type_error: "Port is required" })
        .int()
        .min(1, "Port must be at least 1")
        .max(65535, "Port must be at most 65535"),
    version: z.string().max(50, "Version must be at most 50 characters"),
    database: AppConfigKindSettingsDatabaseFormSchema.optional(),
    cache: AppConfigKindSettingsCacheFormSchema.optional(),
    storage: AppConfigKindSettingsStorageFormSchema.optional(),
});

export type AppConfigKindSettingsFormSchemaInput = z.input<typeof AppConfigKindSettingsFormSchema>;
export type AppConfigKindSettingsFormSchemaOutput = z.output<typeof AppConfigKindSettingsFormSchema>;
