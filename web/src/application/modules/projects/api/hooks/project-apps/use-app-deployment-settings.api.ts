import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { useApiErrorNotifications } from "@infrastructure/api";

import { type AppDeploymentSettings } from "../../../domain/apps/deployment-settings";
import { ProjectsApiContext } from "../../api-context/projects.api.context";

function createHook() {
    return function useAppDeploymentSettingsApi() {
        const context = use(ProjectsApiContext);
        const { api } = context;
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (request: { projectID: string; appID: string }, signal?: AbortSignal) => {
                    const result = await api.projects.apps.deploymentSettings.$.findOne({ data: request }, signal);
                    return match(result, {
                        Ok: _ => _,
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
                updateOne: async (request: {
                    projectID: string;
                    appID: string;
                    updateVer: number;
                    payload: AppDeploymentSettings;
                }) => {
                    const result = await api.projects.apps.deploymentSettings.$.updateOne({ data: request });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update deployment settings", error });
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

export const useAppDeploymentSettingsApi = createHook();
