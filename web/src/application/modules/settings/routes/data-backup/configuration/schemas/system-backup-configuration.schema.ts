import { z } from "zod";
import { SYSTEM_BACKUP_SPEC_SECRETS } from "~/system-settings/domain";

import { ESettingStatus } from "@application/shared/enums";
import { DURATION_HINT, isDuration } from "@application/shared/utils";

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

export const SystemBackupConfigurationFormSchema = z
    .object({
        status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
        scheduleMode: z.enum([SystemBackupScheduleMode.Interval, SystemBackupScheduleMode.Cron]),
        scheduleInterval: z.string(),
        scheduleCronExpr: z.string(),
        scheduleFrom: z.date().nullable(),
        includeDB: z.boolean(),
        includeSpec: z.boolean(),
        specSecrets: z.nativeEnum(SYSTEM_BACKUP_SPEC_SECRETS),
        specPassphrase: z.string(),
        targetRepository: SettingsRefSchema.optional(),
        notification: NotificationSchema,
    })
    .superRefine((data, ctx) => {
        const issue = (message: string, path: string) => {
            ctx.addIssue({ code: z.ZodIssueCode.custom, path: [path], message });
        };
        // Read by the server even while disabled.
        if (data.scheduleMode === SystemBackupScheduleMode.Interval && !isDuration(data.scheduleInterval)) {
            issue(DURATION_HINT, "scheduleInterval");
        }
        // A disabled backup is not asked for what it would take.
        if (data.status !== ESettingStatus.Active) {
            return;
        }
        if (!data.includeDB && !data.includeSpec) {
            issue("Back up the database, the spec, or both", "includeDB");
        }
        if (!data.targetRepository) {
            issue("Pick a backup repository", "targetRepository");
        }
        if (data.includeSpec && data.specSecrets === SYSTEM_BACKUP_SPEC_SECRETS.Encrypted) {
            if (!data.specPassphrase) {
                issue("A passphrase is required to encrypt the spec's secrets", "specPassphrase");
            } else if (data.specPassphrase.length > 256) {
                issue("At most 256 characters", "specPassphrase");
            }
        }
    });

export type SystemBackupConfigurationFormInput = z.input<typeof SystemBackupConfigurationFormSchema>;
export type SystemBackupConfigurationFormOutput = z.output<typeof SystemBackupConfigurationFormSchema>;
