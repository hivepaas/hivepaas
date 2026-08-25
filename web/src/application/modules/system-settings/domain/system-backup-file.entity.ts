import type { SettingsBaseEntity } from "~/settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

import type { ESystemBackupFileStorageType } from "../module-shared/enums";

export interface SystemBackupFileStorage extends SettingsBaseEntity {
    status: OpenApiConstant<ESettingStatus>;
}

export interface SystemBackupFile {
    id: string;
    type: string;
    kind: string;
    status: OpenApiConstant<ESettingStatus>;
    key: string;
    name: string;
    path: string;
    bucket?: string;
    mimetype: string;
    sizeBytes: number;
    storageType: ESystemBackupFileStorageType;
    storage?: SystemBackupFileStorage | null;
    updateVer: number;
    createdAt: Date;
    updatedAt: Date;
}
