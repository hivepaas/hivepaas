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

const AppFeatureSettingsSchema = SettingsBaseEntitySchema.extend({
    loggingSettings: FeatureToggleSettingsSchema,
    schedJobSettings: FeatureToggleSettingsSchema,
    terminalSettings: FeatureToggleSettingsSchema,
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
