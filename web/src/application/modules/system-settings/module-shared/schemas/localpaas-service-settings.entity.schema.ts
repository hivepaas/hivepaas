import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { ESettingType } from "@application/shared/enums";

const LocalPaaSAppSettingsSchema = z.object({
    replicas: z.number(),
});

const LocalPaaSWorkerSettingsSchema = z.object({
    replicas: z.number(),
    concurrency: z.number(),
    runWorkerInMainApp: z.boolean(),
});

const LocalPaaSTaskSettingsSchema = z.object({
    taskCheckInterval: z.string(),
    taskCreateInterval: z.string(),
});

const LocalPaaSHealthcheckSettingsSchema = z.object({
    baseInterval: z.string(),
});

export const LocalPaaSServiceSettingsEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.literal(ESettingType.LocalPaaSService),
    appSettings: LocalPaaSAppSettingsSchema,
    workerSettings: LocalPaaSWorkerSettingsSchema,
    taskSettings: LocalPaaSTaskSettingsSchema,
    healthcheckSettings: LocalPaaSHealthcheckSettingsSchema,
});
