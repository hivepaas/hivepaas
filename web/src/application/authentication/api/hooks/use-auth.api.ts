import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { AuthenticationApiContext } from "@application/authentication/api/api-context";
import {
    type Auth_ForgotPassword_Req,
    type Auth_ResetPassword_Req,
    type Auth_Send2FAToken_Req,
    type Auth_SignIn2FA_Req,
    type Auth_SignIn_Req,
    type Auth_SignUp_Req,
    type Auth_ValidateInviteToken_Req,
    type Auth_ValidateResetToken_Req,
} from "@application/authentication/api/services";

import { session, useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useAuthenticationApi() {
        const { api } = use(AuthenticationApiContext);

        const { notifyError } = useApiErrorNotifications();

        const mutations = useMemo(
            () => ({
                /**
                 * Sign up
                 */
                signUp: async (data: Auth_SignUp_Req["data"]) => {
                    const result = await api.auth.signUp({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Registration failed",
                                error,
                            });

                            throw error;
                        },
                    });
                },

                /**
                 * Sign in with email and password
                 */
                signIn: async (data: Auth_SignIn_Req["data"]) => {
                    const result = await api.auth.signIn({
                        data,
                    });

                    return match(result, {
                        Ok: res => {
                            if (res.data.type === "success") {
                                session.setToken(res.data.token);
                            }

                            return res;
                        },
                        Err: error => {
                            notifyError({
                                message: "Authentication failed",
                                error,
                            });

                            throw error;
                        },
                    });
                },

                /**
                 * Send 2-factor authentication token
                 */
                send2FAToken: async (data: Auth_Send2FAToken_Req["data"]) => {
                    const result = await api.auth.send2FAToken({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "2-factor authentication service error",
                                error,
                            });

                            throw error;
                        },
                    });
                },

                /**
                 * Sign in with 2-factor authentication
                 */
                signIn2FA: async (data: Auth_SignIn2FA_Req["data"]) => {
                    const result = await api.auth.signIn2FA({
                        data,
                    });

                    return match(result, {
                        Ok: res => {
                            session.setToken(res.data.token);

                            return res;
                        },
                        Err: error => {
                            notifyError({
                                message: "Code verification failed",
                                error,
                            });

                            throw error;
                        },
                    });
                },

                /**
                 * Forgot password
                 */
                forgotPassword: async (data: Auth_ForgotPassword_Req["data"]) => {
                    const result = await api.auth.forgotPassword({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Reset password failed",
                                error,
                            });

                            throw error;
                        },
                    });
                },

                /**
                 * Validate reset token
                 */
                validateResetToken: async (data: Auth_ValidateResetToken_Req["data"]) => {
                    const result = await api.auth.validateResetToken({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Invalid token",
                                error,
                            });

                            throw error;
                        },
                    });
                },

                /**
                 * Reset password
                 */
                resetPassword: async (data: Auth_ResetPassword_Req["data"]) => {
                    const result = await api.auth.resetPassword({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Reset password failed",
                                error,
                            });

                            throw error;
                        },
                    });
                },

                /**
                 * Validate invite token
                 */
                validateInviteToken: async (data: Auth_ValidateInviteToken_Req["data"]) => {
                    const result = await api.auth.validateInviteToken({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            // notifyError({
                            //     message: "Invalid token",
                            //     error,
                            // });

                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const queries = useMemo(
            () => ({
                getLoginOptions: async (signal: AbortSignal) => {
                    const result = await api.auth.getLoginOptions(signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
            }),
            [api],
        );

        return {
            mutations,
            queries,
        };
    };
}

export const useAuthApi = createHook();
