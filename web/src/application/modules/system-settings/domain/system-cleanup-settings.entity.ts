import type { SettingsBaseEntity } from "~/settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export interface SystemCleanupSchedule {
    interval: string;
    cronExpr: string;
    initialTime?: Date | null;
}

export interface SystemCleanupDBObjectRetention {
    enabled: boolean;
    tasks: string;
    sysErrors: string;
    auditLogs: string;
    deployments: string;
    deletedObjects: string;
}

export interface SystemCleanupClusterCleanup {
    enabled: boolean;
    generalRetention: string;
    buildCacheRetention: string;
    pruneImages: boolean;
    pruneVolumes: boolean;
    pruneNetworks: boolean;
    pruneContainers: boolean;
    pruneBuildCache: boolean;
}

export interface SystemCleanupCacheCleanup {
    enabled: boolean;
    repoCacheRetention: string;
}

export interface SystemCleanupFileCleanup {
    enabled: boolean;
}

/**
 * Brings the apps HivePaaS runs for itself - the registry, the logging stack - to their settings, and checks OBI on the
 * nodes. On for a setting saved before it existed.
 */
export interface SystemCleanupSystemAppsSync {
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
    status: OpenApiConstant<ESettingStatus>;
    schedule: SystemCleanupSchedule;
    dbObjectRetention: SystemCleanupDBObjectRetention;
    clusterCleanup: SystemCleanupClusterCleanup;
    cacheCleanup: SystemCleanupCacheCleanup;
    fileCleanup: SystemCleanupFileCleanup;
    systemAppsSync: SystemCleanupSystemAppsSync;
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
