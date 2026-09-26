import type { AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { AppDockerApiSettings_FindOne_Res } from "./app-docker-api-settings.api.contracts";

const StringListSchema = z
    .array(z.string())
    .nullish()
    .transform(value => value ?? []);

const LimitsSchema = z
    .object({
        containers: z.number().optional().default(0),
        // The server writes a zero size as "0".
        memory: z
            .string()
            .nullish()
            .transform(value => (value && value !== "0" ? value : "")),
        cpus: z.number().optional().default(0),
    })
    .nullish()
    .transform(value => value ?? { containers: 0, memory: "", cpus: 0 });

const AppDockerApiSettingsSchema = z.object({
    enabled: z.boolean(),
    mode: z
        .enum(["proxy", "host"])
        .nullish()
        .transform(value => value ?? "proxy"),
    images: StringListSchema,
    sharedDirs: StringListSchema,
    sharedVolumes: z
        .record(z.string(), z.string())
        .nullish()
        .transform(value => value ?? {}),
    networks: StringListSchema,
    allow: StringListSchema,
    limits: LimitsSchema,
    defaultLimits: LimitsSchema,
    hostMode: z
        .object({
            available: z.boolean(),
            blockedBy: z.enum(["", "switch", "admin"]).catch("admin"),
        })
        .nullish()
        .transform(value => value ?? { available: false, blockedBy: "switch" as const }),
    updateVer: z.number(),
});

const FindOneSchema = z.object({
    data: AppDockerApiSettingsSchema,
    meta: BaseMetaApiSchema.nullish(),
});

export class AppDockerApiSettingsApiValidator {
    findOne = (response: AxiosResponse): AppDockerApiSettings_FindOne_Res => {
        return parseApiResponse({ response, schema: FindOneSchema });
    };
}
