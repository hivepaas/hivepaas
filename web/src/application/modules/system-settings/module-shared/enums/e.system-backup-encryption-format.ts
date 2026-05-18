export const ESystemBackupEncryptionFormat = {
    None: "",
    Age: "age",
} as const;

export type ESystemBackupEncryptionFormat =
    (typeof ESystemBackupEncryptionFormat)[keyof typeof ESystemBackupEncryptionFormat];
