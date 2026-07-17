import {
    APP_SCHEDULED_JOB_DEFAULT_CONSOLE_SIZE,
    type AppScheduledJob,
    type AppScheduledJobCommandOutput,
} from "~/projects/domain";
import {
    type CommandArgFormValue,
    type CommandArgGroupFormValue,
    createDefaultCommandArg,
    createDefaultCommandArgGroup,
} from "~/projects/module-shared/components";
import {
    EAppScheduledJobCommandOutputMode,
    EAppScheduledJobScheduleMode,
    EAppScheduledJobTaskPriority,
} from "~/projects/module-shared/enums";

import type { CreateOrEditAppScheduledJobFormInput } from "../schemas";
import { APP_SCHEDULED_JOB_COMMAND_MODE } from "../schemas";

export function createDefaultArgGroup(index: number): CommandArgGroupFormValue {
    return createDefaultCommandArgGroup(index);
}

export function createDefaultArg(): CommandArgFormValue {
    return createDefaultCommandArg();
}

function createDefaultPipeCommand(): CreateOrEditAppScheduledJobFormInput["pipeCommand"] {
    return {
        commandMode: APP_SCHEDULED_JOB_COMMAND_MODE.Command,
        command: "",
        script: "",
        workingDir: "",
        tty: false,
        consoleSize: { ...APP_SCHEDULED_JOB_DEFAULT_CONSOLE_SIZE },
        envVars: [],
        argGroups: [],
    };
}

function createDefaultSaveToFile(): CreateOrEditAppScheduledJobFormInput["saveToFile"] {
    return {
        fileName: "",
        filePath: "",
        storage: null,
        bucket: "",
        compressionFormat: "",
        encryptionFormat: "",
        encryptionSecret: "",
    };
}

export function createEmptyAppScheduledJobFormDefaults(projectId: string = ""): CreateOrEditAppScheduledJobFormInput {
    return {
        name: "",
        scheduleMode: EAppScheduledJobScheduleMode.Interval,
        scheduleInterval: "",
        scheduleCronExpr: "",
        scheduleFrom: null,
        scheduleTo: null,
        timeout: "",
        maxRetry: 0,
        retryDelay: "",
        retryDelayIncr: "",
        retryBackoff: false,
        retryDelayMax: "",
        priority: EAppScheduledJobTaskPriority.Default,
        controlEnabled: false,
        runInShell: "",
        commandMode: APP_SCHEDULED_JOB_COMMAND_MODE.Command,
        command: "",
        script: "",
        workingDir: "",
        tty: false,
        consoleSize: { ...APP_SCHEDULED_JOB_DEFAULT_CONSOLE_SIZE },
        envVars: [],
        argGroups: [],
        commandOutputMode: EAppScheduledJobCommandOutputMode.Ignore,
        saveToFile: createDefaultSaveToFile(),
        pipeTargetProject: { id: projectId, name: "" },
        pipeTargetApp: null,
        pipeCommand: createDefaultPipeCommand(),
        notification: {
            successUseDefault: true,
            success: undefined,
            failureUseDefault: true,
            failure: undefined,
        },
    };
}

function mapCommandOutput(
    commandOutput: AppScheduledJobCommandOutput | null | undefined,
    projectId: string,
): Pick<
    CreateOrEditAppScheduledJobFormInput,
    "commandOutputMode" | "saveToFile" | "pipeTargetProject" | "pipeTargetApp" | "pipeCommand"
