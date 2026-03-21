import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectGitCredentials_FindManyPaginated_Req,
    ProjectGitCredentials_FindManyRepos_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectGitCredentialsApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: ProjectGitCredentials_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.gitCredentials.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project git credentials",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findManyRepos: async (data: ProjectGitCredentials_FindManyRepos_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.gitCredentials.$.findManyRepos(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to list git repositories",
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
        };
    };
}

export const useProjectGitCredentialsApi = createHook();
