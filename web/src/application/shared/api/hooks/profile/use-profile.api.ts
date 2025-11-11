import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { ApplicationApiContext } from "@application/shared/api/api-context";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProfileApi() {
        const { api } = use(ApplicationApiContext);

        const { notifyError } = useApiErrorNotifications();

        const mutations = useMemo(
            () => ({
                // /**
                //  * Update profile
                //  */
                // update: async (data: Profile_UpdateProfile_Req["data"]) => {
                //     const result = await api.profile.update({
                //         data,
                //     });
                //     return match(result, {
                //         Ok: _ => _,
                //         Err: error => {
                //             notifyError({
                //                 message: "Failed to update profile",
                //                 error,
                //             });
                //             throw error;
                //         },
                //     });
                // },
                // /**
                //  * Update profile photo
                //  */
                // updatePhoto: async (data: Profile_UpdateProfilePhoto_Req["data"]) => {
                //     const result = await api.profile.updatePhoto({
                //         data,
                //     });
                //     return match(result, {
                //         Ok: _ => _,
                //         Err: error => {
                //             notifyError({
                //                 message: "Failed to update profile photo",
                //                 error,
                //             });
                //             throw error;
                //         },
                //     });
                // },
                // /**
                //  * Update profile email
                //  */
                // updateEmail: async (data: Profile_UpdateProfileEmail_Req["data"]) => {
                //     const result = await api.profile.updateEmail({
                //         data,
                //     });
                //     return match(result, {
                //         Ok: _ => _,
                //         Err: error => {
                //             notifyError({
                //                 message: "Failed to update profile email",
                //                 error,
                //             });
                //             throw error;
                //         },
                //     });
                // },
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
                // /**
                //  * Update profile locale
                //  */
                // updateLocale: async (data: Profile_UpdateProfileLocale_Req["data"]) => {
                //     const result = await api.profile.updateLocale({
                //         data,
                //     });
                //     return match(result, {
                //         Ok: _ => _,
                //         Err: error => {
                //             notifyError({
                //                 message: "Failed to update profile locale",
                //                 error,
                //             });
                //             throw error;
                //         },
                //     });
                // },

                /**
                 * Get profile 2FA setup
                 */
                getProfile2FASetup: async () => {
                    const result = await api.profile.getProfile2FASetup();
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
            }),
            [api, notifyError],
        );

        return {
            mutations,
        };
    };
}

export const useProfileApi = createHook();