> {
    if (!commandOutput?.enabled) {
        return {
            commandOutputMode: EAppScheduledJobCommandOutputMode.Ignore,
            saveToFile: createDefaultSaveToFile(),
            pipeTargetProject: { id: projectId, name: "" },
            pipeTargetApp: null,
            pipeCommand: createDefaultPipeCommand(),
        };
    }

    if (commandOutput.saveToFile) {
        const stf = commandOutput.saveToFile;

        return {
            commandOutputMode: EAppScheduledJobCommandOutputMode.SaveToFile,
            saveToFile: {
                fileName: stf.fileName,
                filePath: stf.filePath,
                storage: stf.storage ? { id: stf.storage.id, name: stf.storage.name } : null,
                bucket: stf.bucket ?? "",
                compressionFormat: stf.compressionFormat,
                encryptionFormat: stf.encryptionFormat,
                encryptionSecret: stf.encryptionSecret ?? "",
            },
            pipeTargetProject: { id: projectId, name: "" },
            pipeTargetApp: null,
            pipeCommand: createDefaultPipeCommand(),
        };
    }

    if (commandOutput.pipeToApp) {
        const pta = commandOutput.pipeToApp;
        const pipeCmd = pta.command;
        const isScript = (pipeCmd?.script.trim() ?? "").length > 0;

        return {
            commandOutputMode: EAppScheduledJobCommandOutputMode.PipeToApp,
            saveToFile: createDefaultSaveToFile(),
            pipeTargetProject: { id: projectId, name: "" },
            pipeTargetApp: { id: pta.targetApp.id, name: pta.targetApp.name },
            pipeCommand: {
                commandMode: isScript ? APP_SCHEDULED_JOB_COMMAND_MODE.Script : APP_SCHEDULED_JOB_COMMAND_MODE.Command,
                command: isScript ? "" : (pipeCmd?.command ?? ""),
                script: isScript ? (pipeCmd?.script ?? "") : "",
                workingDir: pipeCmd?.workingDir ?? "",
                tty: pipeCmd?.tty ?? false,
                consoleSize: { ...(pipeCmd?.consoleSize ?? APP_SCHEDULED_JOB_DEFAULT_CONSOLE_SIZE) },
                envVars: pipeCmd?.envVars ?? [],
                argGroups: pipeCmd?.argGroups ?? [],
            },
        };
    }

    return {
        commandOutputMode: EAppScheduledJobCommandOutputMode.Ignore,
        saveToFile: createDefaultSaveToFile(),
        pipeTargetProject: { id: projectId, name: "" },
        pipeTargetApp: null,
        pipeCommand: createDefaultPipeCommand(),
    };
}

export function mapAppScheduledJobToFormInput(
    job: AppScheduledJob,
    projectId: string = "",
): CreateOrEditAppScheduledJobFormInput {
    const hasInterval = job.schedule.interval.trim().length > 0;
    const { command } = job;
    const script = command?.script ?? "";

    return {
        name: job.name,
        scheduleMode: hasInterval ? EAppScheduledJobScheduleMode.Interval : EAppScheduledJobScheduleMode.Cron,
        scheduleInterval: job.schedule.interval,
        scheduleCronExpr: job.schedule.cronExpr,
        scheduleFrom: job.schedule.initialTime,
        scheduleTo: job.schedule.endTime,
        timeout: job.timeout,
        maxRetry: job.maxRetry,
        retryDelay: job.retryDelay,
        retryDelayIncr: job.retryDelayIncr,
        retryBackoff: job.retryBackoff,
        retryDelayMax: job.retryDelayMax,
        priority: job.priority,
        controlEnabled: !job.controlDisabled,
        runInShell: command?.runInShell ?? "",
        commandMode:
            script.trim().length > 0 ? APP_SCHEDULED_JOB_COMMAND_MODE.Script : APP_SCHEDULED_JOB_COMMAND_MODE.Command,
        command: command?.command ?? "",
        script,
        workingDir: command?.workingDir ?? "",
        tty: command?.tty ?? false,
        consoleSize: { ...(command?.consoleSize ?? APP_SCHEDULED_JOB_DEFAULT_CONSOLE_SIZE) },
        envVars: command?.envVars ?? [],
        argGroups: command?.argGroups ?? [],
        ...mapCommandOutput(job.commandOutput, projectId),
        notification: {
            successUseDefault: job.notification?.successUseDefault ?? true,
            success: job.notification?.success,
            failureUseDefault: job.notification?.failureUseDefault ?? true,
            failure: job.notification?.failure,
        },
    };
}
