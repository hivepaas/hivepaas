import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type {
    Email_CreateOne_Req,
    Email_DeleteOne_Req,
    Email_FindManyPaginated_Req,
    Email_FindOneById_Req,
    Email_TestSendMail_Req,
    Email_UpdateOne_Req,
    Email_UpdateStatus_Req,
} from "~/settings/api/services/email-services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useEmailApi() {
        const { api } = use(SettingsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: Email_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.email.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get Email accounts",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: Email_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.email.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get Email account",
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
                createOne: async (data: Email_CreateOne_Req["data"]) => {
                    const result = await api.settings.email.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create Email account",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: Email_UpdateOne_Req["data"]) => {
                    const result = await api.settings.email.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update Email account",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: Email_UpdateStatus_Req["data"]) => {
                    const result = await api.settings.email.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update Email account status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: Email_DeleteOne_Req["data"]) => {
                    const result = await api.settings.email.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete Email account",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                testSendMail: async (data: Email_TestSendMail_Req["data"]) => {
                    const result = await api.settings.email.testSendMail({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to test sending email",
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

export const useEmailApi = createHook();
