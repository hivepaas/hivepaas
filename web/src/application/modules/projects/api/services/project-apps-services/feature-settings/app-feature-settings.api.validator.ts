import type { AxiosResponse } from "axios";
import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { AppFeatureSettings_FindOne_Res } from "./app-feature-settings.api.contracts";

const FeatureToggleSettingsSchema = z
    .object({
        enabled: z.boolean().optional().default(true),
    })
    .nullish()
    .transform(value => ({ enabled: value?.enabled ?? true }));

const AppFeaturePreviewAppRefSchema = z.object({
    id: z.string(),
    name: z.string().optional().default(""),
    photo: z.string().optional(),
    key: z.string().optional(),
    status: z.string().optional(),
    env: z.string().optional(),
});

const AppFeaturePreviewCommandRefSchema = z.object({
    id: z.string(),
    name: z.string().optional().default(""),
    type: z.string().optional(),
});

const AppFeaturePreviewSettingsSchema = z
    .object({
        enabled: z.boolean().optional().default(true),
        creationDelay: z.string().optional().default(""),
        appsToClone: z.array(AppFeaturePreviewAppRefSchema).optional().default([]),
        autoCloneApps: z.boolean().optional().default(false),
        commands: z.array(AppFeaturePreviewCommandRefSchema).optional().default([]),
    })
    .nullish()
    .transform(value => ({
        enabled: value?.enabled ?? true,
        creationDelay: value?.creationDelay ?? "",
        appsToClone: value?.appsToClone ?? [],
        autoCloneApps: value?.autoCloneApps ?? false,
        commands: value?.commands ?? [],
    }));

const AppFeatureSettingsSchema = SettingsBaseEntitySchema.extend({
    loggingSettings: FeatureToggleSettingsSchema,
    schedJobSettings: FeatureToggleSettingsSchema,
    terminalSettings: FeatureToggleSettingsSchema,
    previewSettings: AppFeaturePreviewSettingsSchema,
});

const FindOneSchema = z.object({
    data: AppFeatureSettingsSchema,
    meta: BaseMetaApiSchema.nullable(),
});

export class AppFeatureSettingsApiValidator {
    findOne = (response: AxiosResponse): AppFeatureSettings_FindOne_Res => {
        return parseApiResponse({ response, schema: FindOneSchema });
    };
}
