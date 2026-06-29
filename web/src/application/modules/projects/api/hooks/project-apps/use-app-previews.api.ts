import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    AppPreviews_CreateOne_Req,
    AppPreviews_FindManyPaginated_Req,
    AppPreviews_PrepareCreate_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useAppPreviewsApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: AppPreviews_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.previews.$.findManyPaginated({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get app previews", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                prepareCreate: async (data: AppPreviews_PrepareCreate_Req["data"]) => {
                    const result = await api.projects.apps.previews.$.prepareCreate({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to prepare preview deployment", error });
                            throw error;
                        },
                    });
                },
                createOne: async (data: AppPreviews_CreateOne_Req["data"]) => {
                    const result = await api.projects.apps.previews.$.createOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to create preview deployment", error });
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

export const useAppPreviewsApi = createHook();
