export const EMountType = {
    Bind: "bind",
    Volume: "volume",
    Tmpfs: "tmpfs",
    Npipe: "npipe",
    Cluster: "cluster",
    Image: "image",
} as const;

export type EMountType = (typeof EMountType)[keyof typeof EMountType];
