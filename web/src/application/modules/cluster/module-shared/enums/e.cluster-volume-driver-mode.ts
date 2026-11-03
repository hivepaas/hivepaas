export const EClusterVolumeDriverMode = {
    Local: "local",
    Custom: "custom",
} as const;

export type EClusterVolumeDriverMode = (typeof EClusterVolumeDriverMode)[keyof typeof EClusterVolumeDriverMode];
