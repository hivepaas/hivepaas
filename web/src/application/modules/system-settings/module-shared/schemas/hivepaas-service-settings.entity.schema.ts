import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { ESettingType } from "@application/shared/enums";

const HivePaaSAppSettingsSchema = z.object({
    replicas: z.number(),
});

const HivePaaSWorkerSettingsSchema = z.object({
    replicas: z.number(),
    concurrency: z.number(),
    runWorkerInMainApp: z.boolean(),
});

const HivePaaSTaskSettingsSchema = z.object({
    taskCheckInterval: z.string(),
    taskCreateInterval: z.string(),
});

const HivePaaSHealthcheckSettingsSchema = z.object({
    baseInterval: z.string(),
});

export const HivePaaSServiceSettingsEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.literal(ESettingType.HivePaaSService),
    appSettings: HivePaaSAppSettingsSchema,
    workerSettings: HivePaaSWorkerSettingsSchema,
    taskSettings: HivePaaSTaskSettingsSchema,
    healthcheckSettings: HivePaaSHealthcheckSettingsSchema,
});
