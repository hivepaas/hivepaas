export const SystemTaskStatus = {
    NotStarted: "not-started",
    InProgress: "in-progress",
    Canceled: "canceled",
    Done: "done",
    Failed: "failed",
} as const;

export type SystemTaskStatus = (typeof SystemTaskStatus)[keyof typeof SystemTaskStatus];

export const SystemTaskPriority = {
    Low: "low",
    Default: "default",
    Critical: "critical",
} as const;

export type SystemTaskPriority = (typeof SystemTaskPriority)[keyof typeof SystemTaskPriority];

export const SystemTaskJobName = {
    DataBackup: "data-backup",
    DataCleanup: "data-cleanup",
    SslRenewal: "ssl-renewal",
    BackupRepoCleanup: "backup-repo-cleanup",
} as const;

export type SystemTaskJobName = (typeof SystemTaskJobName)[keyof typeof SystemTaskJobName];

export interface SystemTaskConfig {
    priority: SystemTaskPriority;
    maxRetry: number;
    retry: number;
    retryDelay: string;
    timeout: string;
    controlDisabled: boolean;
}

export interface SystemTaskTargetJob {
    id: string;
    type: string;
    kind: string;
    name: string;
    status: string;
}

export interface SystemTaskScopeProject {
    id: string;
    name: string;
    key?: string;
    photo?: string;
    status?: string;
}

export interface SystemTaskScopeApp {
    id: string;
    name: string;
    key?: string;
    photo?: string;
    status?: string;
    env?: string;
}

export interface SystemTaskScopeUser {
    id: string;
    username: string;
    email?: string;
    fullName?: string;
    photo?: string;
    role?: string;
}

export type SystemTaskScope =
    | { type: "global" }
    | { type: "project"; projectID: string }
    | { type: "project-env"; projectID: string; projectEnvID: string }
    | { type: "app"; projectID: string; projectEnvID: string; appID: string };

export const SystemTaskSequenceStepStatus = {
    Pending: "pending",
    Running: "running",
    Done: "done",
    Failed: "failed",
    Skipped: "skipped",
} as const;

export type SystemTaskSequenceStepStatus =
    (typeof SystemTaskSequenceStepStatus)[keyof typeof SystemTaskSequenceStepStatus];

/** How one step of a job sequence's run went. */
export interface SystemTaskSequenceStep {
    job: { id: string };
    name: string;
    status: SystemTaskSequenceStepStatus;
    exitCode: number | null;
    error: string;
    attempts: number;
    startedAt: Date | null;
    endedAt: Date | null;
    /** What the step wrote to its output file, by key. */
    outputs: Record<string, string>;
}

/** A job sequence's run, which its task keeps as it goes. */
export interface SystemTaskSequenceRun {
    currentStep: number;
    steps: SystemTaskSequenceStep[];
}

/** What fired a scheduled job's run: an event of an app. */
export interface SystemTaskTrigger {
    event: string;
    app: { id: string; name: string };
    deploymentId: string;
}

/** The snapshot a data backup's run took. */
export interface SystemTaskDataBackup {
    snapshotId: string;
    sizeBytes: number;
}

/** What a function's call answered: the runtime's outcome, and the response. */
export interface SystemTaskFunctionInvoke {
    /** ok, error or timeout. */
    outcome: string;
    status: number;
    headers: Record<string, string[]>;
    /** Decoded from base64; cut at 64 KB when bodyTruncated. */
    body: Uint8Array;
    bodyTruncated: boolean;
    requestId: string;
    durationMs: number;
}

/** What a restore's task restored, from where, and how. */
export interface SystemTaskBackupRestore {
    /** The snapshot's record; snapshotId is the repository's. */
    snapshotRecordId: string;
    repoId: string;
    snapshotId: string;
    snapshotPath: string;
    fileName: string;
    mode: string;
    stopApp: boolean;
}

export interface SystemTask {
    id: string;
    type: string;
    status: SystemTaskStatus;
    config: SystemTaskConfig;
    targetJob?: SystemTaskTargetJob;
    lastError: string;
    updateVer: number;
    scopeProject?: SystemTaskScopeProject;
    scopeApp?: SystemTaskScopeApp;
    scopeUser?: SystemTaskScopeUser;
    sequenceRun?: SystemTaskSequenceRun;
    trigger?: SystemTaskTrigger;
    dataBackup?: SystemTaskDataBackup;
    functionInvoke?: SystemTaskFunctionInvoke;
    backupRestore?: SystemTaskBackupRestore;
    runAt: Date | null;
    retryAt: Date | null;
    startedAt: Date | null;
    endedAt: Date | null;
    createdAt: Date;
    updatedAt: Date;
}
