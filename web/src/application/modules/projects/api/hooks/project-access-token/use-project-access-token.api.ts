import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectAccessToken_CreateOne_Req,
    ProjectAccessToken_DeleteOne_Req,
    ProjectAccessToken_FindManyPaginated_Req,
    ProjectAccessToken_FindOneById_Req,
    ProjectAccessToken_UpdateMeta_Req,
    ProjectAccessToken_UpdateOne_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectAccessTokenApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: ProjectAccessToken_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.accessToken.$.findManyPaginated({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get project access tokens", error });
                            throw error;
                        },
                    });
                },
                findOneById: async (data: ProjectAccessToken_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.accessToken.$.findOneById({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get project access token", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                createOne: async (data: ProjectAccessToken_CreateOne_Req["data"]) => {
                    const result = await api.projects.accessToken.$.createOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to create project access token", error });
                            throw error;
                        },
                    });
                },
                updateOne: async (data: ProjectAccessToken_UpdateOne_Req["data"]) => {
                    const result = await api.projects.accessToken.$.updateOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update project access token", error });
                            throw error;
                        },
                    });
                },
                updateMeta: async (data: ProjectAccessToken_UpdateMeta_Req["data"]) => {
                    const result = await api.projects.accessToken.$.updateMeta({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update project access token status", error });
                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectAccessToken_DeleteOne_Req["data"]) => {
                    const result = await api.projects.accessToken.$.deleteOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to delete project access token", error });
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

export const useProjectAccessTokenApi = createHook();
