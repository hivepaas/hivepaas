export const ESystemBackupFileType = {
    SystemBackup: "system-backup",
} as const;

export type ESystemBackupFileType = (typeof ESystemBackupFileType)[keyof typeof ESystemBackupFileType];
