import { type AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    HivePaaSLoggingPerformance_FindOne_Res,
    HivePaaSLoggingPerformance_UpdateOne_Res,
} from "./hivepaas-logging-performance.api.contracts";

const CapacitySchema = z.enum(["small", "medium", "large"]);

const NodeStatusSchema = z
    .object({
        time: z.string().catch(""),
        wanted: z.boolean().catch(false),
        running: z.boolean().catch(false),
        apps: z.number().catch(0),
        ok: z.boolean().catch(false),
        reasons: z
            .array(z.string())
            .nullish()
            .transform(value => value ?? []),
        kernel: z.string().catch(""),
        memTotalMb: z.number().catch(0),
        memAvailableMb: z.number().catch(0),
        capacity: z.string().catch(""),
    })
    .nullish()
    .transform(value => value ?? null);

const NodeSchema = z.object({
    id: z.string(),
    hostname: z.string().catch(""),
    role: z.string().catch(""),
    state: z.string().catch(""),
    availability: z.string().catch(""),
    memoryBytes: z.number().catch(0),
    recommended: CapacitySchema.catch("small"),
    enabled: z.boolean().catch(false),
    capacity: z.enum(["auto", "small", "medium", "large"]).catch("auto"),
    status: NodeStatusSchema,
});

const FindOneSchema = z.object({
    data: z.object({
        enabled: z.boolean().catch(false),
        configured: z.boolean().catch(false),
        logsStored: z.boolean().catch(false),
        updateVer: z.number().catch(0),
        statusReason: z
            .enum(["off", "logs-not-stored", "unreadable"])
            .nullish()
            .catch(null)
            .transform(value => value ?? null),
        capacities: z
            .array(
                z.object({
                    capacity: CapacitySchema,
                    memoryMiB: z.number().catch(0),
                    tracked: z.number().catch(0),
                }),
            )
            .nullish()
            .transform(value => value ?? []),
        nodes: z
            .array(NodeSchema)
            .nullish()
            .transform(value => value ?? []),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class HivePaaSLoggingPerformanceApiValidator {
    findOne = (response: AxiosResponse): HivePaaSLoggingPerformance_FindOne_Res => {
        return parseApiResponse({ response, schema: FindOneSchema });
    };

    updateOne = (response: AxiosResponse): HivePaaSLoggingPerformance_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
