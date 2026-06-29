import type {
    EAppScheduledJobArgSeparator,
    EAppScheduledJobTaskPriority,
    EAppScheduledJobType,
} from "~/projects/module-shared/enums";

import type { ESettingStatus, ESettingType } from "@application/shared/enums";

export interface AppScheduledJobNamedRef {
    id: string;
    name: string;
}

export interface AppScheduledJobSchedule {
    cronExpr: string;
    interval: string;
    initialTime: Date;
    endTime: Date | null;
}

export interface AppScheduledJobCommandEnvVar {
    key: string;
    value: string;
    isLiteral: boolean;
}

export interface AppScheduledJobCommandArg {
    use: boolean;
    name: string;
    value: string;
}

export interface AppScheduledJobCommandArgGroup {
    enabled: boolean;
    exportEnv: string;
    separator: EAppScheduledJobArgSeparator;
    args: AppScheduledJobCommandArg[];
}

export interface AppScheduledJobCommandConsoleSize {
    width: number;
    height: number;
}

export const APP_SCHEDULED_JOB_DEFAULT_CONSOLE_SIZE = {
    width: 120,
    height: 40,
} as const satisfies AppScheduledJobCommandConsoleSize;

export interface AppScheduledJobCommand {
    runInShell: string;
    command: string;
    script: string;
    workingDir: string;
    consoleSize: AppScheduledJobCommandConsoleSize;
    tty: boolean;
    envVars: AppScheduledJobCommandEnvVar[];
    argGroups: AppScheduledJobCommandArgGroup[];
}

export interface AppScheduledJobNotification {
    successUseDefault: boolean;
    success?: AppScheduledJobNamedRef;
    failureUseDefault: boolean;
    failure?: AppScheduledJobNamedRef;
}

export interface AppScheduledJob {
    id: string;
    type: ESettingType;
    name: string;
    kind?: string;
    status: ESettingStatus;
    inherited: boolean;
    availableInProjects: boolean;
    default: boolean;
    updateVer: number;
    createdAt: Date;
    updatedAt: Date;
    expireAt: Date | null;

    jobType: EAppScheduledJobType;
    schedule: AppScheduledJobSchedule;
    app?: AppScheduledJobNamedRef;
    priority: EAppScheduledJobTaskPriority;
    maxRetry: number;
    retryDelay: string;
    retryDelayIncr: string;
    retryBackoff: boolean;
    retryDelayMax: string;
    timeout: string;
    controlDisabled: boolean;
    command: AppScheduledJobCommand | null;
    notification: AppScheduledJobNotification | null;
    nextRuns: Date[];
}
