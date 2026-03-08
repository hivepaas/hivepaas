import type { EProjectSecretStatus } from "~/projects/module-shared/enums";

export interface AppSecret {
    id: string;
    name: string;
    updateVer: number;
    key: string;
    status: EProjectSecretStatus;
    inherited: boolean;

    createdAt: Date;
    updatedAt: Date | null;
}
