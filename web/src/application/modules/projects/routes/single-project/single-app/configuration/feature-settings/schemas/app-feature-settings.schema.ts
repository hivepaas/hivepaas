import { z } from "zod";

const AppFeaturePreviewAppRefSchema = z.object({
    id: z.string(),
    name: z.string(),
    photo: z.string().optional(),
});

const AppFeaturePreviewCommandRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

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
    previewSettings: z.object({
        enabled: z.boolean(),
        creationDelay: z.string().trim(),
        appsToClone: z.array(AppFeaturePreviewAppRefSchema),
        autoCloneApps: z.boolean(),
        commands: z.array(AppFeaturePreviewCommandRefSchema),
    }),
});

export type AppFeatureSettingsFormSchemaInput = z.input<typeof AppFeatureSettingsFormSchema>;
export type AppFeatureSettingsFormSchemaOutput = z.output<typeof AppFeatureSettingsFormSchema>;

export const DEFAULT_PREVIEW_CREATION_DELAY = "30s";
export const FEATURE_SETTINGS_TITLE_WIDTH = 295;

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
    previewSettings: {
        enabled: true,
        creationDelay: DEFAULT_PREVIEW_CREATION_DELAY,
        appsToClone: [],
        autoCloneApps: false,
        commands: [],
    },
};
