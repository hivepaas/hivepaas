import type { SettingsBaseEntity } from "~/settings/domain";

import type { EClusterVolumePropagation } from "../../module-shared/enums";

export interface ClusterVolumeBindOptions {
    directory?: string;
    propagation?: EClusterVolumePropagation;
    readonly?: boolean;
    extraOptions?: string;
}

export interface ClusterVolumeNfsOptions {
    addr?: string;
    device?: string;
    version?: string;
    readonly?: boolean;
    extraOptions?: string;
}

export interface ClusterVolumeTmpfsOptions {
    device?: string;
    size?: string;
    uid?: string | number;
    gid?: string | number;
    mode?: string;
    extraOptions?: string;
}

export interface ClusterVolumeBtrfsOptions {
    device?: string;
    subvol?: string;
    readonly?: boolean;
    extraOptions?: string;
}

export interface ClusterVolume extends SettingsBaseEntity {
    driver: string;
    scope: string;
    mountpoint: string;
    options: Record<string, string>;
    labels: Record<string, string>;
    refCount: number;
    size: number;
    clusterVolumeSpec?: unknown;
    bindOptions?: ClusterVolumeBindOptions | null;
    nfsOptions?: ClusterVolumeNfsOptions | null;
    tmpfsOptions?: ClusterVolumeTmpfsOptions | null;
    btrfsOptions?: ClusterVolumeBtrfsOptions | null;
}

export interface ClusterVolumeBasePayload {
    name: string;
    driver: string;
    bindOptions?: ClusterVolumeBindOptions | null;
    nfsOptions?: ClusterVolumeNfsOptions | null;
    tmpfsOptions?: ClusterVolumeTmpfsOptions | null;
    btrfsOptions?: ClusterVolumeBtrfsOptions | null;
    options: Record<string, string>;
    labels: Record<string, string>;
}

export interface ClusterVolumeCreatePayload extends ClusterVolumeBasePayload {
    availableInProjects: boolean;
    default: boolean;
}

export interface ClusterVolumeUpdatePayload {
    updateVer: number;
    availableInProjects: boolean;
    default: boolean;
}

export interface ClusterVolumeUpdateStatusPayload extends ClusterVolumeUpdatePayload {
    status?: SettingsBaseEntity["status"];
    expireAt?: Date | null;
}
