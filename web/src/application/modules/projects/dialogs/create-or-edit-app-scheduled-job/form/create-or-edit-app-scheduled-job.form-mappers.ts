import type { AppScheduledJob } from "~/projects/domain";
import {
    EAppScheduledJobArgSeparator,
    EAppScheduledJobScheduleMode,
    EAppScheduledJobTaskPriority,
} from "~/projects/module-shared/enums";

import type { CreateOrEditAppScheduledJobFormInput } from "../schemas";

type ArgGroupFormInput = CreateOrEditAppScheduledJobFormInput["argGroups"][number];
type ArgFormInput = ArgGroupFormInput["args"][number];

export function createDefaultArgGroup(index: number): ArgGroupFormInput {
    return {
        enabled: true,
        exportEnv: `CMD_ARG_GROUP_${index + 1}`,
        separator: EAppScheduledJobArgSeparator.Equal,
        args: [],
    };
}

export function createDefaultArg(): ArgFormInput {
    return {
        use: true,
        name: "",
        value: "",
    };
}

export function createEmptyAppScheduledJobFormDefaults(): CreateOrEditAppScheduledJobFormInput {
    return {
        name: "",
        scheduleMode: EAppScheduledJobScheduleMode.Interval,
        scheduleInterval: "",
        scheduleCronExpr: "",
        scheduleFrom: null,
        timeout: "",
        maxRetry: undefined,
        retryDelay: "",
        priority: EAppScheduledJobTaskPriority.Default,
        controlEnabled: false,
        runInShell: "",
        command: "",
        workingDir: "",
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

    return {
        name: job.name,
        scheduleMode: hasInterval ? EAppScheduledJobScheduleMode.Interval : EAppScheduledJobScheduleMode.Cron,
        scheduleInterval: job.schedule.interval,
        scheduleCronExpr: job.schedule.cronExpr,
        scheduleFrom: job.schedule.initialTime,
        timeout: job.timeout,
        maxRetry: job.maxRetry,
        retryDelay: job.retryDelay,
        priority: job.priority,
        controlEnabled: !job.controlDisabled,
        runInShell: command?.runInShell ?? "",
        command: command?.command ?? "",
        workingDir: command?.workingDir ?? "",
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
