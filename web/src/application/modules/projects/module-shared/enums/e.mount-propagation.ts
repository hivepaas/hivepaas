export const EMountPropagation = {
    Default: "",
    RPrivate: "rprivate",
    Private: "private",
    RShared: "rshared",
    Shared: "shared",
    RSlave: "rslave",
    Slave: "slave",
} as const;

export type EMountPropagation = (typeof EMountPropagation)[keyof typeof EMountPropagation];
