import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type {
    ImService_CreateOne_Req,
    ImService_DeleteOne_Req,
    ImService_FindManyPaginated_Req,
    ImService_FindOneById_Req,
    ImService_TestSendMsg_Req,
    ImService_UpdateOne_Req,
    ImService_UpdateStatus_Req,
} from "~/settings/api/services/im-service-services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useImServiceApi() {
        const { api } = use(SettingsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: ImService_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.imService.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get IM platforms",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: ImService_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.imService.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get IM platform",
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
                createOne: async (data: ImService_CreateOne_Req["data"]) => {
                    const result = await api.settings.imService.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create IM platform",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: ImService_UpdateOne_Req["data"]) => {
                    const result = await api.settings.imService.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update IM platform",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: ImService_UpdateStatus_Req["data"]) => {
                    const result = await api.settings.imService.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update IM platform status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ImService_DeleteOne_Req["data"]) => {
                    const result = await api.settings.imService.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete IM platform",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                testSendMsg: async (data: ImService_TestSendMsg_Req["data"]) => {
                    const result = await api.settings.imService.testSendMsg({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to test sending IM message",
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

export const useImServiceApi = createHook();
