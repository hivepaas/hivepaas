import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

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
});

export const HivePaaSServiceSettingsEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.string(),
    appSettings: HivePaaSAppSettingsSchema,
    workerSettings: HivePaaSWorkerSettingsSchema,
    taskSettings: HivePaaSTaskSettingsSchema,
    periodicSettings: HivePaaSPeriodicSettingsSchema,
    proxySettings: HivePaaSProxySettingsSchema,
});
