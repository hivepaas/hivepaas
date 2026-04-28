import type { SettingsBaseEntity } from "./settings.base.entity";

export interface StorageBindSettings {
    enabled?: boolean;
    baseDirs?: string[];
    baseSubpath?: string;
    appsMustUseSubPaths?: boolean;
}

export interface StorageVolumeSettings {
    enabled?: boolean;
    volumes?: { id: string; name: string }[];
    baseSubpath?: string;
    appsMustUseSubPaths?: boolean;
}

export interface StorageClusterVolumeSettings {
    enabled?: boolean;
    volumes?: { id: string; name: string }[];
    baseSubpath?: string;
    appsMustUseSubPaths?: boolean;
}

export interface StorageTmpfsSettings {
    enabled?: boolean;
    maxSize?: number;
}

export interface SettingStorageSettings extends SettingsBaseEntity {
    bindSettings?: StorageBindSettings;
    volumeSettings?: StorageVolumeSettings;
    clusterVolumeSettings?: StorageClusterVolumeSettings;
    tmpfsSettings?: StorageTmpfsSettings;
}
