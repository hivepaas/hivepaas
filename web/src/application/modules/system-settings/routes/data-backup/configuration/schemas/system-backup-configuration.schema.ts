import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

import { ESystemBackupCompressionFormat, ESystemBackupEncryptionFormat } from "../../../../module-shared/enums";

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
    scheduleInterval: z.string().min(1, "Run interval is required"),
    scheduleFrom: z.date({ required_error: "First run start time is required" }),
    compressionFormat: z.nativeEnum(ESystemBackupCompressionFormat),
    encryptionFormat: z.nativeEnum(ESystemBackupEncryptionFormat),
    encryptionSecret: z.string(),
    cloudStorage: SettingsRefSchema.optional(),
    cloudStorageDestinationDir: z.string(),
    backupDeletedObjects: z.boolean(),
    notification: NotificationSchema,
});

export type SystemBackupConfigurationFormInput = z.input<typeof SystemBackupConfigurationFormSchema>;
export type SystemBackupConfigurationFormOutput = z.output<typeof SystemBackupConfigurationFormSchema>;
