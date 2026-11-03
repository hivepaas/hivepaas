import type { EProjectAppStatus } from "~/projects/module-shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export interface ProjectAppBase {
    id: string;
    name: string;
    photo: string;
    status: OpenApiConstant<EProjectAppStatus>;
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
    status: OpenApiConstant<EProjectAppStatus> | "";
    env: string;
}
