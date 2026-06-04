import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectNetworks_CreateOne_Req,
    ProjectNetworks_DeleteOne_Req,
    ProjectNetworks_FindManyPaginated_Req,
    ProjectNetworks_FindOneById_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectNetworksApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: ProjectNetworks_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.networks.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project networks",
                                error,
                            });
                            throw error;
                        },
                    });
                },
                findOneById: async (data: ProjectNetworks_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.networks.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project network",
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
                createOne: async (data: ProjectNetworks_CreateOne_Req["data"]) => {
                    const result = await api.projects.networks.$.createOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project network",
                                error,
                            });
                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectNetworks_DeleteOne_Req["data"]) => {
                    const result = await api.projects.networks.$.deleteOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project network",
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

export const useProjectNetworksApi = createHook();
