import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    AppDataFiles_DeleteOne_Req,
    AppDataFiles_FindManyPaginated_Req,
    AppDataFiles_GetDownloadUrl_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useAppDataFilesApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: AppDataFiles_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.dataFiles.$.findManyPaginated({ data }, signal);

                    return match(result, {
                        Ok: response => response,
                        Err: error => {
                            throw error;
                        },
                    });
                },
                getDownloadUrl: async (data: AppDataFiles_GetDownloadUrl_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.dataFiles.$.getDownloadUrl({ data }, signal);

                    return match(result, {
                        Ok: response => response,
                        Err: error => {
                            throw error;
                        },
                    });
                },
            }),
            [api],
        );

        const mutations = useMemo(
            () => ({
                deleteOne: async (data: AppDataFiles_DeleteOne_Req["data"]) => {
                    const result = await api.projects.apps.dataFiles.$.deleteOne({ data });

                    return match(result, {
                        Ok: response => response,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete app data file",
                                error,
                            });

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

export const useAppDataFilesApi = createHook();
