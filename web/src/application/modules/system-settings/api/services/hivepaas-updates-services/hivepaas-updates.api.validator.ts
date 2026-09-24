import { type AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    HivePaaSUpdates_FindPlan_Res,
    HivePaaSUpdates_FindReleaseInfo_Res,
} from "./hivepaas-updates.api.contracts";

const DateSchema = z.coerce
    .date()
    .nullish()
    .catch(null)
    .transform(value => value ?? null);
const ChannelSchema = z.enum(["stable", "beta"]).catch("stable");

const RunningReleaseSchema = z.object({
    appVersion: z.string().catch(""),
    channel: ChannelSchema,
    releaseDate: DateSchema,
});

const PublishedReleaseSchema = z.object({
    appVersion: z.string().catch(""),
    releaseDate: DateSchema,
    notesUrl: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    canUpdate: z.boolean().catch(false),
    relation: z.enum(["newer", "same", "older", ""]).catch(""),
});

const ReleaseInfoSchema = z.object({
    data: z.object({
        current: RunningReleaseSchema.nullish().transform(value => value ?? null),
        stable: PublishedReleaseSchema.nullish().transform(value => value ?? null),
        beta: PublishedReleaseSchema.nullish().transform(value => value ?? null),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const PlanSchema = z.object({
    data: z.object({
        current: RunningReleaseSchema.nullish().transform(value => value ?? null),
        target: z.object({
            appVersion: z.string(),
            channel: ChannelSchema,
            releaseDate: DateSchema,
            notesUrl: z
                .string()
                .nullish()
                .transform(value => value ?? ""),
        }),
        components: z
            .array(
                z.object({
                    key: z.string(),
                    currentImage: z.string().catch(""),
                    targetImage: z.string().catch(""),
                    change: z.enum(["none", "update", "major", "blocked", "not-deployed"]).catch("none"),
                    reason: z.string().catch(""),
                    requiresBackup: z.boolean().catch(false),
                    interruptsTraffic: z.boolean().catch(false),
                }),
            )
            .nullish()
            .transform(value => value ?? []),
        blocked: z.boolean().catch(false),
        requiresBackup: z.boolean().catch(false),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

export class HivePaaSUpdatesApiValidator {
    findReleaseInfo = (response: AxiosResponse): HivePaaSUpdates_FindReleaseInfo_Res => {
        const { data, meta } = parseApiResponse({ response, schema: ReleaseInfoSchema });
        return { data, meta };
    };

    findPlan = (response: AxiosResponse): HivePaaSUpdates_FindPlan_Res => {
        const { data, meta } = parseApiResponse({ response, schema: PlanSchema });
        return { data, meta };
    };
}
