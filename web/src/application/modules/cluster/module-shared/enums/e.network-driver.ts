export const EClusterNetworkDriver = {
    Overlay: "overlay",
    Bridge: "bridge",
} as const;

export type EClusterNetworkDriver = (typeof EClusterNetworkDriver)[keyof typeof EClusterNetworkDriver];
