import type { ProjectBaseEntity } from "./project.base.entity";

export interface ProjectUserBaseEntity {
    id: string;
    username: string;
    email: string;
    fullName: string;
    photo: string;
}

export interface ProjectDetailsEntity extends ProjectBaseEntity {
    photo: string;
    owner: ProjectUserBaseEntity;
    userAccesses: (ProjectUserBaseEntity & {
        access: {
            read: boolean;
            execute: boolean;
            write: boolean;
            delete: boolean;
        };
    })[];
}
