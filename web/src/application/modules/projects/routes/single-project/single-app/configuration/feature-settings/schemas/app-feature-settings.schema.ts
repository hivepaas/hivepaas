import { z } from "zod";

export const AppFeatureSettingsFormSchema = z.object({
    loggingSettings: z.object({
        enabled: z.boolean(),
    }),
    schedJobSettings: z.object({
        enabled: z.boolean(),
    }),
    terminalSettings: z.object({
        enabled: z.boolean(),
    }),
});

export type AppFeatureSettingsFormSchemaInput = z.input<typeof AppFeatureSettingsFormSchema>;
export type AppFeatureSettingsFormSchemaOutput = z.output<typeof AppFeatureSettingsFormSchema>;

export const emptyAppFeatureSettingsFormDefaults: AppFeatureSettingsFormSchemaInput = {
    loggingSettings: {
        enabled: true,
    },
    schedJobSettings: {
        enabled: true,
    },
    terminalSettings: {
        enabled: true,
    },
};
