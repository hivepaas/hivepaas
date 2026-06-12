import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { ESettingType } from "@application/shared/enums";

const SystemSslRenewalScheduleSchema = z
    .object({
        cronExpr: z.string().nullish(),
        CronExpr: z.string().nullish(),
        interval: z.string().nullish(),
        initialTime: z.coerce.date().nullish(),
    })
    .transform(({ cronExpr, CronExpr, interval, initialTime }) => ({
        cronExpr: cronExpr ?? CronExpr ?? "",
        interval: interval ?? "",
        initialTime: initialTime ?? null,
    }));

const SystemSslRenewalNotificationRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

const SystemSslRenewalNotificationSchema = z
    .object({
        success: SystemSslRenewalNotificationRefSchema.nullish().transform(value => value ?? undefined),
        successUseDefault: z.boolean(),
        failure: SystemSslRenewalNotificationRefSchema.nullish().transform(value => value ?? undefined),
        failureUseDefault: z.boolean(),
    })
    .nullish()
    .transform(value => value ?? null);

const SystemSslRenewalSettingsBaseSchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.literal(ESettingType.SSLRenewal),
});

const SystemSslRenewalSettingsSchema = SystemSslRenewalSettingsBaseSchema.extend({
    schedule: SystemSslRenewalScheduleSchema,
    notification: SystemSslRenewalNotificationSchema,
    nextRuns: z
        .array(z.coerce.date())
        .nullish()
        .transform(value => value ?? []),
});

const LegacySystemSslRenewalSettingsSchema = SystemSslRenewalSettingsBaseSchema.extend({
    scheduleInterval: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    scheduleFrom: z.coerce
        .date()
        .nullish()
        .transform(value => value ?? null),
    notification: SystemSslRenewalNotificationSchema.optional(),
    nextRuns: z
        .array(z.coerce.date())
        .nullish()
        .transform(value => value ?? []),
}).transform(({ scheduleInterval, scheduleFrom, notification, nextRuns, ...settings }) => ({
    ...settings,
    schedule: {
        interval: scheduleInterval,
        cronExpr: "",
        initialTime: scheduleFrom,
    },
    notification: notification ?? null,
    nextRuns,
}));

export const SystemSslRenewalSettingsEntitySchema = z.union([
    SystemSslRenewalSettingsSchema,
    LegacySystemSslRenewalSettingsSchema,
]);
