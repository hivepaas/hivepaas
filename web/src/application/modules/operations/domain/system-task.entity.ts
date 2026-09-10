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
    runAt: Date | null;
    retryAt: Date | null;
    startedAt: Date | null;
    endedAt: Date | null;
    createdAt: Date;
    updatedAt: Date;
}
