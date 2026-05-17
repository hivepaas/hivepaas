import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectCloudStorage_CreateOne_Req,
    ProjectCloudStorage_DeleteOne_Req,
    ProjectCloudStorage_FindManyPaginated_Req,
    ProjectCloudStorage_FindOneById_Req,
    ProjectCloudStorage_UpdateMeta_Req,
    ProjectCloudStorage_UpdateOne_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectCloudStorageApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: ProjectCloudStorage_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.cloudStorage.$.findManyPaginated({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get project cloud storages", error });
                            throw error;
                        },
                    });
                },
                findOneById: async (data: ProjectCloudStorage_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.cloudStorage.$.findOneById({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get project cloud storage", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                createOne: async (data: ProjectCloudStorage_CreateOne_Req["data"]) => {
                    const result = await api.projects.cloudStorage.$.createOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to create project cloud storage", error });
                            throw error;
                        },
                    });
                },
                updateOne: async (data: ProjectCloudStorage_UpdateOne_Req["data"]) => {
                    const result = await api.projects.cloudStorage.$.updateOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update project cloud storage", error });
                            throw error;
                        },
                    });
                },
                updateMeta: async (data: ProjectCloudStorage_UpdateMeta_Req["data"]) => {
                    const result = await api.projects.cloudStorage.$.updateMeta({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update project cloud storage status", error });
                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectCloudStorage_DeleteOne_Req["data"]) => {
                    const result = await api.projects.cloudStorage.$.deleteOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to delete project cloud storage", error });
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

export const useProjectCloudStorageApi = createHook();
