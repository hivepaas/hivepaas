import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";
import { DURATION_HINT, isDuration } from "@application/shared/utils";

export const SystemBackupRepoCleanupScheduleMode = {
    Interval: "interval",
    Cron: "cron",
} as const;

export type SystemBackupRepoCleanupScheduleMode =
    (typeof SystemBackupRepoCleanupScheduleMode)[keyof typeof SystemBackupRepoCleanupScheduleMode];

const SettingsRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

const NotificationSchema = z.object({
    successUseDefault: z.boolean(),
    success: SettingsRefSchema.nullish(),
    failureUseDefault: z.boolean(),
    failure: SettingsRefSchema.nullish(),
});

export const SystemBackupRepoCleanupConfigurationFormSchema = z
    .object({
        status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
        scheduleMode: z.enum([SystemBackupRepoCleanupScheduleMode.Interval, SystemBackupRepoCleanupScheduleMode.Cron]),
        scheduleInterval: z.string(),
        scheduleCronExpr: z.string(),
        scheduleFrom: z.date().nullable(),
        notification: NotificationSchema,
    })
    .superRefine((data, ctx) => {
        if (data.scheduleMode === SystemBackupRepoCleanupScheduleMode.Interval && !isDuration(data.scheduleInterval)) {
            ctx.addIssue({ code: z.ZodIssueCode.custom, path: ["scheduleInterval"], message: DURATION_HINT });
        }
    });

export type SystemBackupRepoCleanupConfigurationFormInput = z.input<
    typeof SystemBackupRepoCleanupConfigurationFormSchema
>;
export type SystemBackupRepoCleanupConfigurationFormOutput = z.output<
    typeof SystemBackupRepoCleanupConfigurationFormSchema
>;
