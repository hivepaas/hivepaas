import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type {
    AccessToken_CreateOne_Req,
    AccessToken_DeleteOne_Req,
    AccessToken_FindManyPaginated_Req,
    AccessToken_FindOneById_Req,
    AccessToken_TestConn_Req,
    AccessToken_UpdateMeta_Req,
    AccessToken_UpdateOne_Req,
} from "~/settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useAccessTokenApi() {
        const { api } = use(SettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: AccessToken_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.accessToken.findManyPaginated({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get access tokens", error });
                            throw error;
                        },
                    });
                },
                findOneById: async (data: AccessToken_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.accessToken.findOneById({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get access token", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                createOne: async (data: AccessToken_CreateOne_Req["data"]) => {
                    const result = await api.settings.accessToken.createOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to create access token", error });
                            throw error;
                        },
                    });
                },
                updateOne: async (data: AccessToken_UpdateOne_Req["data"]) => {
                    const result = await api.settings.accessToken.updateOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update access token", error });
                            throw error;
                        },
                    });
                },
                updateMeta: async (data: AccessToken_UpdateMeta_Req["data"]) => {
                    const result = await api.settings.accessToken.updateMeta({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update access token status", error });
                            throw error;
                        },
                    });
                },
                deleteOne: async (data: AccessToken_DeleteOne_Req["data"]) => {
                    const result = await api.settings.accessToken.deleteOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to delete access token", error });
                            throw error;
                        },
                    });
                },
                testConn: async (data: AccessToken_TestConn_Req["data"]) => {
                    const result = await api.settings.accessToken.testConn({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to test access token", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return { queries, mutations };
    };
}

export const useAccessTokenApi = createHook();
