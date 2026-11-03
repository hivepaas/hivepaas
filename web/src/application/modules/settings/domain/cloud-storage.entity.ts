import type { SettingsBaseEntity } from "~/settings/domain";

import type { ECloudStorageKind } from "@application/shared/enums";

/** The key auth a bucket is reached with: its status is missing when it no longer exists. */
export interface SettingCloudStorageKeyAuthRef {
    id: string;
    name: string;
    status?: string;
}

export interface SettingCloudStorageS3 {
    keyAuth: SettingCloudStorageKeyAuthRef | null;
    region: string;
    bucket: string;
    endpoint: string;
}

/** What a create or an update sends: the key auth by id alone. */
export interface SettingCloudStorageS3Payload {
    keyAuth: { id: string };
    region: string;
    bucket: string;
    endpoint: string;
}

export interface SettingCloudStorage extends SettingsBaseEntity {
    kind?: ECloudStorageKind;
    s3: SettingCloudStorageS3;
    inherited?: boolean;
}
