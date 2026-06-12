import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { ESettingType } from "@application/shared/enums";

const SystemCleanupDBObjectRetentionSchema = z.object({
    enabled: z.boolean(),
    tasks: z.string(),
    sysErrors: z.string(),
    deployments: z.string(),
    deletedObjects: z.string(),
});

const SystemCleanupClusterCleanupSchema = z.object({
    enabled: z.boolean(),
    pruneImages: z.boolean(),
    pruneVolumes: z.boolean(),
    pruneNetworks: z.boolean(),
    pruneContainers: z.boolean(),
});

const SystemCleanupBackupCleanupSchema = z.object({
    enabled: z.boolean(),
    cloudBackupRetention: z.string(),
    localBackupRetention: z.string(),
});

const SystemCleanupScheduleSchema = z
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

const SystemCleanupNotificationRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

const SystemCleanupNotificationSchema = z
    .object({
        success: SystemCleanupNotificationRefSchema.nullish().transform(value => value ?? undefined),
        successUseDefault: z.boolean(),
        failure: SystemCleanupNotificationRefSchema.nullish().transform(value => value ?? undefined),
        failureUseDefault: z.boolean(),
    })
    .nullable();

export const SystemCleanupSettingsEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.literal(ESettingType.SystemCleanup),
    schedule: SystemCleanupScheduleSchema,
    dbObjectRetention: SystemCleanupDBObjectRetentionSchema,
    clusterCleanup: SystemCleanupClusterCleanupSchema,
    backupCleanup: SystemCleanupBackupCleanupSchema,
    notification: SystemCleanupNotificationSchema,
    nextRuns: z
        .array(z.coerce.date())
        .nullish()
        .transform(value => value ?? []),
});
