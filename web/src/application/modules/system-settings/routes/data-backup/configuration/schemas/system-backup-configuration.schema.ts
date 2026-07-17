import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

import { ESystemBackupCompressionFormat, ESystemBackupEncryptionFormat } from "../../../../module-shared/enums";

export const SystemBackupScheduleMode = {
    Interval: "interval",
    Cron: "cron",
} as const;

export type SystemBackupScheduleMode = (typeof SystemBackupScheduleMode)[keyof typeof SystemBackupScheduleMode];

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

export const SystemBackupConfigurationFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    scheduleMode: z.enum([SystemBackupScheduleMode.Interval, SystemBackupScheduleMode.Cron]),
    scheduleInterval: z.string(),
    scheduleCronExpr: z.string(),
    scheduleFrom: z.date().nullable(),
    compressionFormat: z.nativeEnum(ESystemBackupCompressionFormat),
    encryptionFormat: z.nativeEnum(ESystemBackupEncryptionFormat),
    encryptionSecret: z.string(),
    cloudStorage: SettingsRefSchema.optional(),
    cloudStorageBucket: z.string(),
    cloudStorageDestinationDir: z.string(),
    backupDeletedObjects: z.boolean(),
    notification: NotificationSchema,
});

export type SystemBackupConfigurationFormInput = z.input<typeof SystemBackupConfigurationFormSchema>;
export type SystemBackupConfigurationFormOutput = z.output<typeof SystemBackupConfigurationFormSchema>;
