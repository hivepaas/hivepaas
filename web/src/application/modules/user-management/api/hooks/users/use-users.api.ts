import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { UsersApiContext } from "~/user-management/api/api-context";
import { type Users_DeleteOne_Req, type Users_FindManyPaginated_Req } from "~/user-management/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useUsersApi() {
        const { api } = use(UsersApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                /**
                 * Find many users paginated
                 */
                findManyPaginated: async (data: Users_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.users.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get users",
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
                 * Delete a user
                 */
                deleteOne: async (data: Users_DeleteOne_Req["data"]) => {
                    const result = await api.users.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete user",
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

export const useUsersApi = createHook();
