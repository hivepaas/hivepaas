import type { EProjectSecretStatus } from "~/projects/module-shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export interface ProjectSecret {
    id: string;
    name: string;
    updateVer: number;
    key: string;
    value?: string;
    secretMasked?: boolean;
    base64: boolean;
    inherited?: boolean;
    status: OpenApiConstant<EProjectSecretStatus>;

    createdAt: Date;
    updatedAt: Date | null;
}
