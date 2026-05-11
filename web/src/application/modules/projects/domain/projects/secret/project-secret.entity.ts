import type { EProjectSecretStatus } from "~/projects/module-shared/enums";

export interface ProjectSecret {
    id: string;
    name: string;
    updateVer: number;
    key: string;
    base64: boolean;
    status: EProjectSecretStatus;

    createdAt: Date;
    updatedAt: Date | null;
}
