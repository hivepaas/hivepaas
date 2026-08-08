import type { EMountConsistency, EMountPropagation, EMountType } from "~/projects/module-shared/enums";

export type AppStorageSettings = {
    mounts: AppStorageMount[];
    updateVer: number;
};

export type AppStorageMount = {
    key?: string;
    type?: EMountType;
    source?: string;
    target?: string;
    readOnly?: boolean;
    consistency?: EMountConsistency;
    bindOptions?: BindOptions;
    volumeOptions?: VolumeOptions;
    tmpfsOptions?: TmpfsOptions;
    clusterOptions?: ClusterOptions;
};

export type BindOptions = {
    propagation?: EMountPropagation;
    nonRecursive?: boolean;
    createMountpoint?: boolean;
    readOnlyNonRecursive?: boolean;
    readOnlyForceRecursive?: boolean;
};

export type VolumeOptions = {
    subpath?: string;
    noCopy?: boolean;
    labels?: Record<string, string>;
    driverConfig?: VolumeDriver | null;
};

export type VolumeDriver = {
    name?: string;
    options?: Record<string, string>;
};

export type TmpfsOptions = {
    size?: string;
    mode?: string;
    options?: string[][] | null;
};

export type ClusterOptions = {
    subpath?: string;
    noCopy?: boolean;
    labels?: Record<string, string>;
    driverConfig?: VolumeDriver | null;
};
