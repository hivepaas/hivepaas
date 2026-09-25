import type { EProjectSecretStatus } from "~/projects/module-shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export interface AppConfigFile {
    id: string;
    name: string;
    content: string;
    base64: boolean;
    type: string;
    status: OpenApiConstant<EProjectSecretStatus>;
    inherited: boolean;
    updateVer: number;

    createdAt: Date;
    updatedAt: Date | null;
    expireAt: Date | null;
}
