import type { EProjectSecretStatus } from "~/projects/module-shared/enums";

export interface AppConfigFileSwarmRefFile {
    name: string;
    uid: string;
    gid: string;
    mode: string;
}

export interface AppConfigFileSwarmRef {
    file: AppConfigFileSwarmRefFile | null;
}

export interface AppConfigFile {
    id: string;
    name: string;
    content: string;
    base64: boolean;
    type: string;
    status: EProjectSecretStatus;
    inherited: boolean;
    swarmRef: AppConfigFileSwarmRef | null;
    updateVer: number;

    createdAt: Date;
    updatedAt: Date | null;
    expireAt: Date | null;
}
