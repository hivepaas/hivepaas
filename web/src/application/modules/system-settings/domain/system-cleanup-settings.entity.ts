import type { SettingsBaseEntity } from "~/settings/domain";

import type { ESettingStatus, ESettingType } from "@application/shared/enums";

export interface SystemCleanupSchedule {
    interval: string;
    cronExpr: string;
    initialTime?: Date | null;
}

export interface SystemCleanupDBObjectRetention {
    enabled: boolean;
    tasks: string;
    sysErrors: string;
    deployments: string;
    deletedObjects: string;
}

export interface SystemCleanupClusterCleanup {
    enabled: boolean;
    pruneImages: boolean;
    pruneVolumes: boolean;
    pruneNetworks: boolean;
    pruneContainers: boolean;
}

export interface SystemCleanupBackupCleanup {
    enabled: boolean;
    cloudBackupRetention: string;
    localBackupRetention: string;
}

export interface SystemCleanupCacheCleanup {
    enabled: boolean;
    repoCacheRetention: string;
}

export interface SystemCleanupFileCleanup {
    enabled: boolean;
}

export interface SystemCleanupNotification {
    success?: {
        id: string;
        name: string;
    };
    successUseDefault: boolean;
    failure?: {
        id: string;
        name: string;
    };
    failureUseDefault: boolean;
}

export interface SystemCleanupSettings extends SettingsBaseEntity {
    type: typeof ESettingType.SystemCleanup;
    status: ESettingStatus;
    schedule: SystemCleanupSchedule;
    dbObjectRetention: SystemCleanupDBObjectRetention;
    clusterCleanup: SystemCleanupClusterCleanup;
    backupCleanup: SystemCleanupBackupCleanup;
    cacheCleanup: SystemCleanupCacheCleanup;
    fileCleanup: SystemCleanupFileCleanup;
    notification?: SystemCleanupNotification | null;
    nextRuns: Date[];
}

export interface SystemCleanupExecuteResult {
    task: {
        id: string;
    };
}

export interface SystemCleanupRepoCacheInfo {
    totalFiles: number;
    totalSizeBytes: number;
}

export interface SystemCleanupRepoCacheClearResult {
    filesDeleted: number;
    spaceReclaimed: number;
}

export interface SystemCleanupBuildCacheClearResult {
    cachesDeleted: number;
    spaceReclaimed: number;
}
