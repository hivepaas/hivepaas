export const EAppScheduledJobType = {
    ContainerCommand: "container-command",
    SystemCleanup: "system-cleanup",
    SystemBackup: "system-backup",
    SSLRenewal: "ssl-renewal",
    RegistryAuthRenewal: "registry-auth-renewal",
    BackupRepoCleanup: "backup-repo-cleanup",
    JobSequence: "job-sequence",
    DataBackup: "data-backup",
    /** Calls a function once with a request: its runtime's invoke, in a running task of the function. */
    FunctionInvoke: "function-invoke",
} as const;

export type EAppScheduledJobType = (typeof EAppScheduledJobType)[keyof typeof EAppScheduledJobType];
