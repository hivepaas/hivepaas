import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

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
    scheduleInterval: z.string().min(1, "Run interval is required"),
    scheduleFrom: z.date({ required_error: "First run start time is required" }),
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
    notification: NotificationSchema,
});

export type SystemCleanupConfigurationFormInput = z.input<typeof SystemCleanupConfigurationFormSchema>;
export type SystemCleanupConfigurationFormOutput = z.output<typeof SystemCleanupConfigurationFormSchema>;
