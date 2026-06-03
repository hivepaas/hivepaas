import type { ESettingStatus } from "@application/shared/enums";

export interface ProjectImageBuildSettings {
    id: string;
    type: string;
    name: string;
    kind?: string;
    status: ESettingStatus;
    inherited?: boolean;
    availableInProjects?: boolean;
    default?: boolean;
    updateVer: number;
    createdAt: Date;
    updatedAt: Date;
    expireAt?: Date | null;
    resources: ProjectImageBuildResourceSettings;
    sources: ProjectImageBuildSourceSettings;
    noCache: boolean;
    noVerbose: boolean;
}

export interface ProjectImageBuildResourceSettings {
    cpus?: number;
    mem?: string;
    memSwap?: string;
    shmSize?: string;
}

export interface ProjectImageBuildSourceSettings {
    checkoutMaxDepth?: number;
    repoCache: boolean;
}

export interface ProjectImageBuildRepoCacheInfo {
    totalFiles: number;
    totalSizeHR: string;
}
