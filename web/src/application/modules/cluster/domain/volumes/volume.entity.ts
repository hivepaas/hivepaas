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
    /**
     * Where the volume's data is, by node id or by node label - never both.
     *
     * Both empty is an answer rather than a gap: it says the volume is reachable
     * from every node, either a swarm cluster volume or a path backed by storage
     * mounted the same way everywhere.
     */
    nodeId?: string;
    nodeLabel?: string;
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
    /** See ClusterVolume.nodeId. "current" asks the server to resolve its own node. */
    nodeId?: string;
    nodeLabel?: string;
    bindOptions?: ClusterVolumeBindOptions | null;
    nfsOptions?: ClusterVolumeNfsOptions | null;
    tmpfsOptions?: ClusterVolumeTmpfsOptions | null;
    btrfsOptions?: ClusterVolumeBtrfsOptions | null;
    options: Record<string, string>;
    labels: Record<string, string>;
}

export interface ClusterVolumeCreatePayload extends ClusterVolumeBasePayload {
    inheritable: boolean;
    default: boolean;
}

export interface ClusterVolumeUpdatePayload {
    updateVer: number;
    inheritable: boolean;
    default: boolean;
    /**
     * The pinning moves as a pair: sending either field replaces both, and
     * leaving both out keeps whatever the volume already has.
     */
    nodeId?: string;
    nodeLabel?: string;
}

export interface ClusterVolumeUpdateStatusPayload extends ClusterVolumeUpdatePayload {
    status?: SettingsBaseEntity["status"];
    expireAt?: Date | null;
}
