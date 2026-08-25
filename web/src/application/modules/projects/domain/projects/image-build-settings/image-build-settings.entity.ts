import type { ESettingStatus } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export interface ProjectImageBuildSettings {
    id: string;
    type: string;
    name: string;
    kind?: string;
    status: OpenApiConstant<ESettingStatus>;
    inherited?: boolean;
    inheritable?: boolean;
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
    repoCache: boolean;
}

export interface ProjectImageBuildRepoCacheInfo {
    totalFiles: number;
    totalSizeBytes: number;
}

export interface ProjectImageBuildRepoCacheClearResult {
    filesDeleted: number;
    spaceReclaimed: number;
}
