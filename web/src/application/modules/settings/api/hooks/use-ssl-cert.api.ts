import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type {
    SslCert_CreateOne_Req,
    SslCert_DeleteOne_Req,
    SslCert_FindManyPaginated_Req,
    SslCert_FindOneById_Req,
    SslCert_UpdateOne_Req,
    SslCert_UpdateStatus_Req,
} from "~/settings/api/services/ssl-cert-services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useSslCertApi() {
        const { api } = use(SettingsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: SslCert_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.sslCert.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get SSL certificates",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: SslCert_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.sslCert.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get SSL certificate",
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
                createOne: async (data: SslCert_CreateOne_Req["data"]) => {
                    const result = await api.settings.sslCert.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create SSL certificate",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: SslCert_UpdateOne_Req["data"]) => {
                    const result = await api.settings.sslCert.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update SSL certificate",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: SslCert_UpdateStatus_Req["data"]) => {
                    const result = await api.settings.sslCert.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update SSL certificate status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: SslCert_DeleteOne_Req["data"]) => {
                    const result = await api.settings.sslCert.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete SSL certificate",
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

export const useSslCertApi = createHook();
