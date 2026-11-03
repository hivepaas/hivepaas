import { z } from "zod";
import { JobTriggersFormSchema } from "~/projects/module-shared/components/job-triggers-field/job-triggers.helpers";
import {
    EAppScheduledJobArgSeparator,
    EAppScheduledJobScheduleMode,
    EAppScheduledJobTaskPriority,
    ESchedJobDataBackupSource,
} from "~/projects/module-shared/enums";

/** The most tags a data backup puts on its snapshots, as the server allows. */
export const DATA_BACKUP_MAX_TAGS = 20;

/** A file name without a directory: the snapshot's one file. */
const FILE_NAME_PATTERN = /^[A-Za-z0-9._-]{1,100}$/;
/** A tag is kopia's key:value; the key holds no ':', neither holds a space. */
const TAG_KEY_PATTERN = /^[A-Za-z0-9._-]{1,50}$/;
const TAG_VALUE_PATTERN = /^[A-Za-z0-9._:/@+=-]{1,100}$/;
/** HivePaaS's own tags: the job, the app, the run. */
const RESERVED_TAG_PREFIX = "hivepaas.";

export const DATA_BACKUP_COMMAND_MODE = {
    Command: "command",
    Script: "script",
} as const;

const NamedRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

/** The command section's fields, as `CommandConfigSection` edits them under `sourceCommand` or `restoreCommand`. */
export const SourceCommandSchema = z.object({
    commandMode: z.enum([DATA_BACKUP_COMMAND_MODE.Command, DATA_BACKUP_COMMAND_MODE.Script]),
    command: z.string().trim(),
    script: z.string(),
    workingDir: z.string().trim(),
    tty: z.boolean(),
    consoleSize: z.object({ width: z.number(), height: z.number() }),
    envVars: z.array(z.object({ key: z.string(), value: z.string(), isLiteral: z.boolean().optional() })),
    argGroups: z.array(
        z.object({
            enabled: z.boolean(),
            exportEnv: z.string(),
            separator: z.nativeEnum(EAppScheduledJobArgSeparator),
            args: z.array(z.object({ use: z.boolean(), name: z.string(), value: z.string() })),
        }),
    ),
});

/** A path inside a directory: not from the root, never above where it starts. */
export function isRelativeSubpath(subpath: string): boolean {
    if (subpath === "") {
        return true;
    }
    if (subpath.startsWith("/") || subpath.length > 500) {
        return false;
    }
    let depth = 0;
    for (const part of subpath.split("/")) {
        if (part === "..") {
            depth -= 1;
        } else if (part !== "" && part !== ".") {
            depth += 1;
        }
        if (depth < 0) {
            return false;
        }
    }
    return true;
}

export const DataBackupFormSchema = z
    .object({
        name: z.string().trim().min(1, "Name is required"),
        scheduleMode: z.nativeEnum(EAppScheduledJobScheduleMode),
        scheduleInterval: z.string().trim(),
        scheduleCronExpr: z.string().trim(),
        scheduleFrom: z.date().nullable(),
        scheduleTo: z.date().nullable(),
        source: z.nativeEnum(ESchedJobDataBackupSource),
        sourceCommand: SourceCommandSchema,
        sourceFileName: z.string().trim(),
        /** Optional: empty when both the command and the script are. */
        restoreCommand: SourceCommandSchema,
        sourceVolume: NamedRefSchema.nullable(),
        sourceVolumeSubpath: z.string().trim(),
        targetRepository: NamedRefSchema.nullable(),
        tags: z.array(z.object({ key: z.string().trim(), value: z.string().trim() })),
        timeout: z.string().trim(),
        maxRetry: z.number().int().min(0, "Max retry must be greater than or equal to 0").optional(),
        retryDelay: z.string().trim(),
        priority: z.nativeEnum(EAppScheduledJobTaskPriority),
        controlEnabled: z.boolean(),
        triggers: JobTriggersFormSchema,
        notification: z.object({
            successUseDefault: z.boolean(),
            success: NamedRefSchema.nullish(),
            failureUseDefault: z.boolean(),
            failure: NamedRefSchema.nullish(),
        }),
    })
    .superRefine((value, ctx) => {
        const issue = (message: string, path: (string | number)[]) => {
            ctx.addIssue({ code: z.ZodIssueCode.custom, message, path });
        };

        if (value.source === ESchedJobDataBackupSource.Command) {
            const cmd = value.sourceCommand;
            if (cmd.commandMode === DATA_BACKUP_COMMAND_MODE.Command && !cmd.command) {
                issue("Command is required", ["sourceCommand", "command"]);
            }
            if (cmd.commandMode === DATA_BACKUP_COMMAND_MODE.Script && !cmd.script.trim()) {
                issue("Script is required", ["sourceCommand", "script"]);
            }
            if (!FILE_NAME_PATTERN.test(value.sourceFileName)) {
                issue("Letters, digits, '.', '-' and '_', up to 100, no '/'", ["sourceFileName"]);
            }
        } else {
            if (!value.sourceVolume) {
                issue("Pick a volume", ["sourceVolume"]);
            }
            if (!isRelativeSubpath(value.sourceVolumeSubpath)) {
                issue("A path inside the volume: no leading '/', never above it with '..'", ["sourceVolumeSubpath"]);
            }
        }

        if (!value.targetRepository) {
            issue("Pick a repository", ["targetRepository"]);
        }

        if (value.tags.length > DATA_BACKUP_MAX_TAGS) {
            issue(`At most ${DATA_BACKUP_MAX_TAGS} tags`, ["tags"]);
        }
        value.tags.forEach((tag, index) => {
            if (!TAG_KEY_PATTERN.test(tag.key) || tag.key.startsWith(RESERVED_TAG_PREFIX)) {
                issue("Letters, digits, '.', '-' and '_'; not starting with 'hivepaas.'", ["tags", index, "key"]);
            }
            if (!TAG_VALUE_PATTERN.test(tag.value)) {
                issue("Up to 100 characters, no spaces", ["tags", index, "value"]);
            }
        });
    });

export type DataBackupFormInput = z.input<typeof DataBackupFormSchema>;
export type CommandFormInput = z.input<typeof SourceCommandSchema>;
export type DataBackupFormOutput = z.output<typeof DataBackupFormSchema>;
