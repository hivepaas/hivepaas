import type { EProjectSecretStatus } from "~/projects/module-shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export interface AppSecretSwarmRefFile {
    name: string;
    uid: string;
    gid: string;
    mode: string;
}

export interface AppSecretSwarmRef {
    file: AppSecretSwarmRefFile | null;
}

export interface AppSecret {
    id: string;
    name: string;
    updateVer: number;
    key: string;
    base64: boolean;
    type: string;
    status: OpenApiConstant<EProjectSecretStatus>;
    inherited: boolean;
    swarmRef: AppSecretSwarmRef | null;

    createdAt: Date;
    updatedAt: Date | null;
    expireAt: Date | null;
}
