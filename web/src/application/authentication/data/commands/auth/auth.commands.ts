import { type UseMutationOptions, useMutation } from "@tanstack/react-query";

import { useSessionApi } from "@application/shared/api";
import { type Profile } from "@application/shared/entities";

import { useAuthApi } from "@application/authentication/api";
import {
    type Auth_ForgotPassword_Req,
    type Auth_ForgotPassword_Res,
    type Auth_ResetPassword_Req,
    type Auth_ResetPassword_Res,
    type Auth_Send2FAToken_Req,
    type Auth_Send2FAToken_Res,
    type Auth_SignIn2FA_Req,
    type Auth_SignIn2FA_Res,
    type Auth_SignIn_Req,
    type Auth_SignIn_Res,
    type Auth_SignUp_Req,
    type Auth_SignUp_Res,
    type Auth_ValidateInviteToken_Req,
    type Auth_ValidateInviteToken_Res,
    type Auth_ValidateResetToken_Req,
    type Auth_ValidateResetToken_Res,
} from "@application/authentication/api/services";
import { type SignIn, type SignIn2FA, type SignInSSO } from "@application/authentication/domain";

// import { useI18n } from "@i18n/hooks";

import { isToManyLoginAttemptsException, session } from "@infrastructure/api";

/**
 * Sign up command
 */
interface SignUpReq {
    data: Auth_SignUp_Req["data"];
}
type SignUpRes = Auth_SignUp_Res;

type SignUpOptions = Omit<UseMutationOptions<SignUpRes, Error, SignUpReq>, "mutationFn">;

function useSignUp({ ...options }: SignUpOptions = {}) {
    const {
        mutations: { signUp },
    } = useAuthApi();

    async function signUpFn(values: SignUpReq) {
        return await signUp(values.data);
    }

    return useMutation({
        mutationFn: signUpFn,

        ...options,
    });
}

/**
 * Sign in command
 */
type SignInReq = Auth_SignIn_Req["data"];
type SignInRes = Auth_SignIn_Res;

type SignInOptions = Omit<UseMutationOptions<SignInRes, Error, SignInReq>, "mutationFn" | "onSuccess"> & {
    onSuccess?: (profile: Profile) => void;
    on2FARequired?: (data: { email: string; mfaToken: string }) => void;
    onTooManyAttempts?: (error: Error) => void;
};

function useSignIn({ onSuccess, onError, on2FARequired, onTooManyAttempts, ...options }: SignInOptions = {}) {
    // const { changeLanguage } = useI18n();

    const {
        queries: { getProfile },
    } = useSessionApi();

    const {
        mutations: { signIn },
    } = useAuthApi();

    async function signInFn(values: SignIn) {
        const res = await signIn(values);

        switch (res.data.type) {
            case "success": {
                const { data: profile } = await getProfile();

                // await changeLanguage(profile.language);

                if (onSuccess) onSuccess(profile);

                break;
            }

            case "mfa-required": {
                const { mfaToken } = res.data;

                // await send2FAToken({
                //     mfaToken,
                // });

                if (on2FARequired) on2FARequired({ email: values.email, mfaToken });
            }
        }

        return res;
    }

    return useMutation({
        mutationFn: signInFn,
        onError: (error, ...rest) => {
            if (isToManyLoginAttemptsException(error)) {
                if (onTooManyAttempts) onTooManyAttempts(error);

                return;
            }

            if (onError) onError(error, ...rest);
        },

        ...options,
    });
}

/**
 * Send 2-factor authentication token command
 */
type Send2FATokenReq = Auth_Send2FAToken_Req["data"];
type Send2FATokenRes = Auth_Send2FAToken_Res;

type Send2FATokenOptions = Omit<UseMutationOptions<Send2FATokenRes, Error, Send2FATokenReq>, "mutationFn">;

function useSend2FAToken(options: Send2FATokenOptions = {}) {
    const {
        mutations: { send2FAToken },
    } = useAuthApi();

    return useMutation({
        mutationFn: send2FAToken,

        ...options,
    });
}

/**
 * Sign in with 2-factor authentication command
 */
type SignIn2FAReq = Auth_SignIn2FA_Req["data"];
type SignIn2FARes = Auth_SignIn2FA_Res;

type SignIn2FAOptions = Omit<UseMutationOptions<SignIn2FARes, Error, SignIn2FAReq>, "mutationFn" | "onSuccess"> & {
    onSuccess?: (profile: Profile) => void;
    onTooManyAttempts?: (error: Error) => void;
};

