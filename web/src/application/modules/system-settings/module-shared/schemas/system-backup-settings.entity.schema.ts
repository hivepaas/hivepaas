import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { ESettingType } from "@application/shared/enums";

import { ESystemBackupCompressionFormat, ESystemBackupEncryptionFormat } from "../enums";

const SystemBackupCompressionSchema = z
    .object({
        format: z.nativeEnum(ESystemBackupCompressionFormat).default(ESystemBackupCompressionFormat.None),
    })
    .default({ format: ESystemBackupCompressionFormat.None });

const SystemBackupEncryptionSchema = z
    .object({
        format: z.nativeEnum(ESystemBackupEncryptionFormat).default(ESystemBackupEncryptionFormat.None),
        secret: z.string().default(""),
    })
    .default({ format: ESystemBackupEncryptionFormat.None, secret: "" });

const SystemBackupCloudStorageSchema = SettingsBaseEntitySchema.omit({ description: true })
    .extend({
        destinationDir: z.string().default(""),
    })
    .nullable();

const SystemBackupDBConfigSchema = z.object({
    backupDeletedObjects: z.boolean(),
});

const SystemBackupNotificationRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

const SystemBackupNotificationSchema = z
    .object({
        success: SystemBackupNotificationRefSchema.nullish().transform(value => value ?? undefined),
        successUseDefault: z.boolean(),
        failure: SystemBackupNotificationRefSchema.nullish().transform(value => value ?? undefined),
        failureUseDefault: z.boolean(),
    })
    .nullable();

export const SystemBackupSettingsEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.literal(ESettingType.SystemBackup),
    scheduleInterval: z.string(),
    scheduleFrom: z.coerce.date(),
    compression: SystemBackupCompressionSchema,
    encryption: SystemBackupEncryptionSchema,
    cloudStorage: SystemBackupCloudStorageSchema,
    dbBackupConfig: SystemBackupDBConfigSchema,
    notification: SystemBackupNotificationSchema,
});
