import type { SystemCleanupSettings } from "~/system-settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type SystemCleanup_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type SystemCleanup_FindOne_Res = ApiResponseBase<SystemCleanupSettings>;

export type SystemCleanup_UpdateOne_Payload = {
    updateVer: number;
    status: ESettingStatus;
    scheduleInterval: string;
    scheduleFrom: Date;
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
export type SystemCleanup_Execute_Res = ApiResponseBase<{ type: "success" }>;
