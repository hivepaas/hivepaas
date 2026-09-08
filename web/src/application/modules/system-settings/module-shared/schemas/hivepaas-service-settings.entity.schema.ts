import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { SettingsPendingChangeSchema } from "./settings-probation.entity.schema";

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

const HivePaaSPeriodicSettingsSchema = z.object({
    baseInterval: z.string(),
    batchSize: z.number(),
});

const HivePaaSProxySettingsSchema = z.object({
    proxyProvider: z.preprocess(value => value ?? "", z.string()),
    trustedIPs: z.preprocess(value => value ?? [], z.array(z.string())),
    proxyHops: z.preprocess(value => (typeof value === "number" ? value : 0), z.number()),
});

export const HivePaaSServiceSettingsEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.string(),
    appSettings: HivePaaSAppSettingsSchema,
    workerSettings: HivePaaSWorkerSettingsSchema,
    taskSettings: HivePaaSTaskSettingsSchema,
    periodicSettings: HivePaaSPeriodicSettingsSchema,
    proxySettings: HivePaaSProxySettingsSchema,
    pendingChange: SettingsPendingChangeSchema.nullish(),
});
