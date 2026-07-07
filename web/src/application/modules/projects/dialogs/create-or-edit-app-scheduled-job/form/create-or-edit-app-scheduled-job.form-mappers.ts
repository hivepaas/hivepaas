import { APP_SCHEDULED_JOB_DEFAULT_CONSOLE_SIZE, type AppScheduledJob } from "~/projects/domain";
import {
    type CommandArgFormValue,
    type CommandArgGroupFormValue,
    createDefaultCommandArg,
    createDefaultCommandArgGroup,
} from "~/projects/module-shared/components";
import { EAppScheduledJobScheduleMode, EAppScheduledJobTaskPriority } from "~/projects/module-shared/enums";

import type { CreateOrEditAppScheduledJobFormInput } from "../schemas";
import { APP_SCHEDULED_JOB_COMMAND_MODE } from "../schemas";

export function createDefaultArgGroup(index: number): CommandArgGroupFormValue {
    return createDefaultCommandArgGroup(index);
}

export function createDefaultArg(): CommandArgFormValue {
    return createDefaultCommandArg();
}

export function createEmptyAppScheduledJobFormDefaults(): CreateOrEditAppScheduledJobFormInput {
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
        notification: {
            successUseDefault: true,
            success: undefined,
            failureUseDefault: true,
            failure: undefined,
        },
    };
}

export function mapAppScheduledJobToFormInput(job: AppScheduledJob): CreateOrEditAppScheduledJobFormInput {
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
        notification: {
            successUseDefault: job.notification?.successUseDefault ?? true,
            success: job.notification?.success,
            failureUseDefault: job.notification?.failureUseDefault ?? true,
            failure: job.notification?.failure,
        },
    };
}
