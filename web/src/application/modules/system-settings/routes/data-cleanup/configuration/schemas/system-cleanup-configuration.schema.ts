import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const SystemCleanupScheduleMode = {
    Interval: "interval",
    Cron: "cron",
} as const;

export type SystemCleanupScheduleMode = (typeof SystemCleanupScheduleMode)[keyof typeof SystemCleanupScheduleMode];

const SettingsRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

const NotificationSchema = z.object({
    successUseDefault: z.boolean(),
    success: SettingsRefSchema.optional(),
    failureUseDefault: z.boolean(),
    failure: SettingsRefSchema.optional(),
});

export const SystemCleanupConfigurationFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    scheduleMode: z.enum([SystemCleanupScheduleMode.Interval, SystemCleanupScheduleMode.Cron]),
    scheduleInterval: z.string(),
    scheduleCronExpr: z.string(),
    scheduleFrom: z.date().nullable(),
    dbObjectRetention: z.object({
        enabled: z.boolean(),
        tasks: z.string(),
        deployments: z.string(),
        sysErrors: z.string(),
        deletedObjects: z.string(),
    }),
    clusterCleanup: z.object({
        enabled: z.boolean(),
        pruneImages: z.boolean(),
        pruneVolumes: z.boolean(),
        pruneNetworks: z.boolean(),
        pruneContainers: z.boolean(),
    }),
    backupCleanup: z.object({
        enabled: z.boolean(),
        localBackupRetention: z.string(),
        cloudBackupRetention: z.string(),
    }),
    cacheCleanup: z.object({
        enabled: z.boolean(),
        repoCacheRetention: z.string(),
    }),
    fileCleanup: z.object({
        enabled: z.boolean(),
    }),
    notification: NotificationSchema,
});

export type SystemCleanupConfigurationFormInput = z.input<typeof SystemCleanupConfigurationFormSchema>;
export type SystemCleanupConfigurationFormOutput = z.output<typeof SystemCleanupConfigurationFormSchema>;
