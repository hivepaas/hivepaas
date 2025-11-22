// import {
//     type UpdateProfile,
//     type UpdateProfileEmail,
//     type UpdateProfileLocale,
//     type UpdateProfilePassword,
//     type UpdateProfilePhoto,
// } from "@application/shared/entities";
import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

// /**
//  * Update profile
//  */
// export type Profile_UpdateProfile_Req = ApiRequestBase<UpdateProfile>;

// export type Profile_UpdateProfile_Res = ApiResponseBase<{
//     type: "success";
// }>;

// /**
//  * Update profile photo
//  */
// export type Profile_UpdateProfilePhoto_Req = ApiRequestBase<UpdateProfilePhoto>;

// export type Profile_UpdateProfilePhoto_Res = ApiResponseBase<{
//     url: string | null;
// }>;

// /**
//  * Update profile email
//  */
// export type Profile_UpdateProfileEmail_Req = ApiRequestBase<UpdateProfileEmail>;

// export type Profile_UpdateProfileEmail_Res = ApiResponseBase<{
//     type: "success";
// }>;

// /**
//  * Update profile password
//  */
// export type Profile_UpdateProfilePassword_Req = ApiRequestBase<UpdateProfilePassword>;

// export type Profile_UpdateProfilePassword_Res = ApiResponseBase<{
//     type: "success";
// }>;

// /**
//  * Update profile locale
//  */
// export type Profile_UpdateProfileLocale_Req = ApiRequestBase<UpdateProfileLocale>;

// export type Profile_UpdateProfileLocale_Res = ApiResponseBase<{
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
