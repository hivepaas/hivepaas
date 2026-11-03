import type { SystemBackupRepoCleanupSettings } from "~/system-settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type SystemBackupRepoCleanup_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type SystemBackupRepoCleanup_FindOne_Res = ApiResponseBase<SystemBackupRepoCleanupSettings>;

export type SystemBackupRepoCleanup_UpdateOne_Payload = {
    updateVer: number;
    status: ESettingStatus;
    schedule: {
        interval: string;
        cronExpr: string;
        initialTime?: Date;
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

export type SystemBackupRepoCleanup_UpdateOne_Req = ApiRequestBase<{
    payload: SystemBackupRepoCleanup_UpdateOne_Payload;
}>;
export type SystemBackupRepoCleanup_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type SystemBackupRepoCleanup_Execute_Req = ApiRequestBase<{
    targetRepos: { id: string }[];
}>;
export type SystemBackupRepoCleanup_Execute_Res = ApiResponseBase<{
    task: {
        id: string;
    };
}>;
