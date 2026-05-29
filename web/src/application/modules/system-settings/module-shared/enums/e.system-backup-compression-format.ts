export const ESystemBackupCompressionFormat = {
    None: "",
    Gzip: "gzip",
    Zstd: "zstd",
} as const;

export type ESystemBackupCompressionFormat =
    (typeof ESystemBackupCompressionFormat)[keyof typeof ESystemBackupCompressionFormat];
