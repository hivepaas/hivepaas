import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type {
    AcmeDnsProvider_CreateOne_Req,
    AcmeDnsProvider_DeleteOne_Req,
    AcmeDnsProvider_FindManyPaginated_Req,
    AcmeDnsProvider_FindOneById_Req,
    AcmeDnsProvider_TestAccess_Req,
    AcmeDnsProvider_UpdateOne_Req,
    AcmeDnsProvider_UpdateStatus_Req,
} from "~/settings/api/services/acme-dns-provider-services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useAcmeDnsProviderApi() {
        const { api } = use(SettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: AcmeDnsProvider_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.settings.acmeDnsProvider.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get ACME DNS provider settings",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: AcmeDnsProvider_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.acmeDnsProvider.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get ACME DNS provider setting",
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
                createOne: async (data: AcmeDnsProvider_CreateOne_Req["data"]) => {
                    const result = await api.settings.acmeDnsProvider.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create ACME DNS provider setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: AcmeDnsProvider_UpdateOne_Req["data"]) => {
                    const result = await api.settings.acmeDnsProvider.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update ACME DNS provider setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: AcmeDnsProvider_UpdateStatus_Req["data"]) => {
                    const result = await api.settings.acmeDnsProvider.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update ACME DNS provider status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: AcmeDnsProvider_DeleteOne_Req["data"]) => {
                    const result = await api.settings.acmeDnsProvider.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete ACME DNS provider setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                testAccess: async (data: AcmeDnsProvider_TestAccess_Req["data"]) => {
                    const result = await api.settings.acmeDnsProvider.testAccess({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to test ACME DNS provider access",
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

export const useAcmeDnsProviderApi = createHook();
