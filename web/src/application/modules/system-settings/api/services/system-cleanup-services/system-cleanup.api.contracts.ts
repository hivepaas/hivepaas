import type {
    SystemCleanupBuildCacheClearResult,
    SystemCleanupExecuteResult,
    SystemCleanupRepoCacheClearResult,
    SystemCleanupRepoCacheInfo,
    SystemCleanupSettings,
} from "~/system-settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type SystemCleanup_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type SystemCleanup_FindOne_Res = ApiResponseBase<SystemCleanupSettings>;

export type SystemCleanup_UpdateOne_Payload = {
    updateVer: number;
    status: ESettingStatus;
    schedule: {
        interval: string;
        cronExpr: string;
        initialTime?: Date;
    };
    dbObjectRetention: {
        enabled: boolean;
        tasks: string;
        sysErrors: string;
        deployments: string;
        deletedObjects: string;
    };
    clusterCleanup: {
        enabled: boolean;
        pruneImages: boolean;
        pruneVolumes: boolean;
        pruneNetworks: boolean;
        pruneContainers: boolean;
    };
    backupCleanup: {
        enabled: boolean;
        cloudBackupRetention: string;
        localBackupRetention: string;
    };
    cacheCleanup: {
        enabled: boolean;
        repoCacheRetention: string;
    };
    fileCleanup: {
        enabled: boolean;
    };
    notification: {
        success: {
            id: string;
        };
        successUseDefault: boolean;
        failure: {
            id: string;
        };
        failureUseDefault: boolean;
    };
};

export type SystemCleanup_UpdateOne_Req = ApiRequestBase<{
    payload: SystemCleanup_UpdateOne_Payload;
}>;
export type SystemCleanup_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type SystemCleanup_Execute_Req = ApiRequestBase<Record<string, never>>;
export type SystemCleanup_Execute_Res = ApiResponseBase<SystemCleanupExecuteResult>;

export type SystemCleanup_FindRepoCache_Req = ApiRequestBase<Record<string, never>>;
export type SystemCleanup_FindRepoCache_Res = ApiResponseBase<SystemCleanupRepoCacheInfo>;

export type SystemCleanup_ClearRepoCache_Req = ApiRequestBase<Record<string, never>>;
export type SystemCleanup_ClearRepoCache_Res = ApiResponseBase<SystemCleanupRepoCacheClearResult>;

export type SystemCleanup_ClearBuildCache_Req = ApiRequestBase<Record<string, never>>;
export type SystemCleanup_ClearBuildCache_Res = ApiResponseBase<SystemCleanupBuildCacheClearResult>;
