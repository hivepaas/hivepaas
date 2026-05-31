import { z } from "zod";
import {
    EAppScheduledJobArgSeparator,
    EAppScheduledJobScheduleMode,
    EAppScheduledJobTaskPriority,
} from "~/projects/module-shared/enums";

const NotificationRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

const EnvVarSchema = z.object({
    key: z.string(),
    value: z.string(),
    isLiteral: z.boolean().optional(),
});

const CommandArgSchema = z.object({
    use: z.boolean(),
    name: z.string(),
    value: z.string(),
});

const CommandArgGroupSchema = z.object({
    enabled: z.boolean(),
    exportEnv: z.string(),
    separator: z.nativeEnum(EAppScheduledJobArgSeparator),
    args: z.array(CommandArgSchema),
});

export const CreateOrEditAppScheduledJobFormSchema = z
    .object({
        name: z.string().trim().min(1, "Name is required"),
        scheduleMode: z.nativeEnum(EAppScheduledJobScheduleMode),
        scheduleInterval: z.string().trim(),
        scheduleCronExpr: z.string().trim(),
        scheduleFrom: z.date().nullable(),
        timeout: z.string().trim().min(1, "Timeout is required"),
        maxRetry: z.number().int().min(0, "Max retry must be greater than or equal to 0"),
        retryDelay: z.string().trim().min(1, "Retry delay is required"),
        priority: z.nativeEnum(EAppScheduledJobTaskPriority),
        controlEnabled: z.boolean(),
        runInShell: z.string().trim(),
        command: z.string().trim().min(1, "Command is required"),
        workingDir: z.string().trim(),
        envVars: z.array(EnvVarSchema),
        argGroups: z.array(CommandArgGroupSchema),
        notification: z.object({
            successUseDefault: z.boolean(),
            success: NotificationRefSchema.optional(),
            failureUseDefault: z.boolean(),
            failure: NotificationRefSchema.optional(),
        }),
    })
    .superRefine((value, ctx) => {
        if (value.scheduleMode === EAppScheduledJobScheduleMode.Interval && !value.scheduleInterval) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Scheduling interval is required",
                path: ["scheduleInterval"],
            });
        }

        if (value.scheduleMode === EAppScheduledJobScheduleMode.Cron && !value.scheduleCronExpr) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Cron expression is required",
                path: ["scheduleCronExpr"],
            });
        }

        value.argGroups.forEach((group, groupIndex) => {
            if (!group.enabled) {
                return;
            }

            if (!group.exportEnv.trim()) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    message: "Export env is required",
                    path: ["argGroups", groupIndex, "exportEnv"],
                });
            }

            group.args.forEach((arg, argIndex) => {
                if (arg.use && !arg.name.trim()) {
                    ctx.addIssue({
                        code: z.ZodIssueCode.custom,
                        message: "Arg name is required",
                        path: ["argGroups", groupIndex, "args", argIndex, "name"],
                    });
                }
            });
        });
    });

export type CreateOrEditAppScheduledJobFormInput = z.input<typeof CreateOrEditAppScheduledJobFormSchema>;
export type CreateOrEditAppScheduledJobFormOutput = z.output<typeof CreateOrEditAppScheduledJobFormSchema>;
