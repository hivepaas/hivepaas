import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { ApplicationApiContext } from "@application/shared/api/api-context";
import {
    type Profile_UpdateProfileEmail_Req,
    type Profile_UpdateProfileLocale_Req,
    type Profile_UpdateProfilePassword_Req,
    type Profile_UpdateProfilePhoto_Req,
    type Profile_UpdateProfile_Req,
} from "@application/shared/api/services";
import { useWorkspaceId } from "@application/shared/hooks/workspaces";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProfileApi() {
        const { workspaceId } = useWorkspaceId();

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

                /**
                 * Update profile photo
                 */
                updatePhoto: async (data: Profile_UpdateProfilePhoto_Req["data"]) => {
                    if (!workspaceId) {
                        const error = new Error("workspaceId must be defined");

                        notifyError({
                            message: "Failed to update profile photo",
                            error,
                        });

                        throw error;
                    }

                    const result = await api.profile.updatePhoto({
                        data,
                        meta: { workspaceId },
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update profile photo",
                                error,
                            });

                            throw error;
                        },
                    });
                },

                /**
                 * Update profile email
                 */
                updateEmail: async (data: Profile_UpdateProfileEmail_Req["data"]) => {
                    const result = await api.profile.updateEmail({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update profile email",
                                error,
                            });

                            throw error;
                        },
                    });
                },

                /**
                 * Update profile password
                 */
                updatePassword: async (data: Profile_UpdateProfilePassword_Req["data"]) => {
                    const result = await api.profile.updatePassword({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update profile password",
                                error,
                            });

                            throw error;
                        },
                    });
                },

                /**
                 * Update profile locale
                 */
                updateLocale: async (data: Profile_UpdateProfileLocale_Req["data"]) => {
                    const result = await api.profile.updateLocale({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update profile locale",
                                error,
                            });

                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError, workspaceId],
        );

        return {
            mutations,
        };
    };
}

export const useProfileApi = createHook();
