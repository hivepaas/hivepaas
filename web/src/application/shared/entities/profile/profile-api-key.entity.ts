import type { EProfileApiKeyStatus } from "@application/shared/enums";

export interface ProfileApiKey {
    id: string;
    name: string;
    keyId: string;
    accessAction: {
        read: boolean;
        write: boolean;
        delete: boolean;
    } | null;
    expireAt?: Date;
    status: EProfileApiKeyStatus;
}
