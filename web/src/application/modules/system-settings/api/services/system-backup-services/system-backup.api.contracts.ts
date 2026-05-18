import type { SystemBackupSettings } from "~/system-settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

import type { ESystemBackupCompressionFormat, ESystemBackupEncryptionFormat } from "../../../module-shared/enums";

export type SystemBackup_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type SystemBackup_FindOne_Res = ApiResponseBase<SystemBackupSettings>;

export type SystemBackup_UpdateOne_Payload = {
    updateVer: number;
    status: ESettingStatus;
    scheduleInterval: string;
    scheduleFrom: Date;
    compression: {
        format: ESystemBackupCompressionFormat;
    };
    encryption: {
        format: ESystemBackupEncryptionFormat;
        secret: string;
    };
    cloudStorage: {
        id: string;
        destinationDir: string;
    };
    dbBackupConfig: {
        backupDeletedObjects: boolean;
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

export type SystemBackup_UpdateOne_Req = ApiRequestBase<{
    payload: SystemBackup_UpdateOne_Payload;
}>;
export type SystemBackup_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
