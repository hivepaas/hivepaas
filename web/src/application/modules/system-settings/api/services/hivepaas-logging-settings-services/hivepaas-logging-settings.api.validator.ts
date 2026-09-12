import { type AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    HivePaaSLoggingSettings_FindOne_Res,
    HivePaaSLoggingSettings_UpdateOne_Res,
} from "./hivepaas-logging-settings.api.contracts";

const EndpointSchema = z.object({
    url: z.string().catch(""),
    username: z.string().optional(),
    password: z.string().optional(),
    bearerToken: z.string().optional(),
    headers: z.record(z.string(), z.string()).optional(),
    tlsSkipVerify: z.boolean().optional(),
});

const SettingsSchema = z.object({
    enabled: z.boolean().catch(false),
    sources: z.object({
        apps: z.boolean().catch(false),
        hivepaas: z.boolean().catch(false),
        traefikAccess: z.boolean().catch(false),
        nodes: z.boolean().catch(false),
    }),
    collector: z.object({
        type: z.string().catch("vlagent"),
        managed: z.boolean().catch(true),
        image: z.string().optional(),
    }),
    backend: z.object({
        type: z.string().catch("victoria-logs"),
        managed: z.boolean().catch(true),
        ingest: EndpointSchema.nullish(),
        query: EndpointSchema.nullish(),
        victoriaLogs: z
            .object({
                image: z.string().optional(),
                nodeId: z.string().catch(""),
                volumeId: z.string().catch(""),
                retention: z.string().catch("30d"),
                maxDiskUsagePercent: z.number().optional(),
            })
            .nullish(),
    }),
    forwards: z
        .array(z.object({ name: z.string(), format: z.string().optional(), endpoint: EndpointSchema }))
        .nullish()
        .transform(value => value ?? []),
});

const StatusSchema = z
    .object({
        backendReady: z.boolean().catch(false),
        excludedApps: z
            .array(
                z.object({
                    appId: z.string(),
                    name: z.string(),
                    reason: z.string(),
                    driver: z.string().optional(),
                }),
            )
            .nullish()
            .transform(value => value ?? []),
    })
    .nullish()
    .transform(value => value ?? { backendReady: false, excludedApps: [] });

const FindOneSchema = z.object({
    data: SettingsSchema,
    status: StatusSchema,
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class HivePaaSLoggingSettingsApiValidator {
    findOne = (response: AxiosResponse): HivePaaSLoggingSettings_FindOne_Res => {
        const { data, status, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data: { settings: data, status }, meta };
    };

    updateOne = (response: AxiosResponse): HivePaaSLoggingSettings_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
