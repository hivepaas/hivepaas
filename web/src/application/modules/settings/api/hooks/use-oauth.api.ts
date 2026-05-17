import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type {
    OAuth_CreateOne_Req,
    OAuth_DeleteOne_Req,
    OAuth_FindManyPaginated_Req,
    OAuth_FindOneById_Req,
    OAuth_UpdateMeta_Req,
    OAuth_UpdateOne_Req,
} from "~/settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useOAuthApi() {
        const { api } = use(SettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: OAuth_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.oauth.findManyPaginated({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get OAuth settings", error });
                            throw error;
                        },
                    });
                },
                findOneById: async (data: OAuth_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.oauth.findOneById({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get OAuth setting", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                createOne: async (data: OAuth_CreateOne_Req["data"]) => {
                    const result = await api.settings.oauth.createOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to create OAuth setting", error });
                            throw error;
                        },
                    });
                },
                updateOne: async (data: OAuth_UpdateOne_Req["data"]) => {
                    const result = await api.settings.oauth.updateOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update OAuth setting", error });
                            throw error;
                        },
                    });
                },
                updateMeta: async (data: OAuth_UpdateMeta_Req["data"]) => {
                    const result = await api.settings.oauth.updateMeta({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update OAuth status", error });
                            throw error;
                        },
                    });
                },
                deleteOne: async (data: OAuth_DeleteOne_Req["data"]) => {
                    const result = await api.settings.oauth.deleteOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to delete OAuth setting", error });
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

export const useOAuthApi = createHook();
