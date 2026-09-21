import type { EMountConsistency, EMountPropagation, EMountType } from "~/projects/module-shared/enums";

export type AppStorageSettings = {
    mounts: AppStorageMount[];
    /**
     * The apps that have been given a directory of this app's storage. It comes
     * back from the API and is not sent, so an update payload leaves it out.
     */
    borrowedBy?: MountBorrower[];
    updateVer: number;
};

export type MountBorrower = {
    appId: string;
    name: string;
    target: string;
    subpath?: string;
    write?: boolean;
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
    sourceApp?: MountSourceApp;
};

/**
 * The directory this mount reaches belongs to another app - the files of the
 * database a file manager was put there to work on. Absent is the ordinary case:
 * the app's own directory.
 *
 * `name` and `dangling` come back from the API and are not sent; `dangling`
 * means the directory is still there and the app that owned it is not.
 */
export type MountSourceApp = {
    appId?: string;
    write?: boolean;
    name?: string;
    dangling?: boolean;
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