function useSignIn2FA({ onSuccess, onError, onTooManyAttempts, ...options }: SignIn2FAOptions = {}) {
    // const { changeLanguage } = useI18n();

    const {
        queries: { getProfile },
    } = useSessionApi();

    const {
        mutations: { signIn2FA },
    } = useAuthApi();

    async function signIn2FAFn(values: SignIn2FA) {
        const res = await signIn2FA({
            code: values.code,
            mfaToken: values.mfaToken,
        });

        const { data: profile } = await getProfile();

        // await changeLanguage(profile.language);

        if (onSuccess) onSuccess(profile);

        return res;
    }

    return useMutation({
        mutationFn: signIn2FAFn,
        onError: (error, ...rest) => {
            if (isToManyLoginAttemptsException(error)) {
                if (onTooManyAttempts) onTooManyAttempts(error);

                return;
            }

            if (onError) onError(error, ...rest);
        },

        ...options,
    });
}

/**
 * Sign in after SSO success
 */
type SignInSSOOptions = Omit<UseMutationOptions<void, Error, SignInSSO>, "mutationFn" | "onSuccess"> & {
    onSuccess?: (profile: Profile) => void;
};

function useSignInSSO({ onSuccess, ...options }: SignInSSOOptions = {}) {
    // const { changeLanguage } = useI18n();

    const {
        queries: { getProfile },
    } = useSessionApi();

    async function signInSSOFn({ token }: SignInSSO) {
        session.setToken(token);

        const { data: profile } = await getProfile();

        // await changeLanguage(profile.language);

        if (onSuccess) onSuccess(profile);
    }

    return useMutation({
        mutationFn: signInSSOFn,

        ...options,
    });
}

/**
 * Forgot password command
 */
type ForgotPasswordReq = Auth_ForgotPassword_Req["data"];
type ForgotPasswordRes = Auth_ForgotPassword_Res;

type ForgotPasswordOptions = Omit<UseMutationOptions<ForgotPasswordRes, Error, ForgotPasswordReq>, "mutationFn">;

function useForgotPassword(options: ForgotPasswordOptions = {}) {
    const {
        mutations: { forgotPassword },
    } = useAuthApi();

    return useMutation({
        mutationFn: forgotPassword,

        ...options,
    });
}

/**
 * Validate reset token command
 */
type ValidateResetTokenReq = Auth_ValidateResetToken_Req["data"];
type ValidateResetTokenRes = Auth_ValidateResetToken_Res;

type ValidateResetTokenOptions = Omit<
    UseMutationOptions<ValidateResetTokenRes, Error, ValidateResetTokenReq>,
    "mutationFn"
>;

function useValidateResetToken(options: ValidateResetTokenOptions = {}) {
    const {
        mutations: { validateResetToken },
    } = useAuthApi();

    return useMutation({
        mutationFn: validateResetToken,

        ...options,
    });
}

/**
 * Reset password command
 */
type ResetPasswordReq = Auth_ResetPassword_Req["data"];
type ResetPasswordRes = Auth_ResetPassword_Res;

type ResetPasswordOptions = Omit<UseMutationOptions<ResetPasswordRes, Error, ResetPasswordReq>, "mutationFn">;

function useResetPassword(options: ResetPasswordOptions = {}) {
    const {
        mutations: { resetPassword },
    } = useAuthApi();

    return useMutation({
        mutationFn: resetPassword,

        ...options,
    });
}

/**
 * Validate invite token command
 */
type ValidateInviteTokenReq = Auth_ValidateInviteToken_Req["data"];
type ValidateInviteTokenRes = Auth_ValidateInviteToken_Res;

type ValidateInviteTokenOptions = Omit<
    UseMutationOptions<ValidateInviteTokenRes, Error, ValidateInviteTokenReq>,
    "mutationFn"
>;

function useValidateInviteToken(options: ValidateInviteTokenOptions = {}) {
    const {
        mutations: { validateInviteToken },
    } = useAuthApi();

    return useMutation({
        mutationFn: validateInviteToken,

        ...options,
    });
}

export const AuthCommands = Object.freeze({
    useSignUp,
    useSignIn,
    useSend2FAToken,
    useSignIn2FA,
    useSignInSSO,
    useForgotPassword,
    useValidateResetToken,
    useResetPassword,
    useValidateInviteToken,
});
