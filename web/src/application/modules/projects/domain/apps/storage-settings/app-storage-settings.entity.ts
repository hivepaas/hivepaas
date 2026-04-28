import type { EMountConsistency, EMountPropagation, EMountType } from "~/projects/module-shared/enums";

export type AppStorageSettings = {
    mounts: AppStorageMount[];
    updateVer: number;
};

export type AppStorageMount = {
    type?: EMountType;
    target?: string;
    readOnly?: boolean;
    consistency?: EMountConsistency;
    bindOptions?: BindOptions;
    volumeOptions?: VolumeOptions;
    tmpfsOptions?: TmpfsOptions;
    clusterOptions?: ClusterOptions;
};

export type BindOptions = {
    baseDir?: string;
    subpath?: string;
    subpathRequired?: string;
    propagation?: EMountPropagation;
    nonRecursive?: boolean;
    createMountpoint?: boolean;
    readOnlyNonRecursive?: boolean;
    readOnlyForceRecursive?: boolean;
};

export type VolumeOptions = {
    volume?: string;
    subpath?: string;
    subpathRequired?: string;
    noCopy?: boolean;
    labels?: Record<string, string>;
    driverConfig?: VolumeDriver;
};

export type VolumeDriver = {
    name?: string;
    options?: Record<string, string>;
};

export type TmpfsOptions = {
    size?: string;
    mode?: number;
    options?: string[][];
};

export type ClusterOptions = {
    volume?: string;
    subpath?: string;
    subpathRequired?: string;
    noCopy?: boolean;
    labels?: Record<string, string>;
    driverConfig?: VolumeDriver;
};
