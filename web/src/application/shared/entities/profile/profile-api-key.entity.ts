import type { EProfileApiKeyAction, EProfileApiKeyStatus } from "@application/shared/enums";

export interface ProfileApiKey {
    id: string;
    name: string;
    keyId: string;
    accessAction?: EProfileApiKeyAction;
    expireAt?: Date;
    status: EProfileApiKeyStatus;
}
