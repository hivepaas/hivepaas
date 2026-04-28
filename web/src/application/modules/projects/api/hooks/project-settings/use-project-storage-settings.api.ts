import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { ProjectsApiContext } from "../../api-context/projects.api.context";

function createHook() {
    return function useProjectStorageSettingsApi() {
        const context = use(ProjectsApiContext);
        const { api } = context;

        const queries = useMemo(
            () => ({
                findOne: async (request: { projectID: string }, signal?: AbortSignal) => {
                    const result = await api.projects.storageSettings.$.findOne({ data: request }, signal);
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

        return { queries };
    };
}

export const useProjectStorageSettingsApi = createHook();
