export const EAppScheduledJobType = {
    ContainerCommand: "container-command",
    SystemCleanup: "system-cleanup",
    SystemBackup: "system-backup",
    SSLRenewal: "ssl-renewal",
} as const;

export type EAppScheduledJobType = (typeof EAppScheduledJobType)[keyof typeof EAppScheduledJobType];
