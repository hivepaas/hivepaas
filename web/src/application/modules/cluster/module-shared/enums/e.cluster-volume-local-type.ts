export const EClusterVolumeLocalType = {
    Bind: "bind",
    Nfs: "nfs",
    Tmpfs: "tmpfs",
} as const;

export type EClusterVolumeLocalType = (typeof EClusterVolumeLocalType)[keyof typeof EClusterVolumeLocalType];
