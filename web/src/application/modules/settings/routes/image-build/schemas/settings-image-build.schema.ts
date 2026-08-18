import { z } from "zod";

const optionalTrimmedString = z
    .string()
    .optional()
    .transform(value => {
        const trimmedValue = value?.trim();
        if (!trimmedValue) {
            return undefined;
        }

        return trimmedValue;
    });

export const SettingsImageBuildFormSchema = z.object({
    workers: z.object({
        nodes: z.array(
            z.object({
                id: z.string(),
                name: z.string(),
            }),
        ),
        nodeLabels: z.array(z.string()),
        maxParallelism: z.number(),
    }),
    resources: z.object({
        cpus: z.number().optional(),
        mem: optionalTrimmedString,
        memSwap: optionalTrimmedString,
        shmSize: optionalTrimmedString,
    }),
    sources: z.object({
        repoCache: z.boolean(),
    }),
    noCache: z.boolean(),
    noVerbose: z.boolean(),
});

export type SettingsImageBuildFormSchemaInput = z.input<typeof SettingsImageBuildFormSchema>;
export type SettingsImageBuildFormSchemaOutput = z.output<typeof SettingsImageBuildFormSchema>;

export const emptySettingsImageBuildFormDefaults: SettingsImageBuildFormSchemaInput = {
    workers: {
        nodes: [],
        nodeLabels: [],
        maxParallelism: 0,
    },
    resources: {
        cpus: undefined,
        mem: undefined,
        memSwap: undefined,
        shmSize: undefined,
    },
    sources: {
        repoCache: false,
    },
    noCache: false,
    noVerbose: false,
};
