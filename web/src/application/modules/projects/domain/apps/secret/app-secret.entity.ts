import type { EProjectSecretStatus } from "~/projects/module-shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export interface AppSecret {
    id: string;
    name: string;
    updateVer: number;
    key: string;
    value?: string;
    secretMasked?: boolean;
    base64: boolean;
    inheritable: boolean;
    type: string;
    status: OpenApiConstant<EProjectSecretStatus>;
    inherited: boolean;

    createdAt: Date;
    updatedAt: Date | null;
    expireAt: Date | null;
}
