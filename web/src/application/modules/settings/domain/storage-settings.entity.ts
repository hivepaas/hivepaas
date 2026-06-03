import type { SettingsBaseEntity } from "./settings.base.entity";

export interface StorageBindSettings {
    enabled?: boolean;
    baseDirs: string[];
    subpathTemplate?: string;
    subpathRequired?: string;
}

export interface StorageVolumeSettings {
    enabled?: boolean;
    volumes: { id: string; name: string }[];
    subpathTemplate?: string;
    subpathRequired?: string;
}

export interface StorageClusterVolumeSettings {
    enabled?: boolean;
    volumes: { id: string; name: string }[];
    subpathTemplate?: string;
    subpathRequired?: string;
}

export interface StorageTmpfsSettings {
    enabled?: boolean;
    maxSize?: string;
}

export interface SettingStorageSettings extends SettingsBaseEntity {
    bindSettings?: StorageBindSettings;
    volumeSettings?: StorageVolumeSettings;
    clusterVolumeSettings?: StorageClusterVolumeSettings;
    tmpfsSettings?: StorageTmpfsSettings;
}
