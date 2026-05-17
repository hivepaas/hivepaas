import type { SettingsBaseEntity } from "~/settings/domain";

import type { ECloudStorageKind } from "@application/shared/enums";

export interface SettingCloudStorageS3 {
    accessKeyId: string;
    secretKey: string;
    region: string;
    bucket: string;
    endpoint: string;
}

export interface SettingCloudStorage extends SettingsBaseEntity {
    kind?: ECloudStorageKind;
    s3: SettingCloudStorageS3;
    secretMasked?: boolean;
    inherited?: boolean;
}
