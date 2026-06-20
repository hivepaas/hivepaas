import { z } from "zod";
import {
    EAppScheduledJobArgSeparator,
    EAppScheduledJobScheduleMode,
    EAppScheduledJobTaskPriority,
} from "~/projects/module-shared/enums";

export const APP_SCHEDULED_JOB_COMMAND_MODE = {
    Command: "command",
    Script: "script",
} as const;

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

const ConsoleSizeSchema = z.object({
    width: z.number().int("Width must be an integer").min(1, "Width must be greater than 0"),
    height: z.number().int("Height must be an integer").min(1, "Height must be greater than 0"),
});

export const CreateOrEditAppScheduledJobFormSchema = z
    .object({
        name: z.string().trim().min(1, "Name is required"),
        scheduleMode: z.nativeEnum(EAppScheduledJobScheduleMode),
        scheduleInterval: z.string().trim(),
        scheduleCronExpr: z.string().trim(),
        scheduleFrom: z.date().nullable(),
        timeout: z.string().trim(),
        maxRetry: z.number().int().min(0, "Max retry must be greater than or equal to 0").optional(),
        retryDelay: z.string().trim(),
        priority: z.nativeEnum(EAppScheduledJobTaskPriority),
        controlEnabled: z.boolean(),
        runInShell: z.string().trim(),
        commandMode: z.enum([APP_SCHEDULED_JOB_COMMAND_MODE.Command, APP_SCHEDULED_JOB_COMMAND_MODE.Script]),
        command: z.string().trim(),
        script: z.string(),
        workingDir: z.string().trim(),
        tty: z.boolean(),
        consoleSize: ConsoleSizeSchema,
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
        if (value.commandMode === APP_SCHEDULED_JOB_COMMAND_MODE.Command && !value.command.trim()) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Command is required",
                path: ["command"],
            });
        }

        if (value.commandMode === APP_SCHEDULED_JOB_COMMAND_MODE.Script && !value.script.trim()) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Script is required",
                path: ["script"],
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
