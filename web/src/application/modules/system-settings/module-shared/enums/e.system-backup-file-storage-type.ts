export const ESystemBackupFileStorageType = {
    Local: "local",
    Cloud: "cloud",
} as const;

export type ESystemBackupFileStorageType =
    (typeof ESystemBackupFileStorageType)[keyof typeof ESystemBackupFileStorageType];
