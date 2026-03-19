import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import {
    type Notifications_CreateOne_Req,
    type Notifications_DeleteOne_Req,
    type Notifications_FindManyPaginated_Req,
    type Notifications_FindOneById_Req,
    type Notifications_UpdateOne_Req,
} from "~/settings/api/services/notifications-services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useNotificationsApi() {
        const { api } = use(SettingsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                /**
                 * Find many notifications paginated
                 */
                findManyPaginated: async (data: Notifications_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.notifications.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get notifications",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Find one notification by id
                 */
                findOneById: async (data: Notifications_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.notifications.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get notification",
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
                 * Create one notification
                 */
                createOne: async (data: Notifications_CreateOne_Req["data"]) => {
                    const result = await api.settings.notifications.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create notification",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Update one notification
                 */
                updateOne: async (data: Notifications_UpdateOne_Req["data"]) => {
                    const result = await api.settings.notifications.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update notification",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Delete one notification
                 */
                deleteOne: async (data: Notifications_DeleteOne_Req["data"]) => {
                    const result = await api.settings.notifications.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete notification",
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

export const useNotificationsApi = createHook();
