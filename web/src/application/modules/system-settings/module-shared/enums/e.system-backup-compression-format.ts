export const ESystemBackupCompressionFormat = {
    None: "",
    Gzip: "gzip",
} as const;

export type ESystemBackupCompressionFormat =
    (typeof ESystemBackupCompressionFormat)[keyof typeof ESystemBackupCompressionFormat];
