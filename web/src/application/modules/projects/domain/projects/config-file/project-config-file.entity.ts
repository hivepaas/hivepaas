import type { EProjectSecretStatus } from "~/projects/module-shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export interface ProjectConfigFile {
    id: string;
    name: string;
    content: string;
    base64: boolean;
    /** Apps of the project, or of the env, get it. */
    inheritable: boolean;
    inherited: boolean;
    status: OpenApiConstant<EProjectSecretStatus>;
    updateVer: number;

    createdAt: Date;
    updatedAt: Date | null;
}
