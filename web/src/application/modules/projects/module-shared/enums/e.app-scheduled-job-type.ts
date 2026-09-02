export const EAppScheduledJobType = {
    ContainerCommand: "container-command",
    SystemCleanup: "system-cleanup",
    SystemBackup: "system-backup",
    SSLRenewal: "ssl-renewal",
    BackupRepoCleanup: "backup-repo-cleanup",
} as const;

export type EAppScheduledJobType = (typeof EAppScheduledJobType)[keyof typeof EAppScheduledJobType];
