import { type Profile } from "@application/shared/entities";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

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

// /**
//  * Update profile password
//  */
// export type Profile_UpdateProfilePassword_Req = ApiRequestBase<UpdateProfilePassword>;

// export type Profile_UpdateProfilePassword_Res = ApiResponseBase<{
//     type: "success";
// }>;

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
