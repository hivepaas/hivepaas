import type {
    EAppScheduledJobArgSeparator,
    EAppScheduledJobTaskPriority,
    EAppScheduledJobType,
    ESchedJobDataBackupSource,
    ESchedJobSeqMode,
    ESchedJobSeqOnFailure,
    ESchedJobTriggerEvent,
} from "~/projects/module-shared/enums";

import type { ESettingStatus } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

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

export interface AppScheduledJobCommandOutputSaveToFile {
    fileName: string;
    filePath: string;
    storage?: AppScheduledJobNamedRef;
    bucket?: string;
    compressionFormat: string;
    encryptionFormat: string;
    encryptionSecret?: string;
}

export interface AppScheduledJobCommandOutputPipeToApp {
    targetApp: AppScheduledJobNamedRef;
    command?: AppScheduledJobCommand;
}

export interface AppScheduledJobCommandOutput {
    enabled: boolean;
    saveToFile?: AppScheduledJobCommandOutputSaveToFile;
    pipeToApp?: AppScheduledJobCommandOutputPipeToApp;
}

/** A step of a job sequence: the job it runs, and the app that job belongs to. */
export interface AppScheduledJobSequenceStep {
    /** The job; its status is "missing" when it was deleted. */
    job: {
        id: string;
        name: string;
        kind: string;
        status: string;
    };
    app?: AppScheduledJobNamedRef;
    name: string;
}

export interface AppScheduledJobSequence {
    mode: ESchedJobSeqMode;
    onFailure: ESchedJobSeqOnFailure;
    steps: AppScheduledJobSequenceStep[];
}

/** An event that runs the job. `apps` is empty for an app's job: its own app. */
export interface AppScheduledJobTrigger {
    event: ESchedJobTriggerEvent;
    apps: AppScheduledJobNamedRef[];
    /** The deploy waits for the run; pre-deploy only. */
    wait: boolean;
}

/** A setting a data backup names; its status is "missing" when it was deleted. */
export interface AppScheduledJobSettingRef {
    id: string;
    name: string;
    status: string;
}

/** What a data backup reads - a command's output, or a volume the app mounts - and the repository it goes to. */
export interface AppScheduledJobDataBackup {
    source: ESchedJobDataBackupSource;
    sourceCommand: AppScheduledJobCommand | null;
    sourceFileName: string;
    sourceVolume: AppScheduledJobSettingRef | null;
    sourceVolumeSubpath: string;
    targetRepository: AppScheduledJobSettingRef;
    tags: Record<string, string>;
}

export interface AppScheduledJob {
    id: string;
    type: string;
    name: string;
    kind?: string;
    status: OpenApiConstant<ESettingStatus>;
    inherited: boolean;
    inheritable: boolean;
    default: boolean;
    updateVer: number;
    createdAt: Date;
    updatedAt: Date;
    expireAt: Date | null;

    jobType: EAppScheduledJobType;
    /** Null for a job without a schedule: it runs by hand, or as a step of a sequence. */
    schedule: AppScheduledJobSchedule | null;
    app?: AppScheduledJobNamedRef;
    sequence: AppScheduledJobSequence | null;
    triggers: AppScheduledJobTrigger[];
    priority: EAppScheduledJobTaskPriority;
    maxRetry: number;
    retryDelay: string;
    retryDelayIncr: string;
    retryBackoff: boolean;
    retryDelayMax: string;
    timeout: string;
    controlDisabled: boolean;
    command: AppScheduledJobCommand | null;
    commandOutput?: AppScheduledJobCommandOutput | null;
    dataBackup: AppScheduledJobDataBackup | null;
    notification: AppScheduledJobNotification | null;
    nextRuns: Date[];
}

/** A job in an env's list: the env's own, or one of an app in the env. */
export interface EnvScheduledJob extends AppScheduledJob {
    scope: "project-env" | "app";
    ownerApp?: AppScheduledJobNamedRef;
}
