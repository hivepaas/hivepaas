import type { EProjectStatus } from "~/projects/module-shared/enums";

export interface ProjectBaseEntity {
    id: string;
    name: string;
    key: string;
    status: EProjectStatus;
    photo: string;
    note: string;
    tags: string[];
    updateVer: number;

    createdAt: Date;
    updatedAt: Date | null;
}
