export interface DockerVolume {
    id: string;
    name: string;
    driver: string;
    createdAt: Date;
    mountpoint: string;
    labels: Record<string, string>;
    options: Record<string, string>;
    scope: string;
    inheritable: boolean;
    refCount: number;
    size: number;
    updateVer: number;
    status: Record<string, unknown>;
    clusterVolumeSpe?: null;
}
