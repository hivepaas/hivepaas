import type { AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { AppAutoscale_FindOne_Res, AppAutoscale_UpdateOne_Res } from "./app-autoscale.api.contracts";

/** A reason the API may leave out: null when it does, or sends none. */
const reason = z
    .string()
    .nullish()
    .catch(null)
    .transform(value => (value === "" ? null : (value ?? null)));

const AppAutoscaleSchema = z.object({
    isFunction: z.boolean().catch(false),
    enabled: z.boolean().catch(false),
    minReplicas: z.number().catch(1),
    maxReplicas: z.number().catch(5),
    target: z.number().catch(70),
    requestsTarget: z.number().catch(0),
    cpuTarget: z.number().catch(0),
    scaleInDelay: z.string().catch("5m"),
    replicas: z.number().catch(0),
    paused: reason,
    requestsUnavailable: reason,
    cpuUnavailable: reason,
    pending: z.number().catch(0),
    writableMounts: z.boolean().catch(false),
    events: z
        .array(
            z.object({
                time: z.string(),
                from: z.number().catch(0),
                to: z.number().catch(0),
                inFlight: z.number().catch(0),
                calls: z.number().catch(0),
                throttled: z.number().catch(0),
                requests: z.number().catch(0),
                cpu: z.number().catch(0),
                reason: z.string().catch(""),
            }),
        )
        .nullish()
        .transform(value => value ?? []),
    updateVer: z.number().catch(0),
});

const FindOneSchema = z.object({
    data: AppAutoscaleSchema,
    meta: BaseMetaApiSchema.nullish(),
});

const UpdateOneSchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class AppAutoscaleApiValidator {
    findOne = (response: AxiosResponse): AppAutoscale_FindOne_Res => {
        return parseApiResponse({ response, schema: FindOneSchema });
    };

    updateOne = (response: AxiosResponse): AppAutoscale_UpdateOne_Res => {
        const { meta } = parseApiResponse({ response, schema: UpdateOneSchema });
        return { data: { type: "success", warning: meta?.warning ?? null } };
    };
}
