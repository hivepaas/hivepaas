import type { PaginationState, SortingState } from "@infrastructure/data";

import { type Profile } from "@application/shared/entities";
import type { ProfileApiKey } from "@application/shared/entities/profile";
import type { EProfileApiKeyStatus } from "@application/shared/enums";

import { type ApiRequestBase, type ApiResponseBase, type ApiResponsePaginated } from "@infrastructure/api";

/**
 * Update profile
 */
export type Profile_UpdateProfile_Req = ApiRequestBase<{
    profile: Pick<Profile, "fullName" | "email" | "username" | "position" | "notes"> & {
        photo?: { fileName: string; dataBase64: string } | { delete: true };
    };
}>;

export type Profile_UpdateProfile_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Update profile password
 */
export type Profile_UpdateProfilePassword_Req = ApiRequestBase<{
    currentPassword: string;
    newPassword: string;
}>;

export type Profile_UpdateProfilePassword_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Get profile 2FA setup
 */
export type Profile_GetProfile2FASetup_Req = ApiRequestBase<{
    passcode?: string;
}>;

export type Profile_GetProfile2FASetup_Res = ApiResponseBase<{
    totpToken: string;
    totpQRCode: string;
    secretKey: string;
}>;

/**
 * Verify profile 2FA setup
 */
export type Profile_Complete2FASetup_Req = ApiRequestBase<{
    totpToken: string;
    passcode: string;
}>;

export type Profile_Complete2FASetup_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Find many account API keys paginated
 */
export type Profile_FindManyApiKeysPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type Profile_FindManyApiKeysPaginated_Res = ApiResponsePaginated<ProfileApiKey>;

/**
 * Create one account API key
 */
export type Profile_CreateOneApiKey_Req = ApiRequestBase<{
    name: string;
    accessAction: {
        read: boolean;
        write: boolean;
        delete: boolean;
    };
    expireAt?: Date;
}>;

export type Profile_CreateOneApiKey_Res = ApiResponseBase<{
    id: string;
    keyId: string;
    secretKey: string;
}>;

/**
 * Delete one account API key
 */
export type Profile_DeleteOneApiKey_Req = ApiRequestBase<{
    id: string;
}>;

export type Profile_DeleteOneApiKey_Res = ApiResponseBase<{
    id: string;
}>;

/**
 * update one account API key status
 */
export type Profile_UpdateOneApiKeyStatus_Req = ApiRequestBase<{
    id: string;
    status: EProfileApiKeyStatus;
    expireAt?: Date;
    updateVer: number;
}>;

export type Profile_UpdateOneApiKeyStatus_Res = ApiResponseBase<{
    type: "success";
}>;
