import type { EProjectStatus } from "~/projects/module-shared/enums";

export interface ProjectBaseEntity {
    id: string;
    name: string;
    key: string;
    status: EProjectStatus;
    photo: string;
    note: string;
    envs: ProjectEnvEntity[];
    tags: string[];
    updateVer: number;

    createdAt: Date;
    updatedAt: Date | null;
}

export interface ProjectEnvEntity {
    name: string;
    color: string;
}
