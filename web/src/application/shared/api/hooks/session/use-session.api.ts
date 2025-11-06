import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { ApplicationApiContext } from "@application/shared/api/api-context";

import { isSessionInvalidException, session, useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useSessionApi() {
        const { api } = use(ApplicationApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                /**
                 * Get profile
                 */
                getProfile: async (signal?: AbortSignal) => {
                    const result = await api.session.getProfile(signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            if (isSessionInvalidException(error)) {
                                session.removeToken();
                            }

                            notifyError({
                                message: "Failed to get user profile",
                                error,
                            });

                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                /**
                 * Logout
                 */
                logout: async () => {
                    const result = await api.session.logout();

                    return match(result, {
                        Ok: res => {
                            session.removeToken();

                            return res;
                        },
                        Err: error => {
                            if (isSessionInvalidException(error)) {
                                session.removeToken();
                            }

                            notifyError({
                                message: "Logout failed",
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
            queries,
            mutations,
        };
    };
}

export const useSessionApi = createHook();
