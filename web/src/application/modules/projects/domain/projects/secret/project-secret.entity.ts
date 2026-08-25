import type { EProjectSecretStatus } from "~/projects/module-shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export interface ProjectSecret {
    id: string;
    name: string;
    updateVer: number;
    key: string;
    base64: boolean;
    status: OpenApiConstant<EProjectSecretStatus>;

    createdAt: Date;
    updatedAt: Date | null;
}
