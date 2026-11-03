import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { AppFeatureSettings_FindOne_Res } from "./app-feature-settings.api.contracts";

const FeatureToggleSettingsSchema = z
    .object({
        enabled: z.boolean().optional().default(true),
    })
    .nullish()
    .transform(value => ({ enabled: value?.enabled ?? true }));

/** A feature that is off unless the app asks for it. */
const FeatureOptInSettingsSchema = z
    .object({
        enabled: z.boolean().optional().default(false),
    })
    .nullish()
    .transform(value => ({ enabled: value?.enabled ?? false }));

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
        allowPRComments: z.boolean().optional().default(false),
        commands: z.array(AppFeaturePreviewCommandRefSchema).optional().default([]),
    })
    .nullish()
    .transform(value => ({
        enabled: value?.enabled ?? true,
        creationDelay: value?.creationDelay ?? "",
        appsToClone: value?.appsToClone ?? [],
        autoCloneApps: value?.autoCloneApps ?? false,
        allowPRComments: value?.allowPRComments ?? false,
        commands: value?.commands ?? [],
    }));

const AppFeatureSettingsSchema = SettingsBaseEntitySchema.extend({
    loggingSettings: FeatureToggleSettingsSchema,
    schedJobSettings: FeatureToggleSettingsSchema,
    terminalSettings: FeatureToggleSettingsSchema,
    previewSettings: AppFeaturePreviewSettingsSchema,
    performanceSettings: FeatureOptInSettingsSchema,
});

const FindOneSchema = z.object({
    data: AppFeatureSettingsSchema,
    meta: BaseMetaApiSchema.nullable(),
});

export class AppFeatureSettingsApiValidator {
    findOne = (response: ApiHttpResponse): AppFeatureSettings_FindOne_Res => {
        return parseApiResponse({ response, schema: FindOneSchema });
    };
}
