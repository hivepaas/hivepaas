import type { EProfileApiKeyStatus } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export interface ProfileApiKey {
    id: string;
    name: string;
    keyId: string;
    updateVer: number;
    accessAction: {
        read: boolean;
        execute: boolean;
        write: boolean;
        delete: boolean;
    } | null;
    expireAt?: Date;
    status: OpenApiConstant<EProfileApiKeyStatus>;
}
