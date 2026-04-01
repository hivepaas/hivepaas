import { z } from "zod";
import { EAppServicePlacement, EServiceMode } from "~/projects/module-shared/enums";

export const PlacementConstraintSchema = z.object({
    name: z.string(),
    op: z.nativeEnum(EAppServicePlacement),
    value: z.string(),
});

export const PlacementPreferenceSchema = z.object({
    name: z.string(),
    value: z.string(),
});

const BaseConfigSchema = z.object({
    constraints: z.array(PlacementConstraintSchema),
    preferences: z.array(PlacementPreferenceSchema),
});

export const AppConfigAvailabilitySchema = z.discriminatedUnion("mode", [
    BaseConfigSchema.extend({
        mode: z.literal(EServiceMode.Global),
    }),
    BaseConfigSchema.extend({
        mode: z.literal(EServiceMode.GlobalJob),
    }),
    BaseConfigSchema.extend({
        mode: z.literal(EServiceMode.Replicated),
        serviceReplicas: z.number().nullable(),
    }),
    BaseConfigSchema.extend({
        mode: z.literal(EServiceMode.ReplicatedJob),
        jobMaxConcurrent: z.number().nullable(),
        jobTotalCompletions: z.number().nullable(),
    }),
]);

export type AppConfigAvailabilitySchemaInput = z.input<typeof AppConfigAvailabilitySchema>;
export type AppConfigAvailabilitySchemaOutput = z.output<typeof AppConfigAvailabilitySchema>;
