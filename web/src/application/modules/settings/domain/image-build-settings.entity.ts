import type { SettingsBaseEntity } from "./settings.base.entity";

export interface ImageBuildWorkerNode {
    id: string;
    name: string;
}

export interface ImageBuildWorkerSettings {
    nodes: ImageBuildWorkerNode[];
    nodeLabels: string[];
    maxParallelism: number;
}

export interface ImageBuildResourceSettings {
    cpus?: number;
    mem?: string;
    memSwap?: string;
    shmSize?: string;
}

export interface ImageBuildSourceSettings {
    repoCache: boolean;
}

export interface ImageBuildSettings extends SettingsBaseEntity {
    workers: ImageBuildWorkerSettings;
    resources: ImageBuildResourceSettings;
    sources: ImageBuildSourceSettings;
    noCache: boolean;
    noVerbose: boolean;
}

export interface ImageBuildRepoCacheInfo {
    totalFiles: number;
    totalSizeBytes: number;
}

export interface ImageBuildRepoCacheClearResult {
    filesDeleted: number;
    spaceReclaimed: number;
}
