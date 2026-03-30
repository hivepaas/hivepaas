import { type AxiosResponse } from "axios";
import { z } from "zod";
import { EAppServicePlacement, EServiceMode } from "~/projects/module-shared/enums";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import {
    type AppServiceSettings_FindOne_Res,
    type AppServiceSettings_UpdateOne_Res,
} from "./app-service-settings.api.contracts";

const ServiceModeSpecSchema = z.discriminatedUnion("mode", [
    z.object({ mode: z.literal(EServiceMode.Global) }),
    z.object({ mode: z.literal(EServiceMode.GlobalJob) }),
    z.object({
        mode: z.literal(EServiceMode.Replicated),
        serviceReplicas: z.number().nullable(),
    }),
    z.object({
        mode: z.literal(EServiceMode.ReplicatedJob),
        jobMaxConcurrent: z.number().nullable(),
        jobTotalCompletions: z.number().nullable(),
    }),
]);

const PlacementConstraintSchema = z.object({
    name: z.string(),
    value: z.string(),
    op: z.nativeEnum(EAppServicePlacement),
});

const PlacementPreferenceSchema = z.object({
    name: z.string(),
    value: z.string(),
});

const PlacementSchema = z.object({
    constraints: z.array(PlacementConstraintSchema),
    preferences: z.array(PlacementPreferenceSchema),
});

const AppServiceSettingsSchema = z.object({
    modeSpec: ServiceModeSpecSchema,
    placement: PlacementSchema.nullable(),
    updateVer: z.number(),
});

const FindOneSchema = z.object({
    data: AppServiceSettingsSchema,
    meta: BaseMetaApiSchema.nullable(),
});

const UpdateOneSchema = z.object({
    data: z.object({ type: z.literal("success") }),
    meta: BaseMetaApiSchema.nullable(),
});

export class AppServiceSettingsApiValidator {
    findOne = (response: AxiosResponse): AppServiceSettings_FindOne_Res => {
        return parseApiResponse({ response, schema: FindOneSchema });
    };

    updateOne = (response: AxiosResponse): AppServiceSettings_UpdateOne_Res => {
        return parseApiResponse({ response, schema: UpdateOneSchema });
    };
}
