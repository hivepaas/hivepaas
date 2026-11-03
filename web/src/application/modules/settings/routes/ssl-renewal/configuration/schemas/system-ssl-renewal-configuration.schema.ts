import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";
import { DURATION_HINT, isDuration } from "@application/shared/utils";

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
    success: SettingsRefSchema.nullish(),
    failureUseDefault: z.boolean(),
    failure: SettingsRefSchema.nullish(),
});

export const SystemSslRenewalConfigurationFormSchema = z
    .object({
        status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
        scheduleMode: z.enum([SystemSslRenewalScheduleMode.Interval, SystemSslRenewalScheduleMode.Cron]),
        scheduleInterval: z.string(),
        scheduleCronExpr: z.string(),
        scheduleFrom: z.date().nullable(),
        notification: NotificationSchema,
    })
    .superRefine((data, ctx) => {
        if (data.scheduleMode === SystemSslRenewalScheduleMode.Interval && !isDuration(data.scheduleInterval)) {
            ctx.addIssue({ code: z.ZodIssueCode.custom, path: ["scheduleInterval"], message: DURATION_HINT });
        }
    });

export type SystemSslRenewalConfigurationFormInput = z.input<typeof SystemSslRenewalConfigurationFormSchema>;
export type SystemSslRenewalConfigurationFormOutput = z.output<typeof SystemSslRenewalConfigurationFormSchema>;
