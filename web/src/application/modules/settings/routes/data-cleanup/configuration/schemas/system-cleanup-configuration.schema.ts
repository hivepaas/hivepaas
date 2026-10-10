import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";
import { DURATION_HINT, isDuration } from "@application/shared/utils";

export const SystemCleanupScheduleMode = {
    Interval: "interval",
    Cron: "cron",
} as const;

export type SystemCleanupScheduleMode = (typeof SystemCleanupScheduleMode)[keyof typeof SystemCleanupScheduleMode];

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

// A retention left empty is none; one filled in must read as a duration.
const RetentionSchema = z.string().refine(value => value.trim() === "" || isDuration(value), DURATION_HINT);

export const SystemCleanupConfigurationFormSchema = z
    .object({
        status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
        scheduleMode: z.enum([SystemCleanupScheduleMode.Interval, SystemCleanupScheduleMode.Cron]),
        scheduleInterval: z.string(),
        scheduleCronExpr: z.string(),
        scheduleFrom: z.date().nullable(),
        dbObjectRetention: z.object({
            enabled: z.boolean(),
            tasks: RetentionSchema,
            deployments: RetentionSchema,
            sysErrors: RetentionSchema,
            auditLogs: RetentionSchema,
            deletedObjects: RetentionSchema,
        }),
        clusterCleanup: z.object({
            enabled: z.boolean(),
            generalRetention: RetentionSchema,
            buildCacheRetention: RetentionSchema,
            pruneImages: z.boolean(),
            pruneVolumes: z.boolean(),
            pruneNetworks: z.boolean(),
            pruneContainers: z.boolean(),
            pruneBuildCache: z.boolean(),
        }),
        cacheCleanup: z.object({
            enabled: z.boolean(),
            repoCacheRetention: RetentionSchema,
        }),
        fileCleanup: z.object({
            enabled: z.boolean(),
        }),
        systemAppsSync: z.object({
            enabled: z.boolean(),
        }),
        notification: NotificationSchema,
    })
    .superRefine((data, ctx) => {
        if (data.scheduleMode === SystemCleanupScheduleMode.Interval && !isDuration(data.scheduleInterval)) {
            ctx.addIssue({ code: z.ZodIssueCode.custom, path: ["scheduleInterval"], message: DURATION_HINT });
        }
    });

export type SystemCleanupConfigurationFormInput = z.input<typeof SystemCleanupConfigurationFormSchema>;
export type SystemCleanupConfigurationFormOutput = z.output<typeof SystemCleanupConfigurationFormSchema>;
