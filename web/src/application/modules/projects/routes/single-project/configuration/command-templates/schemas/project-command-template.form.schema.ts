import { z } from "zod";
import { EAppScheduledJobArgSeparator } from "~/projects/module-shared/enums";

import { ECommandTemplateKind } from "@application/shared/enums";

export const PROJECT_COMMAND_TEMPLATE_COMMAND_MODE = {
    Command: "command",
    Script: "script",
} as const;

const COMMAND_TEMPLATE_MAX_SIZE = 300 * 1024;

const KeyValueSchema = z.object({
    key: z.string(),
    value: z.string(),
    isLiteral: z.boolean().optional(),
});

const ArgSchema = z.object({
    use: z.boolean(),
    name: z.string().trim(),
    value: z.string(),
});

const ArgGroupSchema = z.object({
    enabled: z.boolean(),
    exportEnv: z.string().trim(),
    separator: z.enum([EAppScheduledJobArgSeparator.Equal, EAppScheduledJobArgSeparator.Whitespace]),
    args: z.array(ArgSchema),
});

const CommandTemplateKindSchema = z.union([z.nativeEnum(ECommandTemplateKind), z.literal("")]);

export const ProjectCommandTemplateFormSchema = z
    .object({
        name: z.string().trim().min(1, "Name is required"),
        kind: CommandTemplateKindSchema,
        commandMode: z.enum([
            PROJECT_COMMAND_TEMPLATE_COMMAND_MODE.Command,
            PROJECT_COMMAND_TEMPLATE_COMMAND_MODE.Script,
        ]),
        command: z.string().trim(),
        script: z.string(),
        workingDir: z.string().trim(),
        tty: z.boolean(),
        consoleSize: z.object({
            width: z.coerce.number().int().min(1, "Width must be at least 1"),
            height: z.coerce.number().int().min(1, "Height must be at least 1"),
        }),
        envVars: z.array(KeyValueSchema),
        argGroups: z.array(ArgGroupSchema),
        default: z.boolean(),
    })
    .superRefine((value, ctx) => {
        if (!value.kind) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["kind"],
                message: "Type is required",
            });
        }

        if (value.commandMode === PROJECT_COMMAND_TEMPLATE_COMMAND_MODE.Command) {
            if (!value.command) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    path: ["command"],
                    message: "Command is required",
                });
            }

            if (value.command.length > COMMAND_TEMPLATE_MAX_SIZE) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    path: ["command"],
                    message: "Command must be at most 300kb",
                });
            }
        }

        if (value.commandMode === PROJECT_COMMAND_TEMPLATE_COMMAND_MODE.Script) {
            if (!value.script.trim()) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    path: ["script"],
                    message: "Script is required",
                });
            }

            if (value.script.length > COMMAND_TEMPLATE_MAX_SIZE) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    path: ["script"],
                    message: "Script must be at most 300kb",
                });
            }
        }

        value.argGroups.forEach((group, groupIndex) => {
            if (group.enabled && !group.exportEnv) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    path: ["argGroups", groupIndex, "exportEnv"],
                    message: "Export Env is required",
                });
            }

            group.args.forEach((arg, argIndex) => {
                if (arg.use && !arg.name) {
                    ctx.addIssue({
                        code: z.ZodIssueCode.custom,
                        path: ["argGroups", groupIndex, "args", argIndex, "name"],
                        message: "Arg name is required",
                    });
                }
            });
        });
    });

export type ProjectCommandTemplateFormInput = z.input<typeof ProjectCommandTemplateFormSchema>;
export type ProjectCommandTemplateFormOutput = z.output<typeof ProjectCommandTemplateFormSchema>;
