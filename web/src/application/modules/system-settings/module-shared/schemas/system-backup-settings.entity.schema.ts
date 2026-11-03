import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";
import { SYSTEM_BACKUP_SPEC_SECRETS } from "~/system-settings/domain";

const SystemBackupScheduleSchema = z
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
    }))
    .default({ cronExpr: "", interval: "", initialTime: null });

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
    type: z.string(),
    schedule: SystemBackupScheduleSchema,
    includeDB: z.boolean().optional().default(false),
    includeSpec: z.boolean().optional().default(false),
    specSecrets: z.nativeEnum(SYSTEM_BACKUP_SPEC_SECRETS).catch(SYSTEM_BACKUP_SPEC_SECRETS.Encrypted),
    specPassphrase: z.string().optional().default(""),
    targetRepository: z
        .object({
            id: z.string(),
            name: z.string().optional().default(""),
            status: z.string().optional().default(""),
        })
        .nullish()
        .transform(value => value ?? null),
    notification: SystemBackupNotificationSchema,
    secretMasked: z.boolean().optional(),
    nextRuns: z
        .array(z.coerce.date())
        .nullish()
        .transform(value => value ?? []),
});
