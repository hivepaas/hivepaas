import type { EProjectEnvStatus, EProjectStatus } from "~/projects/module-shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export interface ProjectBaseEntity {
    id: string;
    name: string;
    key: string;
    status: OpenApiConstant<EProjectStatus>;
    photo: string;
    note: string;
    envs: ProjectEnvEntity[];
    tags: string[];
    updateVer: number;

    createdAt: Date;
    updatedAt: Date | null;
}

export interface ProjectEnvEntity {
    id?: string;
    name: string;
    color: string;
    status?: EProjectEnvStatus;
    updateVer?: number;
}
