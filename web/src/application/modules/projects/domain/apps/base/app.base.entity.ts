import type { EProjectAppStatus } from "~/projects/module-shared/enums";

export interface ProjectAppBase {
    id: string;
    name: string;
    status: EProjectAppStatus;
    env: string;
    note: string;
    tags: string[];

    createdAt: Date;
    updatedAt: Date | null;
}

export interface ProjectAppBaseRef {
    id: string;
    name: string;
    key: string;
    localKey: string;
    status: EProjectAppStatus | "";
    env: string;
}
