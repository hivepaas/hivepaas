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

export const ProjectImageBuildSettingsFormSchema = z.object({
    resources: z.object({
        cpus: z.number().optional(),
        mem: optionalTrimmedString,
        memSwap: optionalTrimmedString,
        shmSize: optionalTrimmedString,
    }),
    sources: z.object({
        checkoutMaxDepth: z.number().optional(),
        repoCache: z.boolean(),
    }),
    noCache: z.boolean(),
    noVerbose: z.boolean(),
});

export type ProjectImageBuildSettingsFormSchemaInput = z.input<typeof ProjectImageBuildSettingsFormSchema>;
export type ProjectImageBuildSettingsFormSchemaOutput = z.output<typeof ProjectImageBuildSettingsFormSchema>;

export const emptyProjectImageBuildSettingsFormDefaults: ProjectImageBuildSettingsFormSchemaInput = {
    resources: {
        cpus: undefined,
        mem: undefined,
        memSwap: undefined,
        shmSize: undefined,
    },
    sources: {
        checkoutMaxDepth: undefined,
        repoCache: false,
    },
    noCache: false,
    noVerbose: false,
};
