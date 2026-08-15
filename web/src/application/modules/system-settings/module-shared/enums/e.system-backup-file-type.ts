export const ESystemBackupFileType = {
    System: "system",
} as const;

export type ESystemBackupFileType = (typeof ESystemBackupFileType)[keyof typeof ESystemBackupFileType];
