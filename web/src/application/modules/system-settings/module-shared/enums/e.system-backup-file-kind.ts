export const ESystemBackupFileKind = {
    SystemBackup: "system-backup",
} as const;

export type ESystemBackupFileKind = (typeof ESystemBackupFileKind)[keyof typeof ESystemBackupFileKind];
