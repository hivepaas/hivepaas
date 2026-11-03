import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectDomainSettings_DeleteOne_Req,
    ProjectDomainSettings_FindOne_Req,
    ProjectDomainSettings_UpdateOne_Req,
    ProjectDomainSettings_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectDomainSettingsApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (data: ProjectDomainSettings_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.domainSettings.$.findOne(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project domain settings",
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
                updateOne: async (data: ProjectDomainSettings_UpdateOne_Req["data"]) => {
                    const result = await api.projects.domainSettings.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project domain settings",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: ProjectDomainSettings_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.domainSettings.$.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project domain settings status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectDomainSettings_DeleteOne_Req["data"]) => {
                    const result = await api.projects.domainSettings.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project domain settings",
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

export const useProjectDomainSettingsApi = createHook();
