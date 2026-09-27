import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const SystemSslRenewalScheduleMode = {
    Interval: "interval",
    Cron: "cron",
} as const;

export type SystemSslRenewalScheduleMode =
    (typeof SystemSslRenewalScheduleMode)[keyof typeof SystemSslRenewalScheduleMode];

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

export const SystemSslRenewalConfigurationFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    scheduleMode: z.enum([SystemSslRenewalScheduleMode.Interval, SystemSslRenewalScheduleMode.Cron]),
    scheduleInterval: z.string(),
    scheduleCronExpr: z.string(),
    scheduleFrom: z.date().nullable(),
    notification: NotificationSchema,
});

export type SystemSslRenewalConfigurationFormInput = z.input<typeof SystemSslRenewalConfigurationFormSchema>;
export type SystemSslRenewalConfigurationFormOutput = z.output<typeof SystemSslRenewalConfigurationFormSchema>;
