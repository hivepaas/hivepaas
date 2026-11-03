export const BACKUP_REPO_ENGINE = {
    Kopia: "kopia",
} as const;

export type BackupRepoEngine = (typeof BACKUP_REPO_ENGINE)[keyof typeof BACKUP_REPO_ENGINE];

export const BACKUP_REPO_ENGINE_OPTIONS = [{ label: "Kopia", value: BACKUP_REPO_ENGINE.Kopia }] as const;

export const KOPIA_COMPRESSION_LEVELS = [
    "none",
    "deflate-best-compression",
    "deflate-best-speed",
    "deflate-default",
    "gzip",
    "gzip-best-compression",
    "gzip-best-speed",
    "lz4",
    "pgzip",
    "pgzip-best-compression",
    "pgzip-best-speed",
    "s2-better",
    "s2-default",
    "s2-parallel-4",
    "s2-parallel-8",
    "zstd",
    "zstd-best-compression",
    "zstd-better-compression",
    "zstd-fastest",
] as const;

export type KopiaCompressionLevel = (typeof KOPIA_COMPRESSION_LEVELS)[number];

export const BACKUP_REPO_ENGINE_COMPRESSION_OPTIONS: Record<string, readonly string[]> = {
    [BACKUP_REPO_ENGINE.Kopia]: KOPIA_COMPRESSION_LEVELS,
};

export const DEFAULT_BACKUP_REPO_COMPRESSION = "zstd-fastest";
export const DEFAULT_BACKUP_REPO_PACK_SIZE = "32mb";

export const DEFAULT_BACKUP_REPO_RETENTION = {
    keepLast: 10,
    keepHourly: 48,
    keepDaily: 7,
    keepWeekly: 4,
    keepMonthly: 24,
} as const;

export const BACKUP_REPO_ACTION = {
    CreateNew: "create-new",
    ImportExisting: "import-existing",
} as const;

export type BackupRepoAction = (typeof BACKUP_REPO_ACTION)[keyof typeof BACKUP_REPO_ACTION];

export const BACKUP_REPO_STORAGE_TYPE = {
    CloudStorage: "cloudStorage",
    Volume: "volume",
} as const;

export type BackupRepoStorageType = (typeof BACKUP_REPO_STORAGE_TYPE)[keyof typeof BACKUP_REPO_STORAGE_TYPE];
