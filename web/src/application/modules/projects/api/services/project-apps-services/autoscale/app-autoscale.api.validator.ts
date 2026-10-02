import type { AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { AppAutoscale_FindOne_Res, AppAutoscale_UpdateOne_Res } from "./app-autoscale.api.contracts";

const AppAutoscaleSchema = z.object({
    enabled: z.boolean().catch(false),
    minReplicas: z.number().catch(1),
    maxReplicas: z.number().catch(5),
    target: z.number().catch(70),
    scaleInDelay: z.string().catch("5m"),
    replicas: z.number().catch(0),
    paused: z
        .enum([
            "disabled",
            "apps-not-collected",
            "no-query-endpoint",
            "driver-unreadable",
            "identity-missing",
            "not-replicated",
        ])
        .nullish()
        .catch(null)
        .transform(value => value ?? null),
    events: z
        .array(
            z.object({
                time: z.string(),
                from: z.number().catch(0),
                to: z.number().catch(0),
                inFlight: z.number().catch(0),
                calls: z.number().catch(0),
                throttled: z.number().catch(0),
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
