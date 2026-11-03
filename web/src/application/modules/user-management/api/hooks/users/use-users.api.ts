import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { UsersApiContext } from "~/user-management/api/api-context";
import type {
    Users_DeleteOne_Req,
    Users_FindManyPaginated_Req,
    Users_FindOneById_Req,
    Users_InviteOne_Req,
    Users_ResetPassword_Req,
    Users_UpdateOne_Req,
} from "~/user-management/api/services";

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
                /**
                 * Find one user by id
                 */
                findOneById: async (data: Users_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.users.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get user",
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
                /**
                 * Update a user
                 */
                updateOne: async (data: Users_UpdateOne_Req["data"]) => {
                    const result = await api.users.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update user",
                                error,
                            });

                            throw error;
                        },
                    });
                },

                /**
                 * Invite a user
                 */
                inviteOne: async (data: Users_InviteOne_Req["data"]) => {
                    const result = await api.users.$.inviteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to invite user",
                                error,
                            });

                            throw error;
                        },
                    });
                },

                /**
                 * Reset user password
                 */
                resetPassword: async (data: Users_ResetPassword_Req["data"]) => {
                    const result = await api.users.$.resetPassword({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to reset user password",
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
