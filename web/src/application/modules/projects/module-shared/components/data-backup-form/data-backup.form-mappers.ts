import type { AppScheduledJobs_Command_Payload, AppScheduledJobs_Upsert_Payload } from "~/projects/api/services";
import { APP_SCHEDULED_JOB_DEFAULT_CONSOLE_SIZE, type AppScheduledJob } from "~/projects/domain";
import {
    EAppScheduledJobScheduleMode,
    EAppScheduledJobTaskPriority,
    EAppScheduledJobType,
    ESchedJobDataBackupSource,
} from "~/projects/module-shared/enums";

import { ESettingStatus } from "@application/shared/enums";

import {
    createDefaultJobScheduleFormValues,
    mapJobScheduleFormValuesToPayload,
    mapJobScheduleToFormValues,
} from "../job-schedule-fields";
import { mapJobTriggersFormValuesToPayload, mapJobTriggersToFormValues } from "../job-triggers-field";

import {
    DATA_BACKUP_COMMAND_MODE,
    type DataBackupFormInput,
    type DataBackupFormOutput,
} from "./data-backup.form.schema";

function emptySourceCommand(): DataBackupFormInput["sourceCommand"] {
    return {
        commandMode: DATA_BACKUP_COMMAND_MODE.Command,
        command: "",
        script: "",
        workingDir: "",
        tty: false,
        consoleSize: { ...APP_SCHEDULED_JOB_DEFAULT_CONSOLE_SIZE },
        envVars: [],
        argGroups: [],
    };
}

/** A new data backup reads a command's output, runs by hand until given a schedule, and can be canceled. */
export function createEmptyDataBackupFormDefaults(): DataBackupFormInput {
    return {
        name: "",
        ...createDefaultJobScheduleFormValues(EAppScheduledJobScheduleMode.None),
        source: ESchedJobDataBackupSource.Command,
        sourceCommand: emptySourceCommand(),
        sourceFileName: "",
        sourceVolume: null,
        sourceVolumeSubpath: "",
        targetRepository: null,
        tags: [],
        timeout: "",
        maxRetry: 0,
        retryDelay: "",
        priority: EAppScheduledJobTaskPriority.Default,
        controlEnabled: true,
        triggers: [],
        notification: {
            successUseDefault: true,
            success: undefined,
            failureUseDefault: true,
            failure: undefined,
        },
    };
}

/** A setting's name, or that it was deleted. */
function refName(ref: { name: string; status: string }, deleted: string): string {
    return ref.status === ESettingStatus.Missing ? deleted : ref.name;
}

export function mapDataBackupToFormInput(job: AppScheduledJob): DataBackupFormInput {
    const backup = job.dataBackup;
    const command = backup?.sourceCommand;
    const script = command?.script ?? "";

    return {
        name: job.name,
        ...mapJobScheduleToFormValues(job.schedule),
        source: backup?.source ?? ESchedJobDataBackupSource.Command,
        sourceCommand: command
            ? {
                  commandMode:
                      script.trim().length > 0 ? DATA_BACKUP_COMMAND_MODE.Script : DATA_BACKUP_COMMAND_MODE.Command,
                  command: command.command,
                  script,
                  workingDir: command.workingDir,
                  tty: false,
                  consoleSize: { ...command.consoleSize },
                  envVars: command.envVars,
                  argGroups: command.argGroups,
              }
            : emptySourceCommand(),
        sourceFileName: backup?.sourceFileName ?? "",
        sourceVolume: backup?.sourceVolume
            ? { id: backup.sourceVolume.id, name: refName(backup.sourceVolume, "Deleted volume") }
            : null,
        sourceVolumeSubpath: backup?.sourceVolumeSubpath ?? "",
        targetRepository: backup
            ? { id: backup.targetRepository.id, name: refName(backup.targetRepository, "Deleted repository") }
            : null,
        tags: Object.entries(backup?.tags ?? {}).map(([key, value]) => ({ key, value })),
        timeout: job.timeout,
        maxRetry: job.maxRetry,
        retryDelay: job.retryDelay,
        priority: job.priority,
        controlEnabled: !job.controlDisabled,
        triggers: mapJobTriggersToFormValues(job.triggers),
        notification: {
            successUseDefault: job.notification?.successUseDefault ?? true,
            success: job.notification?.success,
            failureUseDefault: job.notification?.failureUseDefault ?? true,
            failure: job.notification?.failure,
        },
    };
}

/** The command runs without a TTY: its stdout is the backup, and a TTY would mix its stderr in. */
function mapSourceCommandToPayload(cmd: DataBackupFormOutput["sourceCommand"]): AppScheduledJobs_Command_Payload {
    const isScript = cmd.commandMode === DATA_BACKUP_COMMAND_MODE.Script;
    return {
        command: isScript ? "" : cmd.command,
        script: isScript ? cmd.script : "",
        consoleSize: cmd.consoleSize,
        tty: false,
        ...(cmd.workingDir ? { workingDir: cmd.workingDir } : {}),
        ...(cmd.envVars.length > 0 ? { envVars: cmd.envVars.map(envVar => ({ ...envVar })) } : {}),
        ...(cmd.argGroups.length > 0 ? { argGroups: cmd.argGroups } : {}),
    };
}

/** A data backup is its app's: the command runs there, the volume is one it mounts. */
export function mapDataBackupFormToPayload(
    values: DataBackupFormOutput,
    appId: string,
): AppScheduledJobs_Upsert_Payload {
    const isCommand = values.source === ESchedJobDataBackupSource.Command;

    return {
        inheritable: false,
        default: false,
        name: values.name,
        jobType: EAppScheduledJobType.DataBackup,
        schedule: mapJobScheduleFormValuesToPayload(values),
        app: { id: appId },
        priority: values.priority,
        maxRetry: values.maxRetry ?? 0,
        ...(values.retryDelay ? { retryDelay: values.retryDelay } : {}),
        retryBackoff: false,
        ...(values.timeout ? { timeout: values.timeout } : {}),
        controlDisabled: !values.controlEnabled,
        dataBackup: {
            source: values.source,
            ...(isCommand
                ? {
                      sourceCommand: mapSourceCommandToPayload(values.sourceCommand),
                      sourceFileName: values.sourceFileName,
                  }
                : {
                      sourceVolume: { id: values.sourceVolume?.id ?? "" },
                      sourceVolumeSubpath: values.sourceVolumeSubpath,
                  }),
            targetRepository: { id: values.targetRepository?.id ?? "" },
            ...(values.tags.length > 0
                ? { tags: Object.fromEntries(values.tags.map(tag => [tag.key, tag.value])) }
                : {}),
        },
        triggers: mapJobTriggersFormValuesToPayload(values.triggers),
        notification: {
            successUseDefault: values.notification.successUseDefault,
            ...(!values.notification.successUseDefault && values.notification.success
                ? { success: values.notification.success }
                : {}),
            failureUseDefault: values.notification.failureUseDefault,
            ...(!values.notification.failureUseDefault && values.notification.failure
                ? { failure: values.notification.failure }
                : {}),
        },
    };
}
