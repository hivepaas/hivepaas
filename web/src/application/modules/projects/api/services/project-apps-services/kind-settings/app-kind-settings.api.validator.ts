import { type AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { AppKindSettings_FindOne_Res } from "./app-kind-settings.api.contracts";

const OptionalStringSchema = z
    .string()
    .nullish()
    .transform(value => value ?? "");

const CategorySchema = z.enum(["webapp", "database", "cache", "storage"]);
const SSLModeSchema = z.enum(["disable", "prefer", "require", "verify-ca", "verify-full"]);

const DatabaseSchema = z
    .object({
        dbName: OptionalStringSchema,
        username: OptionalStringSchema,
        password: OptionalStringSchema,
        rootPassword: OptionalStringSchema,
        sslMode: SSLModeSchema.optional().default("disable"),
        sslCert: z
            .object({
                id: z.string(),
                name: z.string().optional(),
            })
            .nullish(),
        tlsPassthrough: z.boolean().optional().default(false),
    })
    .nullish();

const CacheSchema = z
    .object({
        password: OptionalStringSchema,
        maxMemory: OptionalStringSchema,
        evictionRule: OptionalStringSchema,
        persistenceMode: OptionalStringSchema,
        sslCert: z
            .object({
                id: z.string(),
                name: z.string().optional(),
            })
            .nullish(),
    })
    .nullish();

const StorageSchema = z
    .object({
        keyId: OptionalStringSchema,
        secret: OptionalStringSchema,
        bucket: OptionalStringSchema,
        region: OptionalStringSchema,
    })
    .nullish();

const WebappSchema = z.object({}).nullish();

const AppKindSettingsSchema = z.object({
    category: CategorySchema.optional().default("webapp"),
    engine: OptionalStringSchema,
    port: z.number().optional().default(0),
    version: OptionalStringSchema,
    webapp: WebappSchema,
    database: DatabaseSchema,
    cache: CacheSchema,
    storage: StorageSchema,
    secretMasked: z.boolean().optional().default(false),
    updateVer: z.number().optional().default(0),
});

const FindOneSchema = z.object({
    data: AppKindSettingsSchema.nullish().transform(
        val =>
            val ?? {
                category: "webapp" as const,
                engine: "",
                port: 0,
                version: "",
                webapp: null,
                database: null,
                cache: null,
                storage: null,
                secretMasked: false,
                updateVer: 0,
            },
    ),
    meta: BaseMetaApiSchema.nullish(),
});

export class AppKindSettingsApiValidator {
    findOne = (response: AxiosResponse): AppKindSettings_FindOne_Res => {
        return parseApiResponse({ response, schema: FindOneSchema });
    };
}
