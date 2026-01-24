import type { ProjectBaseEntity } from "./project.base.entity";

export interface ProjectDetailsEntity extends ProjectBaseEntity {
    photo: string;
    userAccesses: {
        id: string;
        fullName: string;
        access: {
            read: boolean;
            write: boolean;
            delete: boolean;
        };
    }[];
}
