import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectImageBuildSettings_ClearRepoCache_Req,
    ProjectImageBuildSettings_FindOne_Req,
    ProjectImageBuildSettings_FindRepoCache_Req,
    ProjectImageBuildSettings_UpdateOne_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectImageBuildSettingsApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (data: ProjectImageBuildSettings_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.imageBuildSettings.$.findOne({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get image build settings", error });
                            throw error;
                        },
                    });
                },
                findRepoCache: async (
                    data: ProjectImageBuildSettings_FindRepoCache_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.imageBuildSettings.$.findRepoCache({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get repo cache info", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                updateOne: async (data: ProjectImageBuildSettings_UpdateOne_Req["data"]) => {
                    const result = await api.projects.imageBuildSettings.$.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update image build settings", error });
                            throw error;
                        },
                    });
                },
                clearRepoCache: async (data: ProjectImageBuildSettings_ClearRepoCache_Req["data"]) => {
                    const result = await api.projects.imageBuildSettings.$.clearRepoCache({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to clear repo cache", error });
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

export const useProjectImageBuildSettingsApi = createHook();
