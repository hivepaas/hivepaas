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
    /** The owner's capabilities the key may use, such as cap::secret::reveal. */
    capabilities: string[];
    expireAt?: Date;
    status: OpenApiConstant<EProfileApiKeyStatus>;
}
