import type { SettingsBaseEntity } from "~/settings/domain";

import type { ESettingStatus, ESettingType } from "@application/shared/enums";

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
    scheduleInterval: string;
    scheduleFrom?: Date | null;
    dbObjectRetention: SystemCleanupDBObjectRetention;
    clusterCleanup: SystemCleanupClusterCleanup;
    backupCleanup: SystemCleanupBackupCleanup;
    notification?: SystemCleanupNotification | null;
}
