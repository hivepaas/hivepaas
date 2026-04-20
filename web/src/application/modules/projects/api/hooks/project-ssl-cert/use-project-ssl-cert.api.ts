import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectSslCert_CreateOne_Req,
    ProjectSslCert_DeleteOne_Req,
    ProjectSslCert_FindManyPaginated_Req,
    ProjectSslCert_FindOneById_Req,
    ProjectSslCert_UpdateOne_Req,
    ProjectSslCert_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectSslCertApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: ProjectSslCert_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.sslCert.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project SSL certificates",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (
                    data: ProjectSslCert_FindOneById_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.sslCert.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project SSL certificate",
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
                createOne: async (data: ProjectSslCert_CreateOne_Req["data"]) => {
                    const result = await api.projects.sslCert.$.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project SSL certificate",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: ProjectSslCert_UpdateOne_Req["data"]) => {
                    const result = await api.projects.sslCert.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project SSL certificate",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: ProjectSslCert_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.sslCert.$.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project SSL certificate status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectSslCert_DeleteOne_Req["data"]) => {
                    const result = await api.projects.sslCert.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project SSL certificate",
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

export const useProjectSslCertApi = createHook();
