export const EClusterVolumePropagation = {
    Default: "",
    RPrivate: "rprivate",
    Private: "private",
    RShared: "rshared",
    Shared: "shared",
    RSlave: "rslave",
    Slave: "slave",
} as const;

export type EClusterVolumePropagation = (typeof EClusterVolumePropagation)[keyof typeof EClusterVolumePropagation];
