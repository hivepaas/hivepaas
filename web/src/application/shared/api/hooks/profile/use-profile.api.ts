import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { ApplicationApiContext } from "@application/shared/api/api-context";

import { useApiErrorNotifications } from "@infrastructure/api";

import type {
    Profile_Complete2FASetup_Req,
    Profile_GetProfile2FASetup_Req,
    Profile_UpdateProfile_Req,
} from "../../services";

function createHook() {
    return function useProfileApi() {
        const { api } = use(ApplicationApiContext);

        const { notifyError } = useApiErrorNotifications();

        const mutations = useMemo(
            () => ({
                /**
                 * Update profile
                 */
                update: async (data: Profile_UpdateProfile_Req["data"]) => {
                    const result = await api.profile.update({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update profile",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                // /**
                //  * Update profile password
                //  */
                // updatePassword: async (data: Profile_UpdateProfilePassword_Req["data"]) => {
                //     const result = await api.profile.updatePassword({
                //         data,
                //     });
                //     return match(result, {
                //         Ok: _ => _,
                //         Err: error => {
                //             notifyError({
                //                 message: "Failed to update profile password",
                //                 error,
                //             });
                //             throw error;
                //         },
                //     });
                // },

                /**
                 * Get profile 2FA setup
                 */
                getProfile2FASetup: async (request: Profile_GetProfile2FASetup_Req["data"]) => {
                    const result = await api.profile.getProfile2FASetup({
                        data: request,
                    });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get profile 2FA setup",
                                error,
                            });
                            throw error;
                        },
                    });
                },

                /**
                 * Verify profile 2FA setup
                 */
                complete2FASetup: async (request: Profile_Complete2FASetup_Req["data"]) => {
                    const result = await api.profile.complete2FASetup({
                        data: request,
                    });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to verify profile 2FA setup",
                                error,
                            });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return {
            mutations,
        };
    };
}

export const useProfileApi = createHook();
