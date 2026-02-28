import type { EProjectAppStatus } from "~/projects/module-shared/enums";

export interface ProjectAppBase {
    id: string;
    name: string;
    status: EProjectAppStatus;
    note: string;
    tags: string[];

    createdAt: Date;
    updatedAt: Date | null;
}
