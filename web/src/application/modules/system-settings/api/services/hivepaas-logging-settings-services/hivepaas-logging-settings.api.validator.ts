import { type AxiosResponse } from "axios";
import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

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

const StatusSchema = z
    .object({
        collectorReady: z.boolean().catch(false),
        backendReady: z.boolean().catch(false),
    })
    .nullish()
    .transform(value => value ?? { collectorReady: false, backendReady: false });

const VictoriaLogsSchema = z
    .object({
        volume: z.object({ id: z.string(), name: z.string().optional() }).nullish(),
        volumeSubpath: z.string().optional(),
        retention: z.string().catch("30d"),
        maxDiskUsagePercent: z.number().optional(),
        cpuLimit: z.number().optional(),
        memoryLimit: z.string().optional(),
    })
    .nullish();

const SettingsSchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    id: z.string().catch(""),
    name: z
        .string()
        .nullish()
        .transform(val => val ?? ""),
    status: z.string().catch("active"),
    type: z.string().catch("logging"),
    updateVer: z.number().catch(0),
    createdAt: z.coerce.date().catch(() => new Date()),
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
    }),
    backend: z.object({
        type: z.string().catch("victoria-logs"),
        managed: z.boolean().catch(true),
        ingest: EndpointSchema.nullish(),
        query: EndpointSchema.nullish(),
        victoriaLogs: VictoriaLogsSchema,
    }),
    forwards: z
        .array(z.object({ name: z.string(), format: z.string().optional(), endpoint: EndpointSchema }))
        .nullish()
        .transform(value => value ?? []),
    secretMasked: z.boolean().optional(),
    loggingStatus: StatusSchema,
});

const FindOneSchema = z.object({
    data: SettingsSchema,
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class HivePaaSLoggingSettingsApiValidator {
    findOne = (response: AxiosResponse): HivePaaSLoggingSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data: { settings: data, loggingStatus: data.loggingStatus }, meta };
    };

    updateOne = (response: AxiosResponse): HivePaaSLoggingSettings_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
