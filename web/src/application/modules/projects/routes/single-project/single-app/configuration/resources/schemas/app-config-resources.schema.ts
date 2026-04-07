import { z } from "zod";

export const GenericResourceFormSchema = z.object({
    kind: z.string(),
    value: z.string(),
});

export const UlimitFormSchema = z.object({
    name: z.string(),
    hard: z.coerce.number(),
    soft: z.coerce.number(),
});

export const SysctlFormSchema = z.object({
    name: z.string(),
    value: z.string(),
});

export const AppConfigResourcesFormSchema = z.object({
    reservations: z.object({
        cpus: z.number().optional(),
        memoryMB: z.number().optional(),
        genericResources: z.array(GenericResourceFormSchema),
    }),

    limits: z.object({
        cpus: z.number().optional(),
        memoryMB: z.number().optional(),
        pids: z.number().optional(),
    }),

    ulimits: z.array(UlimitFormSchema),

    capabilities: z.object({
        capabilityAdd: z.string(),
        capabilityDrop: z.string(),
        enableGPU: z.boolean(),
        oomScoreAdj: z.number().optional(),
        sysctls: z.array(SysctlFormSchema),
    }),
});

export type AppConfigResourcesFormSchemaInput = z.input<typeof AppConfigResourcesFormSchema>;
export type AppConfigResourcesFormSchemaOutput = z.output<typeof AppConfigResourcesFormSchema>;

export const emptyAppConfigResourcesFormDefaults: AppConfigResourcesFormSchemaInput = {
    reservations: {
        cpus: undefined,
        memoryMB: undefined,
        genericResources: [],
    },

    limits: {
        cpus: undefined,
        memoryMB: undefined,
        pids: undefined,
    },

    ulimits: [],

    capabilities: {
        capabilityAdd: "",
        capabilityDrop: "",
        enableGPU: false,
        oomScoreAdj: undefined,
        sysctls: [],
    },
};
