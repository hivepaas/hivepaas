import {
    type Candidate,
    type ForgotPassword,
    type ResetPassword,
    type SignIn,
    type SignIn2FA,
    type SignUp,
    type ValidateAccessCode,
    type ValidateInviteToken,
    type ValidateResetToken,
} from "@application/authentication/domain";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

import { type PreserveUnionOmit } from "@infrastructure/utility-types";

/**
 * Sign up
 */
export type Auth_SignUp_Req = ApiRequestBase<SignUp>;

export type Auth_SignUp_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Validate invite token
 */
export type Auth_ValidateInviteToken_Req = ApiRequestBase<ValidateInviteToken>;

export type Auth_ValidateInviteToken_Res = ApiResponseBase<{
    workspaceId: string;
    candidate: Candidate;
}>;

/**
 * Validate access code
 */
export type Auth_ValidateAccessCode_Req = ApiRequestBase<ValidateAccessCode>;

export type Auth_ValidateAccessCode_Res = ApiResponseBase<{
    workspaceId: string;
    candidate: PreserveUnionOmit<Candidate, "email">;
}>;

/**
 * Sign in with email and password
 */
export type Auth_SignIn_Req = ApiRequestBase<SignIn>;

export type Auth_SignIn_Res = ApiResponseBase<
    | {
          type: "success";
          token: string;
      }
    | {
          type: "mfa-required";
          mfaToken: string;
      }
>;

/**
 * Send 2-factor authentication token
 */
export type Auth_Send2FAToken_Req = ApiRequestBase<{
    mfaToken: string;
}>;

export type Auth_Send2FAToken_Res = ApiResponseBase<{
    requestBlockingDuration: number;
}>;

/**
 * Sign in with 2-factor authentication
 */
export type Auth_SignIn2FA_Req = ApiRequestBase<SignIn2FA>;

export type Auth_SignIn2FA_Res = ApiResponseBase<{
    token: string;
}>;

/**
 * Forgot password
 */
export type Auth_ForgotPassword_Req = ApiRequestBase<ForgotPassword>;

export type Auth_ForgotPassword_Res = ApiResponseBase<{
    linkExpirationMins: number;
}>;

/**
 * Validate reset password token
 */
export type Auth_ValidateResetToken_Req = ApiRequestBase<ValidateResetToken>;

export type Auth_ValidateResetToken_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Reset password
 */
export type Auth_ResetPassword_Req = ApiRequestBase<ResetPassword>;

export type Auth_ResetPassword_Res = ApiResponseBase<{
    type: "success";
}>;

export type Auth_GetLoginOptions_Res = ApiResponseBase<{
    options: {
        allowGithubLogin: boolean;
        allowGitlabLogin: boolean;
    };
}>;
