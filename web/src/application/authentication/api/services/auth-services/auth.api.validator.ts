import { type AxiosResponse } from "axios";
import { z } from "zod";

import { ESecuritySettings, EUserRole } from "@application/shared/enums";

import {
    type Auth_ForgotPassword_Res,
    type Auth_GetLoginOptions_Res,
    type Auth_ResetPassword_Res,
    type Auth_Send2FAToken_Res,
    type Auth_SignIn2FA_Res,
    type Auth_SignIn_Res,
    type Auth_SignUp_Res,
    type Auth_ValidateInviteToken_Res,
    type Auth_ValidateResetToken_Res,
} from "@application/authentication/api/services";

import { parseApiResponse } from "@infrastructure/api";

/**
 * Sign in API response schema
 */
const SignInSchema = z.object({
    data: z.union([
        z.object({
            session: z.object({
                accessToken: z.string(),
            }),
            nextStep: z.string().optional(),
        }),

        z.object({
            mfaType: z.string(),
            mfaToken: z.string(),
        }),
    ]),
});

/**
 * Send 2FA token API response schema
 */
const Send2FATokenSchema = z.object({
    data: z.object({
        requestBlockingDuration: z.number(),
    }),
});

/**
 * Sign in 2FA API response schema
 */
const SignIn2FASchema = z.object({
    data: z.object({
        session: z.object({
            accessToken: z.string(),
        }),
    }),
});

/**
 * Forgot password API response schema
 */
const ForgotPasswordSchema = z.object({
    data: z.object({
        linkExpirationMins: z.number(),
    }),
});

/**
 * Get login options API response schema
 */
const GetLoginOptionsSchema = z.object({
    data: z.array(
        z.object({
            authURL: z.string(),
            name: z.string(),
            type: z.string(),
        }),
    ),
});

/**
 * Validate invite token API response schema
 */
const ValidateInviteTokenSchema = z.object({
    data: z.discriminatedUnion("securityOption", [
        z.object({
            email: z.string().trim().email(),
            role: z.nativeEnum(EUserRole),
            accessExpireAt: z.coerce.date(),
            securityOption: z.literal(ESecuritySettings.PasswordOnly),
        }),
        z.object({
            email: z.string().trim().email(),
            role: z.nativeEnum(EUserRole),
            accessExpireAt: z.coerce.date(),
            mfaTotpSecret: z.string(),
            qrCode: z.object({
                dataBase64: z.string(),
            }),
            securityOption: z.literal(ESecuritySettings.Password2FA),
        }),
        z.object({
            email: z.string().trim().email(),
            role: z.nativeEnum(EUserRole),
            accessExpireAt: z.coerce.date(),
            securityOption: z.literal(ESecuritySettings.EnforceSSO),
        }),
    ]),
});

export class AuthApiValidator {
    /**
     * Validate and transform the sign-up API response
     */
    signUp = (_: AxiosResponse): Auth_SignUp_Res => {
        return {
            data: {
                type: "success",
            },
        };
    };

    /**
     * Validate and transform the sign in API response
     */
    signIn = (response: AxiosResponse): Auth_SignIn_Res => {
        const { data } = parseApiResponse({
            response,
            schema: SignInSchema,
        });

        if ("session" in data) {
            if ("nextStep" in data && data.nextStep === "NextMfaSetup") {
                return {
                    data: {
                        type: "mfa-setup-required",
                        token: data.session.accessToken,
                    },
                };
            }

            return {
                data: {
                    type: "success",
                    token: data.session.accessToken,
                },
            };
        }

        return {
            data: {
                type: "mfa-required",
                mfaToken: data.mfaToken,
            },
        };
    };

    /**
     * Validate and transform the send 2FA token API response
     */
    send2FAToken = (response: AxiosResponse): Auth_Send2FAToken_Res => {
        const { data } = parseApiResponse({
            response,
            schema: Send2FATokenSchema,
        });

        return {
            data,
        };
    };

    /**
     * Validate and transform the sign in 2FA API response
     */
    signIn2FA = (response: AxiosResponse): Auth_SignIn2FA_Res => {
        const { data } = parseApiResponse({
            response,
            schema: SignIn2FASchema,
        });

        return {
            data: {
                token: data.session.accessToken,
            },
        };
    };

    /**
     * Validate and transform the forgot password API response
     */
    forgotPassword = (response: AxiosResponse): Auth_ForgotPassword_Res => {
        const { data } = parseApiResponse({
            response,
            schema: ForgotPasswordSchema,
        });

        return {
            data,
        };
    };

    /**
     * Validate and transform the reset token validation API response
     */
    validateResetToken = (_: AxiosResponse): Auth_ValidateResetToken_Res => {
        return {
            data: {
                type: "success",
            },
        };
    };

    /**
     * Validate and transform the reset password API response
     */
    resetPassword = (_: AxiosResponse): Auth_ResetPassword_Res => {
        return {
            data: {
                type: "success",
            },
        };
    };

    /**
     * Validate and transform the get login options API response
     */
    getLoginOptions = (response: AxiosResponse): Auth_GetLoginOptions_Res => {
        const { data } = parseApiResponse({
            response,
            schema: GetLoginOptionsSchema,
        });

        return {
            data,
        };
    };

    /**
     * Validate and transform the validate invite token API response
     */
    validateInviteToken = (response: AxiosResponse): Auth_ValidateInviteToken_Res => {
        const { data } = parseApiResponse({
            response,
            schema: ValidateInviteTokenSchema,
        });

        if (
            data.securityOption === ESecuritySettings.PasswordOnly ||
            data.securityOption === ESecuritySettings.EnforceSSO
        ) {
            return {
                data: {
                    candidate: {
                        email: data.email,
                        role: data.role,
                        securityOption: data.securityOption,
                        accessExpiration: data.accessExpireAt,
                    },
                },
            };
        }
        return {
            data: {
                candidate: {
                    email: data.email,
                    role: data.role,
                    securityOption: data.securityOption,
                    accessExpiration: data.accessExpireAt,
                    mfaTotpSecret: data.mfaTotpSecret,
                    qrCode: data.qrCode.dataBase64,
                },
            },
        };
    };
}
