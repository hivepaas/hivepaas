import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

const SystemBackupRepoCleanupScheduleSchema = z
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

const SystemBackupRepoCleanupNotificationRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

const SystemBackupRepoCleanupNotificationSchema = z
    .object({
        success: SystemBackupRepoCleanupNotificationRefSchema.nullish().transform(value => value ?? undefined),
        successUseDefault: z.boolean(),
        failure: SystemBackupRepoCleanupNotificationRefSchema.nullish().transform(value => value ?? undefined),
        failureUseDefault: z.boolean(),
    })
    .nullish()
    .transform(value => value ?? null);

const SystemBackupRepoCleanupSettingsBaseSchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.string(),
});

const SystemBackupRepoCleanupSettingsSchema = SystemBackupRepoCleanupSettingsBaseSchema.extend({
    schedule: SystemBackupRepoCleanupScheduleSchema,
    notification: SystemBackupRepoCleanupNotificationSchema,
    nextRuns: z
        .array(z.coerce.date())
        .nullish()
        .transform(value => value ?? []),
});

const LegacySystemBackupRepoCleanupSettingsSchema = SystemBackupRepoCleanupSettingsBaseSchema.extend({
    scheduleInterval: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    scheduleFrom: z.coerce
        .date()
        .nullish()
        .transform(value => value ?? null),
    notification: SystemBackupRepoCleanupNotificationSchema.optional(),
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

export const SystemBackupRepoCleanupSettingsEntitySchema = z.union([
    SystemBackupRepoCleanupSettingsSchema,
    LegacySystemBackupRepoCleanupSettingsSchema,
]);
